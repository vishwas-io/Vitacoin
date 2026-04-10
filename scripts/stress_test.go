//go:build ignore
// +build ignore

// stress_test.go — VitaCoin Testnet Stress Test Tool
// Usage: go run scripts/stress_test.go --mode light|medium|heavy|burst --rpc http://localhost:26657
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ─── Config ──────────────────────────────────────────────────────────────────

type Config struct {
	RPC     string
	REST    string
	ChainID string
	Denom   string
	Gas     int
	Fees    string

	Mode     string
	TxPerSec int
	Duration time.Duration
	Accounts int
}

var modes = map[string][2]int{
	"light":  {10, 60},
	"medium": {50, 60},
	"heavy":  {100, 30},
	"burst":  {500, 10},
}

// ─── Types ───────────────────────────────────────────────────────────────────

type Account struct {
	Name    string
	Address string
	mu      sync.Mutex
	Seq     uint64
}

type TxResult struct {
	Success bool
	Latency time.Duration
	Error   string
}

type MempoolSnapshot struct {
	Time  time.Time
	Count int
}

// ─── RPC Helpers ─────────────────────────────────────────────────────────────

func rpcGET(rpc, path string) (map[string]interface{}, error) {
	resp, err := http.Get(rpc + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out map[string]interface{}
	err = json.Unmarshal(body, &out)
	return out, err
}

func getMempoolSize(rpc string) int {
	data, err := rpcGET(rpc, "/num_unconfirmed_txs")
	if err != nil {
		return -1
	}
	if result, ok := data["result"].(map[string]interface{}); ok {
		if n, ok := result["n_txs"].(string); ok {
			var count int
			fmt.Sscan(n, &count)
			return count
		}
	}
	return 0
}

func getBlockTime(rpc string) (time.Time, int64, error) {
	data, err := rpcGET(rpc, "/status")
	if err != nil {
		return time.Time{}, 0, err
	}
	result, ok := data["result"].(map[string]interface{})
	if !ok {
		return time.Time{}, 0, fmt.Errorf("bad result")
	}
	syncInfo, ok := result["sync_info"].(map[string]interface{})
	if !ok {
		return time.Time{}, 0, fmt.Errorf("no sync_info")
	}
	latestTime, _ := syncInfo["latest_block_time"].(string)
	latestHeight, _ := syncInfo["latest_block_height"].(string)
	t, _ := time.Parse(time.RFC3339Nano, latestTime)
	var h int64
	fmt.Sscan(latestHeight, &h)
	return t, h, nil
}

// ─── Account Helpers ─────────────────────────────────────────────────────────

func vitacoind(args ...string) (string, error) {
	cmd := exec.Command("vitacoind", args...)
	cmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH")+":/usr/local/go/bin")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, stderr.String())
	}
	return strings.TrimSpace(out.String()), nil
}

func createAccount(name string) (*Account, error) {
	out, err := vitacoind("keys", "add", name, "--output", "json", "--keyring-backend", "test")
	if err != nil {
		// try to recover existing key
		out, err = vitacoind("keys", "show", name, "--output", "json", "--keyring-backend", "test")
		if err != nil {
			return nil, fmt.Errorf("keys add/show %s: %w", name, err)
		}
	}
	var keyInfo map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(out), &keyInfo); jsonErr != nil {
		return nil, fmt.Errorf("parse key json: %w", jsonErr)
	}
	addr, _ := keyInfo["address"].(string)
	if addr == "" {
		return nil, fmt.Errorf("empty address for %s", name)
	}
	return &Account{Name: name, Address: addr}, nil
}

func fundAccount(cfg *Config, faucetName, toAddr string, amount int64) error {
	amtStr := fmt.Sprintf("%d%s", amount, cfg.Denom)
	_, err := vitacoind("tx", "bank", "send",
		faucetName, toAddr, amtStr,
		"--chain-id", cfg.ChainID,
		"--node", cfg.RPC,
		"--gas", fmt.Sprint(cfg.Gas),
		"--fees", cfg.Fees,
		"--keyring-backend", "test",
		"--yes", "--broadcast-mode", "sync",
		"--output", "json",
	)
	return err
}

