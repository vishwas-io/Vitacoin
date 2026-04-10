#!/usr/bin/env python3
"""
stress_broadcast.py — VitaCoin REST Broadcast Stress Tester
============================================================

USAGE:
    python3 scripts/stress_broadcast.py [OPTIONS]

OPTIONS:
    --accounts N       Number of stress accounts to use (default: 10, range: 1-10)
    --txs-per-account  Number of txs to pre-sign per account (default: 50)
    --tps N            Target broadcasts per second (default: 50)
    --duration N       Test duration in seconds (default: 30)
    --ramp-up N        Ramp-up period in seconds (default: 5)
    --rest-url URL     REST API base URL (default: http://34.93.89.182:1317)
    --chain-id ID      Chain ID (default: vitacoin-testnet-1)
    --keyring-dir DIR  Keyring directory (default: /root/.vitacoin-testnet)
    --vitacoind PATH   Path to vitacoind binary (default: vitacoind)
    --dry-run          Show what would be done without sending any txs
    --skip-presign     Skip pre-signing phase (use cached signed txs from /tmp)
    --workers N        Number of concurrent HTTP workers (default: 20)

DESCRIPTION:
    This script separates signing (slow CLI subprocess) from broadcasting (fast HTTP).
    
    Phase 1 — Pre-sign:
        Uses vitacoind CLI to generate-only + sign N txs per account offline.
        Txs vary in memo and amount to avoid mempool deduplication.
        Each signed tx JSON is stored in memory (or /tmp for --skip-presign runs).
    
    Phase 2 — Broadcast:
        All signed txs are broadcast via REST POST to /cosmos/tx/v1beta1/txs
        using a thread pool. A token-bucket rate limiter enforces the --tps target.
        Actual TPS, latency percentiles (p50/p95/p99), and fail rates are reported.

PREREQUISITES:
    - vitacoind binary in PATH (or specify --vitacoind)
    - Keyring with stress0..stress9 keys at --keyring-dir
    - python3 with standard library only (no pip installs needed)

EXAMPLE:
    # Dry run to see what it would do
    python3 scripts/stress_broadcast.py --accounts 5 --txs-per-account 20 --dry-run

    # Full test: 10 accounts, 50 txs each, target 100 TPS, 30s duration
    python3 scripts/stress_broadcast.py --accounts 10 --txs-per-account 100 --tps 100 --duration 60

    # Quick sanity check
    python3 scripts/stress_broadcast.py --accounts 2 --txs-per-account 5 --tps 10 --duration 5

CHAIN DEFAULTS:
    RPC:   http://34.93.89.182:26657
    gRPC:  34.93.89.182:9090
    REST:  http://34.93.89.182:1317
    Chain: vitacoin-testnet-1
    Denom: uvita
"""

import argparse
import json
import os
import subprocess
import sys
import tempfile
import threading
import time
import urllib.request
import urllib.error
from collections import defaultdict
from concurrent.futures import ThreadPoolExecutor, as_completed
from statistics import median, quantiles

# ─── Constants ────────────────────────────────────────────────────────────────
DEFAULT_REST_URL   = "http://34.93.89.182:1317"
DEFAULT_CHAIN_ID   = "vitacoin-testnet-1"
DEFAULT_KEYRING    = "/root/.vitacoin-testnet"
DEFAULT_VITACOIND  = "vitacoind"
DEFAULT_ACCOUNTS   = 10
DEFAULT_TXS_PA     = 50
DEFAULT_TPS        = 50
DEFAULT_DURATION   = 30
DEFAULT_RAMP_UP    = 5
DEFAULT_WORKERS    = 20
DENOM              = "uvita"
DEST_ADDR          = "cosmos1rh5vuyeh2dm4afm6hzldrj368s27pq6wvsn6td"  # faucet addr
GAS_LIMIT          = 80000
FEE_AMOUNT         = "2000uvita"
MIN_AMOUNT         = 1
MAX_AMOUNT         = 100

# ─── Helpers ──────────────────────────────────────────────────────────────────

