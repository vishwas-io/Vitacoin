// stress_client.go — VitaCoin Proper Stress Test Client
// Uses vitacoind CLI with explicit --sequence to avoid sequence mismatch
// Each account has ONE serialized goroutine — no concurrent sequence races
// Build: go build -o stress_client stress_client.go
// Usage: ./stress_client -tps=10 -duration=30 -mode=ramp
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ---- Config ----

const (
	restURL  = "http://localhost:1317"
	rpcURL   = "tcp://localhost:26657"
	chainID  = "vitacoin-testnet-1"
	denom    = "uvita"
	sendAmt  = "1000"
	feeAmt   = "2000"
	gasLimit = 200000
	binPath  = "/home/gcp-instance-20260311/vitacoind"
	homeDir  = "/home/gcp-instance-20260311/.vitacoin-testnet"
)

type accountCfg struct {
	Name    string
	Address string
	AccNum  uint64
}

var accounts = []accountCfg{
	{Name: "stress0", Address: "cosmos1gs3rwt98h34zw0kdu4z7zd4s2u7gvur4yx0z83", AccNum: 14},
	{Name: "stress1", Address: "cosmos1rlg9qnj70f8rgfr0pwww9gk2ndamrp2axexw52", AccNum: 16},
}

// ---- Stats ----

var (
	txSent    int64
	txSuccess int64
	txFail    int64
	startTime time.Time
)

// ---- REST types ----

type AuthAccountResp struct {
	Account struct {
		AccountNumber string `json:"account_number"`
		Sequence      string `json:"sequence"`
	} `json:"account"`
}

type BroadcastResp struct {
	TxResponse struct {
		Code   int    `json:"code"`
		Txhash string `json:"txhash"`
		RawLog string `json:"raw_log"`
	} `json:"tx_response"`
}

// ---- Main ----

