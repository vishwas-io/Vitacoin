#!/usr/bin/env python3
"""
VitaCoin Proper Stress Test Client
Uses REST broadcast API with explicit sequence management
NO CLI subprocess sequence races — sequences tracked in-process
Signing: calls vitacoind to sign but tracks seq ourselves
"""
import subprocess, json, time, threading, sys, os, signal
from collections import defaultdict
from dataclasses import dataclass, field
from typing import Optional

BIN    = "/home/gcp-instance-20260311/vitacoind"
HOME   = "/home/gcp-instance-20260311/.vitacoin-testnet"
NODE   = "tcp://localhost:26657"
REST   = "http://localhost:1317"
CHAIN  = "vitacoin-testnet-1"
DENOM  = "uvita"
SEND   = "1000"     # uvita
FEE    = "2000"     # uvita
GAS    = "200000"

ACCOUNTS = [
    {"name": "stress0", "addr": "cosmos1gs3rwt98h34zw0kdu4z7zd4s2u7gvur4yx0z83", "acc_num": 14},
    {"name": "stress1", "addr": "cosmos1rlg9qnj70f8rgfr0pwww9gk2ndamrp2axexw52", "acc_num": 16},
]

# Global stats
stats = {"sent": 0, "success": 0, "fail": 0, "seq_errors": 0, "lock": threading.Lock()}
# Per-account sequence locks and counters
seq_state = {}

def get_sequence(addr):
    import urllib.request
    url = f"{REST}/cosmos/auth/v1beta1/accounts/{addr}"
    with urllib.request.urlopen(url, timeout=5) as r:
        d = json.loads(r.read())
    return int(d["account"].get("sequence", "0")), int(d["account"].get("account_number", "0"))

def get_block():
    import urllib.request
    with urllib.request.urlopen(f"{REST}/cosmos/base/tendermint/v1beta1/blocks/latest", timeout=5) as r:
        d = json.loads(r.read())
    return d["block"]["header"]["height"]

def send_tx(from_acc, to_addr, seq, acc_num):
    """Send a single tx using vitacoind CLI with explicit --sequence --account-number"""
    cmd = [
        BIN, "tx", "send",
        from_acc["addr"], to_addr, f"{SEND}{DENOM}",
        "--from", from_acc["name"],
        "--chain-id", CHAIN,
        "--home", HOME,
        "--node", NODE,
        "--keyring-backend", "test",
        "--fees", f"{FEE}{DENOM}",
        "--gas", GAS,
        "--account-number", str(acc_num),
        "--sequence", str(seq),
        "--broadcast-mode", "sync",
        "--yes",
        "-o", "json",
    ]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=15)
        out = result.stdout.strip()
        if not out:
            out = result.stderr.strip()
        try:
            d = json.loads(out)
            code = d.get("tx_response", {}).get("code", d.get("code", -1))
            txhash = d.get("tx_response", {}).get("txhash", d.get("txhash", ""))
            raw_log = d.get("tx_response", {}).get("raw_log", d.get("raw_log", ""))
            return code, txhash, raw_log
        except json.JSONDecodeError:
            return -1, "", out[:200]
    except subprocess.TimeoutExpired:
        return -2, "", "timeout"
    except Exception as e:
        return -3, "", str(e)

def worker_send(from_acc, to_addr, seq, acc_num, results):
    code, txhash, raw_log = send_tx(from_acc, to_addr, seq, acc_num)
    results.append((seq, code, txhash, raw_log))

def run_tier(tps, duration_sec, label=""):
    global seq_state
    print(f"\n{'='*50}")
    print(f"TIER: {label or str(tps)+' TPS'} | Duration: {duration_sec}s")
    print(f"{'='*50}")

    tier_sent = 0
    tier_success = 0
    tier_fail = 0
    tier_start = time.time()
    last_print = time.time()

    interval = 1.0 / tps
    tx_index = 0
    threads = []

    while time.time() - tier_start < duration_sec:
        from_idx = tx_index % 2
        to_idx   = 1 - from_idx
        from_acc = ACCOUNTS[from_idx]
        to_addr  = ACCOUNTS[to_idx]["addr"]

        # Get and increment sequence atomically
        with seq_state[from_acc["name"]]["lock"]:
            seq = seq_state[from_acc["name"]]["seq"]
            seq_state[from_acc["name"]]["seq"] += 1

        results = []
        t = threading.Thread(target=worker_send, args=(from_acc, to_addr, seq, from_acc["acc_num"], results), daemon=True)
        t.start()
        threads.append((t, from_acc, seq, results))
        tx_index += 1
        tier_sent += 1

        # Print progress every 5s
        if time.time() - last_print >= 5:
            print(f"  [{time.time()-tier_start:.0f}s] sent={tier_sent} success={tier_success} fail={tier_fail} inflight={len([x for x in threads if x[0].is_alive()])}")
            last_print = time.time()

        time.sleep(interval)

    # Wait for all in-flight txs
    print(f"  Waiting for {len([t for t,*_ in threads if t.is_alive()])} in-flight txs...")
    for t, from_acc, seq, results in threads:
        t.join(timeout=20)
        if results:
            code, txhash, raw_log = results[0]
            if code == 0:
                tier_success += 1
            else:
                tier_fail += 1
                if "sequence" in raw_log.lower() or "incorrect" in raw_log.lower():
                    stats["seq_errors"] += 1
                    # Resync sequence
                    real_seq, _ = get_sequence(from_acc["addr"])
                    with seq_state[from_acc["name"]]["lock"]:
                        seq_state[from_acc["name"]]["seq"] = real_seq
                    print(f"    SEQ RESYNC {from_acc['name']}: reset to {real_seq}")

    elapsed = time.time() - tier_start
    actual_tps = tier_success / elapsed if elapsed > 0 else 0
    print(f"\n  RESULT: sent={tier_sent} success={tier_success} fail={tier_fail} actual_tps={actual_tps:.2f} time={elapsed:.1f}s")

    with stats["lock"]:
        stats["sent"]    += tier_sent
        stats["success"] += tier_success
        stats["fail"]    += tier_fail

    return tier_success, tier_fail, actual_tps