def log(msg: str, level: str = "INFO"):
    ts = time.strftime("%H:%M:%S")
    print(f"[{ts}] [{level}] {msg}", flush=True)


def rest_get(url: str, timeout: int = 10) -> dict:
    req = urllib.request.Request(url, headers={"Accept": "application/json"})
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read())


def rest_post(url: str, body: dict, timeout: int = 10) -> dict:
    data = json.dumps(body).encode()
    req = urllib.request.Request(
        url, data=data,
        headers={"Content-Type": "application/json", "Accept": "application/json"},
        method="POST"
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read())


def get_account_info(rest_url: str, address: str) -> tuple[int, int]:
    """Returns (account_number, sequence)."""
    url = f"{rest_url}/cosmos/auth/v1beta1/accounts/{address}"
    data = rest_get(url)
    acct = data.get("account", {})
    # Handle base_account wrapper
    if "base_account" in acct:
        acct = acct["base_account"]
    return int(acct.get("account_number", 0)), int(acct.get("sequence", 0))


def get_address(vitacoind: str, keyring_dir: str, key_name: str) -> str:
    result = subprocess.run(
        [vitacoind, "keys", "show", key_name, "--address",
         "--keyring-backend", "test", "--keyring-dir", keyring_dir],
        capture_output=True, text=True, timeout=15
    )
    if result.returncode != 0:
        raise RuntimeError(f"Failed to get address for {key_name}: {result.stderr.strip()}")
    return result.stdout.strip()


def presign_tx(vitacoind: str, keyring_dir: str, chain_id: str,
               from_key: str, from_addr: str, to_addr: str,
               amount: int, denom: str, account_number: int,
               sequence: int, memo: str) -> dict:
    """Generate and sign a tx offline, return the signed tx JSON dict."""
    # Step 1: generate-only
    gen_cmd = [
        vitacoind, "tx", "send",
        from_addr, to_addr, f"{amount}{denom}",
        "--chain-id", chain_id,
        "--gas", str(GAS_LIMIT),
        "--fees", FEE_AMOUNT,
        "--note", memo,
        "--sequence", str(sequence),
        "--account-number", str(account_number),
        "--generate-only",
        "--keyring-backend", "test",
        "--keyring-dir", keyring_dir,
        "--output", "json",
    ]
    gen_result = subprocess.run(gen_cmd, capture_output=True, text=True, timeout=30)
    if gen_result.returncode != 0:
        raise RuntimeError(f"generate-only failed for {from_key} seq={sequence}: {gen_result.stderr.strip()}")

    # Write unsigned tx to temp file for signing
    with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
        f.write(gen_result.stdout)
        unsigned_path = f.name

    signed_path = unsigned_path + ".signed"
    try:
        sign_cmd = [
            vitacoind, "tx", "sign", unsigned_path,
            "--from", from_key,
            "--chain-id", chain_id,
            "--sequence", str(sequence),
            "--account-number", str(account_number),
            "--offline",
            "--keyring-backend", "test",
            "--keyring-dir", keyring_dir,
            "--output", "json",
        ]
        sign_result = subprocess.run(sign_cmd, capture_output=True, text=True, timeout=30)
        if sign_result.returncode != 0:
            raise RuntimeError(f"sign failed for {from_key} seq={sequence}: {sign_result.stderr.strip()}")

        with open(signed_path, 'w') as f:
            f.write(sign_result.stdout)

        # Encode to broadcast format
        encode_cmd = [
            vitacoind, "tx", "encode", signed_path,
            "--keyring-backend", "test",
            "--keyring-dir", keyring_dir,
        ]
        encode_result = subprocess.run(encode_cmd, capture_output=True, text=True, timeout=30)
        if encode_result.returncode == 0 and encode_result.stdout.strip():
            # Return base64-encoded tx for broadcast_mode=BROADCAST_MODE_SYNC
            return {"tx_bytes": encode_result.stdout.strip(), "mode": "BROADCAST_MODE_SYNC"}
        else:
            # Fall back to JSON broadcast
            return {"tx": json.loads(sign_result.stdout), "mode": "BROADCAST_MODE_SYNC"}
    finally:
        os.unlink(unsigned_path)
        if os.path.exists(signed_path):
            os.unlink(signed_path)