func main() {
	tps      := flag.Int("tps", 10, "target transactions per second")
	duration := flag.Int("duration", 30, "test duration in seconds")
	mode     := flag.String("mode", "ramp", "mode: ramp|fixed|burst")
	flag.Parse()

	fmt.Printf("=== VitaCoin Stress Test ===\n")
	fmt.Printf("Mode: %s | Target TPS: %d | Duration: %ds\n\n", *mode, *tps, *duration)

	block, err := getLatestBlock()
	if err != nil {
		fmt.Printf("ERROR: cannot reach chain: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Chain: %s | Block: %s\n", chainID, block)

	// Get initial sequences
	seqs := make([]uint64, len(accounts))
	for i, acc := range accounts {
		seq, err := getSequence(acc.Address)
		if err != nil {
			fmt.Printf("ERROR: cannot get sequence for %s: %v\n", acc.Name, err)
			os.Exit(1)
		}
		seqs[i] = seq
		accNum, _ := getAccNum(acc.Address)
		accounts[i].AccNum = accNum
		fmt.Printf("%s seq=%d accNum=%d\n", acc.Name, seq, accNum)
	}
	fmt.Println()

	startTime = time.Now()

	switch *mode {
	case "fixed":
		runFixed(*tps, *duration, seqs)
	case "burst":
		runBurst(*duration, seqs)
	default:
		runRamp(*duration, seqs)
	}

	elapsed := time.Since(startTime).Seconds()
	totalSent    := atomic.LoadInt64(&txSent)
	totalSuccess := atomic.LoadInt64(&txSuccess)
	totalFail    := atomic.LoadInt64(&txFail)
	actualTPS := float64(totalSuccess) / elapsed

	fmt.Printf("\n=== FINAL RESULTS ===\n")
	fmt.Printf("Duration:    %.1fs\n", elapsed)
	fmt.Printf("Sent:        %d\n", totalSent)
	fmt.Printf("Success:     %d (%.1f%%)\n", totalSuccess, 100*float64(totalSuccess)/max1(float64(totalSent),1))
	fmt.Printf("Failed:      %d\n", totalFail)
	fmt.Printf("Actual TPS:  %.2f\n", actualTPS)
	finalBlock, _ := getLatestBlock()
	fmt.Printf("Start Block: %s | Final Block: %s\n", block, finalBlock)
}

// runRamp: escalate through TPS tiers
func runRamp(totalDuration int, seqs []uint64) {
	tiers := []struct {
		tps  int
		secs int
	}{
		{10, 30},
		{25, 30},
		{50, 40},
		{100, 40},
	}

	for _, tier := range tiers {
		if time.Since(startTime).Seconds() >= float64(totalDuration) {
			break
		}
		runTier(tier.tps, tier.secs, seqs)
	}
}

func runFixed(tps, duration int, seqs []uint64) {
	runTier(tps, duration, seqs)
}

func runBurst(duration int, seqs []uint64) {
	// Burst: flood as fast as possible for duration
	runTier(9999, duration, seqs)
}

// runTier: core worker loop
// Each account has ONE goroutine that serializes its own txs in order.
// A dispatcher goroutine feeds jobs to account workers round-robin.
func runTier(tps, duration int, seqs []uint64) {
	n := len(accounts)
	tierSent    := int64(0)
	tierSuccess := int64(0)
	tierFail    := int64(0)
	start := time.Now()

	fmt.Printf("--- Tier: %d TPS for %ds ---\n", tps, duration)

	// Per-account job channels (buffered to smooth bursts)
	chans := make([]chan struct{}, n)
	for i := range chans {
		chans[i] = make(chan struct{}, 500)
	}

	// Per-account workers — each maintains its own sequence
	accSeqs := make([]uint64, n)
	copy(accSeqs, seqs)
	var accMus [2]sync.Mutex

	var wgWorkers sync.WaitGroup
	for idx := range accounts {
		wgWorkers.Add(1)
		go func(i int) {
			defer wgWorkers.Done()
			toIdx := (i + 1) % n
			for range chans[i] {
				accMus[i].Lock()
				seq := accSeqs[i]
				accSeqs[i]++
				accMus[i].Unlock()

				err := broadcastViaCLI(accounts[i], accounts[toIdx].Address, seq)
				atomic.AddInt64(&txSent, 1)
				atomic.AddInt64(&tierSent, 1)
				if err != nil {
					atomic.AddInt64(&txFail, 1)
					atomic.AddInt64(&tierFail, 1)
					// On sequence error: refresh from chain
					if isSeqErr(err) {
						realSeq, e2 := getSequence(accounts[i].Address)
						if e2 == nil {
							accMus[i].Lock()
							if realSeq > accSeqs[i] {
								accSeqs[i] = realSeq
							}
							accMus[i].Unlock()
						}
					}
				} else {
					atomic.AddInt64(&txSuccess, 1)
					atomic.AddInt64(&tierSuccess, 1)
				}
			}
		}(idx)
	}

	// Dispatcher: emit jobs at target TPS
	deadline := time.After(time.Duration(duration) * time.Second)
	i := 0

	if tps >= 9999 {
		// Burst: as fast as possible
	dispatchBurst:
		for {
			select {
			case <-deadline:
				break dispatchBurst
			default:
				chans[i%n] <- struct{}{}
				i++
			}
		}
	} else {
		interval := time.Second / time.Duration(tps)
		ticker := time.NewTicker(interval)
	dispatchFixed:
		for {
			select {
			case <-deadline:
				ticker.Stop()
				break dispatchFixed
			case <-ticker.C:
				chans[i%n] <- struct{}{}
				i++
			}
		}
	}

	// Close channels and wait for workers to drain
	for _, ch := range chans {
		close(ch)
	}
	wgWorkers.Wait()

	// Update seqs for next tier
	for idx := range accounts {
		seqs[idx] = accSeqs[idx]
	}

	elapsed := time.Since(start).Seconds()
	fmt.Printf("  Tier %d TPS: sent=%d success=%d fail=%d actual=%.2f TPS (%.0fs)\n",
		tps, tierSent, tierSuccess, tierFail, float64(tierSuccess)/elapsed, elapsed)
	b, _ := getLatestBlock()
	fmt.Printf("  Block: %s\n", b)
}

// broadcastViaCLI calls vitacoind with explicit --sequence
func broadcastViaCLI(from accountCfg, toAddr string, seq uint64) error {
	args := []string{
		"tx", "send",
		from.Address, toAddr,
		sendAmt + denom,
		"--from", from.Name,
		"--chain-id", chainID,
		"--home", homeDir,
		"--node", rpcURL,
		"--keyring-backend", "test",
		"--fees", feeAmt + denom,
		"--gas", fmt.Sprintf("%d", gasLimit),
		"--account-number", fmt.Sprintf("%d", from.AccNum),
		"--sequence", fmt.Sprintf("%d", seq),
		"--broadcast-mode", "sync",
		"--yes",
		"-o", "json",
	}

	cmd := exec.Command(binPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec: %v — %s", err, truncate(string(out), 200))
	}

	outStr := string(out)
	// Find JSON in output (may have extra lines)
	jsonStart := strings.Index(outStr, "{")
	if jsonStart < 0 {
		return fmt.Errorf("no JSON in output: %s", truncate(outStr, 200))
	}
	outStr = outStr[jsonStart:]

	var resp BroadcastResp
	if err := json.Unmarshal([]byte(outStr), &resp); err != nil {
		return fmt.Errorf("parse: %v — %s", err, truncate(outStr, 200))
	}

	if resp.TxResponse.Code != 0 {
		return fmt.Errorf("code %d: %s", resp.TxResponse.Code, truncate(resp.TxResponse.RawLog, 150))
	}
	return nil
}

// ---- REST helpers ----

func getLatestBlock() (string, error) {
	resp, err := http.Get(restURL + "/cosmos/base/tendermint/v1beta1/blocks/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var d struct {
		Block struct {
			Header struct {
				Height string `json:"height"`
			} `json:"header"`
		} `json:"block"`
	}
	json.NewDecoder(resp.Body).Decode(&d)
	return d.Block.Header.Height, nil
}

func getSequence(addr string) (uint64, error) {
	resp, err := http.Get(restURL + "/cosmos/auth/v1beta1/accounts/" + addr)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var d AuthAccountResp
	if err := json.Unmarshal(body, &d); err != nil {
		return 0, fmt.Errorf("parse: %v", err)
	}
	var seq uint64
	fmt.Sscanf(d.Account.Sequence, "%d", &seq)
	return seq, nil
}

func getAccNum(addr string) (uint64, error) {
	resp, err := http.Get(restURL + "/cosmos/auth/v1beta1/accounts/" + addr)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var d AuthAccountResp
	if err := json.Unmarshal(body, &d); err != nil {
		return 0, fmt.Errorf("parse: %v", err)
	}
	var num uint64
	fmt.Sscanf(d.Account.AccountNumber, "%d", &num)
	return num, nil
}

func isSeqErr(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "sequence") || strings.Contains(s, "incorrect account sequence")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func max1(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