func getSequence(cfg *Config, addr string) (uint64, uint64, error) {
	url := fmt.Sprintf("%s/cosmos/auth/v1beta1/accounts/%s", cfg.REST, addr)
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Account struct {
			Sequence      string `json:"sequence"`
			AccountNumber string `json:"account_number"`
		} `json:"account"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, 0, err
	}
	var seq, accNum uint64
	fmt.Sscan(result.Account.Sequence, &seq)
	fmt.Sscan(result.Account.AccountNumber, &accNum)
	return seq, accNum, nil
}

// ─── Transaction Sending ─────────────────────────────────────────────────────

func sendTx(cfg *Config, from *Account, toAddr string, amount int64) TxResult {
	start := time.Now()
	from.mu.Lock()
	seq := from.Seq
	from.Seq++
	from.mu.Unlock()

	amtStr := fmt.Sprintf("%d%s", amount, cfg.Denom)
	args := []string{
		"tx", "bank", "send",
		from.Name, toAddr, amtStr,
		"--chain-id", cfg.ChainID,
		"--node", cfg.RPC,
		"--gas", fmt.Sprint(cfg.Gas),
		"--fees", cfg.Fees,
		"--keyring-backend", "test",
		"--sequence", fmt.Sprint(seq),
		"--yes", "--broadcast-mode", "async",
		"--output", "json",
	}
	out, err := vitacoind(args...)
	latency := time.Since(start)

	if err != nil {
		return TxResult{false, latency, err.Error()}
	}

	var txResp map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(out), &txResp); jsonErr != nil {
		return TxResult{false, latency, "parse error: " + out[:min(100, len(out))]}
	}

	code, _ := txResp["code"].(float64)
	if code != 0 {
		rawLog, _ := txResp["raw_log"].(string)
		// sequence mismatch: resync
		if strings.Contains(rawLog, "account sequence mismatch") {
			seq2, _, err2 := getSequence(cfg, from.Address)
			if err2 == nil {
				from.mu.Lock()
				from.Seq = seq2 + 1
				from.mu.Unlock()
			}
		}
		return TxResult{false, latency, fmt.Sprintf("code=%v: %s", code, rawLog)}
	}
	return TxResult{true, latency, ""}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ─── Block Tracker ───────────────────────────────────────────────────────────

type BlockSample struct {
	Height    int64
	Timestamp time.Time
}

func trackBlocks(ctx context.Context, cfg *Config, samples *[]BlockSample, mu *sync.Mutex) {
	var lastHeight int64
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
			t, h, err := getBlockTime(cfg.RPC)
			if err == nil && h != lastHeight {
				mu.Lock()
				*samples = append(*samples, BlockSample{h, t})
				mu.Unlock()
				lastHeight = h
			}
		}
	}
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	mode := flag.String("mode", "light", "Stress test mode: light|medium|heavy|burst")
	rpc := flag.String("rpc", "http://34.93.89.182:26657", "RPC endpoint")
	rest := flag.String("rest", "http://34.93.89.182:1317", "REST endpoint")
	chainID := flag.String("chain-id", "vitacoin-testnet-1", "Chain ID")
	accounts := flag.Int("accounts", 10, "Number of test accounts")
	faucet := flag.String("faucet", "faucet", "Faucet keyring name")
	flag.Parse()

	modeParams, ok := modes[*mode]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown mode %q. Choose: light, medium, heavy, burst\n", *mode)
		os.Exit(1)
	}

	cfg := &Config{
		RPC:      *rpc,
		REST:     *rest,
		ChainID:  *chainID,
		Denom:    "uvita",
		Gas:      200000,
		Fees:     "50000uvita",
		Mode:     *mode,
		TxPerSec: modeParams[0],
		Duration: time.Duration(modeParams[1]) * time.Second,
		Accounts: *accounts,
	}

	fmt.Printf("⚡ VitaCoin Stress Test — Mode: %s (%d tx/s for %s)\n", cfg.Mode, cfg.TxPerSec, cfg.Duration)
	fmt.Printf("   RPC: %s | Chain: %s\n\n", cfg.RPC, cfg.ChainID)

	// ── Create accounts ────────────────────────────────────────────────────
	fmt.Printf("📦 Creating %d test accounts...\n", cfg.Accounts)
	accts := make([]*Account, cfg.Accounts)
	for i := 0; i < cfg.Accounts; i++ {
		name := fmt.Sprintf("stress-acct-%d", i)
		a, err := createAccount(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ %s: %v\n", name, err)
			os.Exit(1)
		}
		accts[i] = a
		fmt.Printf("  ✓ %s → %s\n", name, a.Address)
	}

	// ── Fund accounts ──────────────────────────────────────────────────────
	fundAmt := int64(10_000_000) // 10 VITA each
	fmt.Printf("\n💸 Funding accounts from '%s' (%d uvita each)...\n", *faucet, fundAmt)
	for _, a := range accts {
		if err := fundAccount(cfg, *faucet, a.Address, fundAmt); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ fund %s: %v\n", a.Address, err)
		} else {
			fmt.Printf("  ✓ funded %s\n", a.Name)
		}
		time.Sleep(500 * time.Millisecond) // avoid sequence errors on faucet
	}

	// ── Sync sequences ─────────────────────────────────────────────────────
	fmt.Println("\n🔄 Syncing account sequences...")
	time.Sleep(3 * time.Second) // wait for funding txs to land
	for _, a := range accts {
		seq, _, err := getSequence(cfg, a.Address)
		if err == nil {
			a.Seq = seq
			fmt.Printf("  ✓ %s seq=%d\n", a.Name, seq)
		}
	}

	// ── Start block tracker ────────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Duration+30*time.Second)
	defer cancel()

	var blockSamples []BlockSample
	var blockMu sync.Mutex
	go trackBlocks(ctx, cfg, &blockSamples, &blockMu)

	// ── Mempool poller ─────────────────────────────────────────────────────
	var mempoolSnaps []MempoolSnapshot
	var mempoolMu sync.Mutex
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				n := getMempoolSize(cfg.RPC)
				mempoolMu.Lock()
				mempoolSnaps = append(mempoolSnaps, MempoolSnapshot{time.Now(), n})
				mempoolMu.Unlock()
			}
		}
	}()

	// ── Send transactions ──────────────────────────────────────────────────
	fmt.Printf("\n🚀 Stress test starting: %d tx/s for %s\n", cfg.TxPerSec, cfg.Duration)
	fmt.Print("   Press Ctrl+C to abort early.\n\n")

	var (
		totalSent    int64
		totalSuccess int64
		totalFailed  int64
		latencySum   int64 // nanoseconds
	)

	results := make(chan TxResult, cfg.TxPerSec*int(cfg.Duration.Seconds())+1000)
	var wg sync.WaitGroup

	deadline := time.Now().Add(cfg.Duration)
	ticker := time.NewTicker(time.Second / time.Duration(cfg.TxPerSec))
	defer ticker.Stop()

	startTime := time.Now()
	lastReport := time.Now()

	for time.Now().Before(deadline) {
		<-ticker.C

		// pick random sender and receiver (different)
		fromIdx := rand.Intn(len(accts))
		toIdx := rand.Intn(len(accts))
		for toIdx == fromIdx {
			toIdx = rand.Intn(len(accts))
		}
		from := accts[fromIdx]
		to := accts[toIdx]

		atomic.AddInt64(&totalSent, 1)
		wg.Add(1)
		go func(f *Account, toAddr string) {
			defer wg.Done()
			r := sendTx(cfg, f, toAddr, 1000) // send 1000 uvita
			results <- r
		}(from, to.Address)

		// Progress report every 5s
		if time.Since(lastReport) >= 5*time.Second {
			elapsed := time.Since(startTime).Seconds()
			s := atomic.LoadInt64(&totalSuccess)
			f := atomic.LoadInt64(&totalFailed)
			sent := atomic.LoadInt64(&totalSent)
			tps := float64(s+f) / elapsed
			fmt.Printf("  [%4.0fs] sent=%d  ok=%d  fail=%d  TPS=%.1f\n",
				elapsed, sent, s, f, tps)
			lastReport = time.Now()
		}
	}

	// Drain results
	go func() {
		wg.Wait()
		close(results)
	}()
	for r := range results {
		if r.Success {
			atomic.AddInt64(&totalSuccess, 1)
			atomic.AddInt64(&latencySum, r.Latency.Nanoseconds())
		} else {
			atomic.AddInt64(&totalFailed, 1)
		}
	}

	elapsed := time.Since(startTime)

	// ── Block time analysis ────────────────────────────────────────────────
	blockMu.Lock()
	bs := blockSamples
	blockMu.Unlock()

	var avgBlockTime float64
	var minBT, maxBT float64
	if len(bs) >= 2 {
		var diffs []float64
		for i := 1; i < len(bs); i++ {
			d := bs[i].Timestamp.Sub(bs[i-1].Timestamp).Seconds()
			diffs = append(diffs, d)
		}
		sum := 0.0
		minBT = diffs[0]
		maxBT = diffs[0]
		for _, d := range diffs {
			sum += d
			if d < minBT {
				minBT = d
			}
			if d > maxBT {
				maxBT = d
			}
		}
		avgBlockTime = sum / float64(len(diffs))
	}

	// ── Mempool analysis ───────────────────────────────────────────────────
	mempoolMu.Lock()
	ms := mempoolSnaps
	mempoolMu.Unlock()

	maxMempool := 0
	for _, s := range ms {
		if s.Count > maxMempool {
			maxMempool = s.Count
		}
	}

	// ── Summary ────────────────────────────────────────────────────────────
	success := atomic.LoadInt64(&totalSuccess)
	failed := atomic.LoadInt64(&totalFailed)
	sent := atomic.LoadInt64(&totalSent)
	actualTPS := float64(success+failed) / elapsed.Seconds()
	successTPS := float64(success) / elapsed.Seconds()
	failRate := 0.0
	if sent > 0 {
		failRate = float64(failed) / float64(sent) * 100
	}
	avgLatency := time.Duration(0)
	if success > 0 {
		avgLatency = time.Duration(latencySum / success)
	}

	fmt.Printf("\n")
	fmt.Println("╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║        ⚡ VitaCoin Stress Test — Results Summary        ║")
	fmt.Println("╠═══════════════════════════════════════════════════════╣")
	fmt.Printf("║  Mode:              %-34s║\n", cfg.Mode)
	fmt.Printf("║  Target TPS:        %-34d║\n", cfg.TxPerSec)
	fmt.Printf("║  Duration:          %-34s║\n", elapsed.Round(time.Millisecond))
	fmt.Println("╠═══════════════════════════════════════════════════════╣")
	fmt.Printf("║  Total Sent:        %-34d║\n", sent)
	fmt.Printf("║  Successful:        %-34d║\n", success)
	fmt.Printf("║  Failed:            %-34d║\n", failed)
	fmt.Printf("║  Fail Rate:         %-33.1f%%║\n", failRate)
	fmt.Println("╠═══════════════════════════════════════════════════════╣")
	fmt.Printf("║  Actual TPS (all):  %-34.2f║\n", actualTPS)
	fmt.Printf("║  Success TPS:       %-34.2f║\n", successTPS)
	fmt.Printf("║  Avg Tx Latency:    %-34s║\n", avgLatency.Round(time.Millisecond))
	fmt.Println("╠═══════════════════════════════════════════════════════╣")
	if len(bs) >= 2 {
		fmt.Printf("║  Blocks Observed:   %-34d║\n", len(bs))
		fmt.Printf("║  Avg Block Time:    %-33.2fs║\n", avgBlockTime)
		fmt.Printf("║  Min Block Time:    %-33.2fs║\n", minBT)
		fmt.Printf("║  Max Block Time:    %-33.2fs║\n", maxBT)
	} else {
		fmt.Printf("║  Block data: insufficient samples                     ║\n")
	}
	fmt.Println("╠═══════════════════════════════════════════════════════╣")
	fmt.Printf("║  Max Mempool Size:  %-34d║\n", maxMempool)
	fmt.Print("╚═══════════════════════════════════════════════════════╝\n")
}