def broadcast_tx(rest_url: str, signed_payload: dict) -> tuple[bool, str, float]:
    """
    Broadcast a signed tx. Returns (success, tx_hash_or_error, latency_ms).
    """
    url = f"{rest_url}/cosmos/tx/v1beta1/txs"
    start = time.monotonic()
    try:
        resp = rest_post(url, signed_payload, timeout=15)
        latency_ms = (time.monotonic() - start) * 1000
        tx_response = resp.get("tx_response", {})
        code = tx_response.get("code", -1)
        txhash = tx_response.get("txhash", "")
        if code == 0:
            return True, txhash, latency_ms
        else:
            raw_log = tx_response.get("raw_log", "unknown error")
            return False, f"code={code}: {raw_log}", latency_ms
    except urllib.error.HTTPError as e:
        latency_ms = (time.monotonic() - start) * 1000
        try:
            body = e.read().decode()
        except Exception:
            body = str(e)
        return False, f"HTTP {e.code}: {body[:200]}", latency_ms
    except Exception as e:
        latency_ms = (time.monotonic() - start) * 1000
        return False, str(e)[:200], latency_ms


# ─── Token Bucket Rate Limiter ─────────────────────────────────────────────────

class TokenBucket:
    def __init__(self, rate: float):
        self.rate = rate  # tokens per second
        self.tokens = rate
        self.last = time.monotonic()
        self._lock = threading.Lock()

    def acquire(self):
        with self._lock:
            now = time.monotonic()
            elapsed = now - self.last
            self.tokens = min(self.rate, self.tokens + elapsed * self.rate)
            self.last = now
            if self.tokens >= 1.0:
                self.tokens -= 1.0
                return True
            return False

    def wait_and_acquire(self):
        while True:
            if self.acquire():
                return
            time.sleep(0.001)


# ─── Main Logic ───────────────────────────────────────────────────────────────

def parse_args():
    p = argparse.ArgumentParser(description="VitaCoin REST Broadcast Stress Tester")
    p.add_argument("--accounts",        type=int,   default=DEFAULT_ACCOUNTS)
    p.add_argument("--txs-per-account", type=int,   default=DEFAULT_TXS_PA)
    p.add_argument("--tps",             type=float, default=DEFAULT_TPS)
    p.add_argument("--duration",        type=int,   default=DEFAULT_DURATION)
    p.add_argument("--ramp-up",         type=int,   default=DEFAULT_RAMP_UP)
    p.add_argument("--rest-url",        type=str,   default=DEFAULT_REST_URL)
    p.add_argument("--chain-id",        type=str,   default=DEFAULT_CHAIN_ID)
    p.add_argument("--keyring-dir",     type=str,   default=DEFAULT_KEYRING)
    p.add_argument("--vitacoind",       type=str,   default=DEFAULT_VITACOIND)
    p.add_argument("--workers",         type=int,   default=DEFAULT_WORKERS)
    p.add_argument("--dry-run",         action="store_true")
    p.add_argument("--skip-presign",    action="store_true",
                   help="Load previously cached signed txs from /tmp/vita_stress_*.json")
    return p.parse_args()