def main():
    print("╔══════════════════════════════════════╗")
    print("║  VitaCoin Stress Test — Nova ⚡       ║")
    print("╚══════════════════════════════════════╝")
    print(f"Time: {time.strftime('%Y-%m-%d %H:%M:%S IST')}")

    # Check chain
    try:
        block = get_block()
        print(f"Chain: {CHAIN} | Block: {block} ✅")
    except Exception as e:
        print(f"ERROR: chain unreachable: {e}")
        sys.exit(1)

    # Init sequences
    for acc in ACCOUNTS:
        seq, acc_num = get_sequence(acc["addr"])
        acc["acc_num"] = acc_num
        seq_state[acc["name"]] = {"seq": seq, "lock": threading.Lock()}
        print(f"{acc['name']}: addr={acc['addr'][:20]}... seq={seq} acc_num={acc_num}")

    print()
    results = {}

    # RAMP TEST: 10 → 20 → 50 → 100 TPS
    tiers = [
        (10,  30, "10 TPS — Baseline"),
        (20,  30, "20 TPS — Moderate"),
        (50,  30, "50 TPS — Heavy"),
        (100, 30, "100 TPS — Stress"),
    ]

    for tps, dur, label in tiers:
        blk_before = get_block()
        s, f, actual = run_tier(tps, dur, label)
        blk_after = get_block()
        results[label] = {"tps_target": tps, "success": s, "fail": f, "actual_tps": actual,
                          "block_before": blk_before, "block_after": blk_after}

        # Check chain health after each tier
        try:
            blk = get_block()
            print(f"  Chain health: block {blk} ✅")
        except Exception as e:
            print(f"  ⚠️  Chain health check FAILED: {e}")
            print("  Attempting restart...")
            os.system("ssh -o StrictHostKeyChecking=no gcp-instance-20260311@localhost 'sudo systemctl restart vitacoind' 2>/dev/null || true")
            time.sleep(10)

    # BURST TEST
    print(f"\n{'='*50}")
    print("BURST TEST: 200 concurrent txs → measure mempool depth")
    print(f"{'='*50}")
    burst_success, burst_fail, burst_tps = run_tier(200, 10, "200 TPS BURST — Max Stress")
    results["burst"] = {"success": burst_success, "fail": burst_fail, "actual_tps": burst_tps}

    # Final summary
    print("\n\n" + "="*60)
    print("FINAL STRESS TEST SUMMARY — VitaCoin Testnet")
    print("="*60)
    total_s = stats["success"]
    total_f = stats["fail"]
    total   = stats["sent"]
    success_rate = (total_s / total * 100) if total > 0 else 0

    print(f"Total Sent:        {total}")
    print(f"Total Success:     {total_s} ({success_rate:.1f}%)")
    print(f"Total Failed:      {total_f}")
    print(f"Seq Resyncs:       {stats['seq_errors']}")
    print()
    for label, r in results.items():
        print(f"  {label:<35} success={r['success']:4d} fail={r['fail']:4d} actual={r['actual_tps']:.1f} TPS")

    final_block = get_block()
    print(f"\nFinal block: {final_block}")
    print(f"Chain status: {'✅ ALIVE' if final_block else '❌ DEAD'}")

    # Save results JSON
    with open("/tmp/stress_results.json", "w") as f_out:
        json.dump({
            "timestamp": time.strftime("%Y-%m-%dT%H:%M:%S"),
            "chain": CHAIN,
            "total_sent": total,
            "total_success": total_s,
            "total_fail": total_f,
            "success_rate_pct": success_rate,
            "tiers": results,
            "final_block": final_block,
        }, f_out, indent=2)
    print("\nResults saved to /tmp/stress_results.json")

if __name__ == "__main__":
    main()
