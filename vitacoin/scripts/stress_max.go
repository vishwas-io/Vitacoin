//go:build ignore
// +build ignore

// stress_max.go — VitaCoin Max Throughput Stress Test
// Goal: find the actual chain limit by maximizing tx throughput
// Strategy: Pre-sign sequence slots, fire parallel CLI calls
// Build: go build -o stress_max stress_max.go
// Usage: ./stress_max -workers=8 -duration=60
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

var (
	txSent    int64
	txSuccess int64
	txFail    int64
)

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

func main() {
	workers  := flag.Int("workers", 4, "parallel workers per account")
	duration := flag.Int("duration", 60, "test duration in seconds")
	flag.Parse()

	fmt.Printf("=== VitaCoin MAX THROUGHPUT Test ===\n")
	fmt.Printf("Workers per account: %d | Duration: %ds\n\n", *workers, *duration)

	block, err := getLatestBlock()
	if err != nil {
		fmt.Printf("ERROR: cannot reach chain: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Chain: %s | Block: %s\n", chainID, block)

	seqs := make([]uint64, len(accounts))
	for i, acc := range accounts {
		seq, accNum, err := getSeqAndAccNum(acc.Address)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
		seqs[i] = seq
		accounts[i].AccNum = accNum
		fmt.Printf("%s seq=%d accNum=%d\n", acc.Name, seq, accNum)
	}
	fmt.Println()

	startTime := time.Now()
	deadline  := time.After(time.Duration(*duration) * time.Second)

	// Each account has its own seq counter (atomic-safe increment)
	// Multiple workers pull from the same sequence space
	atomicSeqs := make([]int64, len(accounts))
	for i, s := range seqs {
		atomicSeqs[i] = int64(s)
	}

	var wg sync.WaitGroup
	ticker := time.NewTicker(500 * time.Millisecond) // progress report

	// Launch workers
	for accIdx := range accounts {
		for w := 0; w < *workers; w++ {
			wg.Add(1)
			go func(ai int) {
				defer wg.Done()
				toIdx := (ai + 1) % len(accounts)
				for {
					select {
					case <-deadline:
						return
					default:
					}
					seq := atomic.AddInt64(&atomicSeqs[ai], 1) - 1
					err := broadcastViaCLI(accounts[ai], accounts[toIdx].Address, uint64(seq))
					atomic.AddInt64(&txSent, 1)
					if err != nil {
						atomic.AddInt64(&txFail, 1)
						if isSeqErr(err) {
							// Re-sync sequence from chain
							realSeq, _, e2 := getSeqAndAccNum(accounts[ai].Address)
							if e2 == nil {
								cur := atomic.LoadInt64(&atomicSeqs[ai])
								if int64(realSeq) > cur {
									atomic.StoreInt64(&atomicSeqs[ai], int64(realSeq))
								}
							}
						}
					} else {
						atomic.AddInt64(&txSuccess, 1)
					}
				}
			}(accIdx)
		}
	}

	// Progress reporter
	lastSuccess := int64(0)
	lastTime    := time.Now()
	go func() {
		for range ticker.C {
			now := time.Now()
			s := atomic.LoadInt64(&txSuccess)
			f := atomic.LoadInt64(&txFail)
			elapsed := now.Sub(lastTime).Seconds()
			rate := float64(s-lastSuccess) / elapsed
			lastSuccess = s
			lastTime = now
			b, _ := getLatestBlock()
			fmt.Printf("  [%.0fs] sent=%d ok=%d fail=%d | rate=%.1f TPS | block=%s\n",
				time.Since(startTime).Seconds(), s+f, s, f, rate, b)
		}
	}()

	wg.Wait()
	ticker.Stop()

	elapsed := time.Since(startTime).Seconds()
	totalSuccess := atomic.LoadInt64(&txSuccess)
	totalFail    := atomic.LoadInt64(&txFail)
	totalSent    := atomic.LoadInt64(&txSent)
	actualTPS    := float64(totalSuccess) / elapsed

	fmt.Printf("\n=== FINAL RESULTS ===\n")
	fmt.Printf("Duration:   %.1fs\n", elapsed)
	fmt.Printf("Sent:       %d\n", totalSent)
	fmt.Printf("Success:    %d (%.1f%%)\n", totalSuccess, 100*float64(totalSuccess)/max1(float64(totalSent),1))
	fmt.Printf("Failed:     %d\n", totalFail)
	fmt.Printf("Actual TPS: %.2f\n", actualTPS)
	finalBlock, _ := getLatestBlock()
	fmt.Printf("Blocks:     %s → %s\n", block, finalBlock)

	// Check chain health
	fmt.Printf("\n=== CHAIN HEALTH ===\n")
	mem, _ := getMempoolSize()
	fmt.Printf("Mempool: %d pending txs\n", mem)

	// Check on-chain sequences (actual txs committed)
	for i, acc := range accounts {
		chainSeq, _, _ := getSeqAndAccNum(acc.Address)
		claimed := atomic.LoadInt64(&atomicSeqs[i])
		fmt.Printf("%s: chain seq=%d | claimed seq=%d | on-chain txs≈%d\n",
			acc.Name, chainSeq, claimed, int64(chainSeq)-int64(seqs[i]))
	}
}

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
	jsonStart := strings.Index(outStr, "{")
	if jsonStart < 0 {
		return fmt.Errorf("no JSON: %s", truncate(outStr, 200))
	}
	var resp BroadcastResp
	if err := json.Unmarshal([]byte(outStr[jsonStart:]), &resp); err != nil {
		return fmt.Errorf("parse: %v", err)
	}
	if resp.TxResponse.Code != 0 {
		return fmt.Errorf("code %d: %s", resp.TxResponse.Code, truncate(resp.TxResponse.RawLog, 150))
	}
	return nil
}

func getSeqAndAccNum(addr string) (seq, accNum uint64, err error) {
	resp, err := http.Get(restURL + "/cosmos/auth/v1beta1/accounts/" + addr)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var d AuthAccountResp
	if err := json.Unmarshal(body, &d); err != nil {
		return 0, 0, fmt.Errorf("parse: %v", err)
	}
	fmt.Sscanf(d.Account.Sequence, "%d", &seq)
	fmt.Sscanf(d.Account.AccountNumber, "%d", &accNum)
	return seq, accNum, nil
}

func getLatestBlock() (string, error) {
	resp, err := http.Get(restURL + "/cosmos/base/tendermint/v1beta1/blocks/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var d struct {
		Block struct{ Header struct{ Height string `json:"height"` } `json:"header"` } `json:"block"`
	}
	json.NewDecoder(resp.Body).Decode(&d)
	return d.Block.Header.Height, nil
}

func getMempoolSize() (int, error) {
	resp, err := http.Get("http://localhost:26657/num_unconfirmed_txs")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var d struct {
		Result struct{ NTxs string `json:"n_txs"` } `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&d)
	var n int
	fmt.Sscanf(d.Result.NTxs, "%d", &n)
	return n, nil
}

func isSeqErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "sequence") || strings.Contains(err.Error(), "incorrect account sequence")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func max1(a, b float64) float64 {
	if a > b { return a }
	return b
}