def run_presign_phase(args, accounts_info: list[dict], dry_run: bool) -> list[dict]:
    """
    Pre-sign all txs for all accounts.
    Returns flat list of {payload, account, seq} dicts.
    """
    total = args.accounts * args.txs_per_account
    log(f"Pre-signing {total} txs ({args.accounts} accounts × {args.txs_per_account} txs each)...")

    if dry_run:
        log("[DRY RUN] Would sign txs with varying amounts and memos", "DRY")
        for acct in accounts_info:
            log(f"  {acct['key']}: addr={acct['address']}, acct_num={acct['account_number']}, seq_start={acct['sequence']}", "DRY")
        log(f"[DRY RUN] Would broadcast to: {args.rest_url}/cosmos/tx/v1beta1/txs", "DRY")
        return []

    signed_txs = []
    errors = 0

    for acct in accounts_info:
        log(f"  Signing {args.txs_per_account} txs for {acct['key']} (seq={acct['sequence']})...")
        for i in range(args.txs_per_account):
            seq = acct['sequence'] + i
            amount = MIN_AMOUNT + (i % (MAX_AMOUNT - MIN_AMOUNT + 1))
            memo = f"stress-{acct['key']}-seq{seq}-t{int(time.time())}-i{i}"
            try:
                payload = presign_tx(
                    vitacoind=args.vitacoind,
                    keyring_dir=args.keyring_dir,
                    chain_id=args.chain_id,
                    from_key=acct['key'],
                    from_addr=acct['address'],
                    to_addr=DEST_ADDR,
                    amount=amount,
                    denom=DENOM,
                    account_number=acct['account_number'],
                    sequence=seq,
                    memo=memo,
                )
                signed_txs.append({
                    "payload": payload,
                    "account": acct['key'],
                    "seq": seq,
                })
            except Exception as e:
                log(f"    SIGN ERROR {acct['key']} seq={seq}: {e}", "WARN")
                errors += 1

    log(f"Pre-sign complete: {len(signed_txs)} signed, {errors} errors")
    return signed_txs


def run_broadcast_phase(args, signed_txs: list[dict], dry_run: bool):
    """
    Broadcast all signed txs with rate limiting and concurrency.
    Prints statistics at the end.
    """
    if dry_run or not signed_txs:
        log(f"[DRY RUN] Would broadcast {len(signed_txs) if signed_txs else args.accounts * args.txs_per_account} txs at {args.tps} TPS over {args.duration}s with {args.workers} workers", "DRY")
        return

    log(f"\n{'='*60}")
    log(f"Starting broadcast: {len(signed_txs)} txs, target {args.tps} TPS")
    log(f"Duration: {args.duration}s | Ramp-up: {args.ramp_up}s | Workers: {args.workers}")
    log(f"{'='*60}\n")

    bucket = TokenBucket(args.tps)
    results = {
        "sent": 0, "success": 0, "failed": 0,
        "latencies": [], "errors": defaultdict(int),
    }
    lock = threading.Lock()
    start_time = time.monotonic()
    stop_time = start_time + args.duration + args.ramp_up

    tx_queue = list(signed_txs)
    tx_index = 0
    tx_lock = threading.Lock()

    def worker():
        nonlocal tx_index
        while True:
            now = time.monotonic()
            if now >= stop_time:
                break

            # Ramp-up: during ramp period, scale TPS linearly
            elapsed = now - start_time
            if elapsed < args.ramp_up and args.ramp_up > 0:
                ramp_factor = elapsed / args.ramp_up
                bucket.rate = args.tps * ramp_factor
            else:
                bucket.rate = args.tps

            bucket.wait_and_acquire()

            with tx_lock:
                if tx_index >= len(tx_queue):
                    break
                item = tx_queue[tx_index]
                tx_index += 1

            success, result, latency_ms = broadcast_tx(args.rest_url, item["payload"])

            with lock:
                results["sent"] += 1
                results["latencies"].append(latency_ms)
                if success:
                    results["success"] += 1
                else:
                    results["failed"] += 1
                    # Bucket error by first word
                    err_key = result[:50].split(":")[0].strip()
                    results["errors"][err_key] += 1

            # Progress every 10 txs
            if results["sent"] % 10 == 0:
                elapsed2 = time.monotonic() - start_time
                actual_tps = results["sent"] / max(elapsed2, 0.001)
                log(f"  Progress: {results['sent']} sent | {results['success']} ok | {results['failed']} fail | {actual_tps:.1f} TPS")

    with ThreadPoolExecutor(max_workers=args.workers) as executor:
        futures = [executor.submit(worker) for _ in range(args.workers)]
        for f in as_completed(futures):
            try:
                f.result()
            except Exception as e:
                log(f"Worker error: {e}", "WARN")

    # ── Final Stats ──────────────────────────────────────────────────────────
    elapsed_total = time.monotonic() - start_time
    actual_tps = results["sent"] / max(elapsed_total, 0.001)
    success_rate = (results["success"] / max(results["sent"], 1)) * 100

    lats = sorted(results["latencies"])
    n = len(lats)
    p50 = lats[int(n * 0.50)] if n else 0
    p95 = lats[int(n * 0.95)] if n else 0
    p99 = lats[int(n * 0.99)] if n else 0
    avg = sum(lats) / n if n else 0

    print()
    print("=" * 60)
    print("  STRESS TEST RESULTS")
    print("=" * 60)
    print(f"  Duration:        {elapsed_total:.1f}s")
    print(f"  Total sent:      {results['sent']}")
    print(f"  Confirmed:       {results['success']}  ({success_rate:.1f}%)")
    print(f"  Failed:          {results['failed']}")
    print(f"  Actual TPS:      {actual_tps:.2f}")
    print(f"  Target TPS:      {args.tps}")
    print()
    print(f"  Latency (ms):")
    print(f"    avg:  {avg:.1f}")
    print(f"    p50:  {p50:.1f}")
    print(f"    p95:  {p95:.1f}")
    print(f"    p99:  {p99:.1f}")
    if results["errors"]:
        print()
        print("  Error breakdown:")
        for err, count in sorted(results["errors"].items(), key=lambda x: -x[1])[:5]:
            print(f"    [{count}] {err}")
    print("=" * 60)


def main():
    args = parse_args()

    if args.dry_run:
        log("=== DRY RUN MODE — no txs will be sent ===")

    log(f"REST endpoint:  {args.rest_url}")
    log(f"Chain ID:       {args.chain_id}")
    log(f"Accounts:       {args.accounts} (stress0..stress{args.accounts-1})")
    log(f"Txs per acct:   {args.txs_per_account}")
    log(f"Target TPS:     {args.tps}")
    log(f"Duration:       {args.duration}s (+{args.ramp_up}s ramp-up)")
    log(f"Workers:        {args.workers}")

    # ── Check REST connectivity ──────────────────────────────────────────────
    if not args.dry_run:
        try:
            node_info = rest_get(f"{args.rest_url}/cosmos/base/tendermint/v1beta1/node_info", timeout=5)
            chain_id = node_info.get("default_node_info", {}).get("network", "unknown")
            log(f"Connected to chain: {chain_id}")
            if chain_id != args.chain_id:
                log(f"WARNING: expected {args.chain_id}, got {chain_id}", "WARN")
        except Exception as e:
            log(f"Could not connect to REST endpoint: {e}", "ERROR")
            sys.exit(1)

    # ── Gather account info ──────────────────────────────────────────────────
    accounts_info = []
    for i in range(args.accounts):
        key_name = f"stress{i}"
        if args.dry_run:
            accounts_info.append({
                "key": key_name, "address": f"vita1dryrun{i}",
                "account_number": i, "sequence": 0,
            })
            continue
        try:
            addr = get_address(args.vitacoind, args.keyring_dir, key_name)
            acct_num, seq = get_account_info(args.rest_url, addr)
            accounts_info.append({
                "key": key_name, "address": addr,
                "account_number": acct_num, "sequence": seq,
            })
            log(f"  {key_name}: {addr} | acct={acct_num} seq={seq}")
        except Exception as e:
            log(f"  Skipping {key_name}: {e}", "WARN")

    if not accounts_info:
        log("No accounts available. Exiting.", "ERROR")
        sys.exit(1)

    # ── Pre-sign phase ───────────────────────────────────────────────────────
    if args.skip_presign:
        log("Loading cached signed txs from /tmp/vita_stress_*.json ...")
        import glob
        signed_txs = []
        for path in sorted(glob.glob("/tmp/vita_stress_*.json")):
            with open(path) as f:
                signed_txs.append(json.load(f))
        log(f"Loaded {len(signed_txs)} cached txs")
    else:
        signed_txs = run_presign_phase(args, accounts_info, args.dry_run)

    # ── Broadcast phase ──────────────────────────────────────────────────────
    run_broadcast_phase(args, signed_txs, args.dry_run)


if __name__ == "__main__":
    main()
