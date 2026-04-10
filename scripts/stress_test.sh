#!/usr/bin/env bash
# stress_test.sh — VitaCoin CLI Stress Test (Quick Bash Alternative)
# Usage: ./scripts/stress_test.sh [--mode light|medium|heavy|burst] [--rpc http://localhost:26657]
#
# Requires: vitacoind in PATH, keyring with a funded 'faucet' account

set -euo pipefail

# ── Defaults ────────────────────────────────────────────────────────────────
MODE="light"
RPC="http://34.93.89.182:26657"
REST="http://34.93.89.182:1317"
CHAIN_ID="vitacoin-testnet-1"
DENOM="uvita"
GAS=200000
FEES="50000uvita"
FAUCET="faucet"
NUM_ACCOUNTS=5

# ── Parse Args ───────────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case $1 in
    --mode)   MODE="$2";  shift 2 ;;
    --rpc)    RPC="$2";   shift 2 ;;
    --rest)   REST="$2";  shift 2 ;;
    *) echo "Unknown arg: $1"; exit 1 ;;
  esac
done

# ── Mode Config ──────────────────────────────────────────────────────────────
case "$MODE" in
  light)  TX_PER_SEC=10;  DURATION=60 ;;
  medium) TX_PER_SEC=50;  DURATION=60 ;;
  heavy)  TX_PER_SEC=100; DURATION=30 ;;
  burst)  TX_PER_SEC=500; DURATION=10 ;;
  *) echo "Unknown mode: $MODE. Use light|medium|heavy|burst"; exit 1 ;;
esac

export PATH=$PATH:/usr/local/go/bin

echo "⚡ VitaCoin Bash Stress Test — Mode: $MODE ($TX_PER_SEC tx/s for ${DURATION}s)"
echo "   RPC: $RPC | Chain: $CHAIN_ID"
echo ""

# ── Create & Fund Test Accounts ──────────────────────────────────────────────
echo "📦 Creating $NUM_ACCOUNTS test accounts..."
ADDRS=()
for i in $(seq 0 $((NUM_ACCOUNTS - 1))); do
  NAME="bstress-acct-$i"
  # Try to add; if exists, show existing
  ADDR=$(vitacoind keys add "$NAME" \
    --keyring-backend test \
    --output json 2>/dev/null | jq -r '.address' \
    || vitacoind keys show "$NAME" \
       --keyring-backend test \
       --output json | jq -r '.address')
  ADDRS+=("$ADDR")
  echo "  ✓ $NAME → $ADDR"
done

echo ""
echo "💸 Funding accounts (10,000,000 uvita each)..."
for i in "${!ADDRS[@]}"; do
  ADDR="${ADDRS[$i]}"
  vitacoind tx bank send "$FAUCET" "$ADDR" "10000000${DENOM}" \
    --chain-id "$CHAIN_ID" \
    --node "$RPC" \
    --gas $GAS \
    --fees "$FEES" \
    --keyring-backend test \
    --yes --broadcast-mode sync \
    --output json > /dev/null 2>&1 && echo "  ✓ funded $ADDR" || echo "  ✗ failed to fund $ADDR"
  sleep 0.5
done

echo ""
echo "⏳ Waiting 5s for funding txs to land..."
sleep 5

# ── Stress Loop ───────────────────────────────────────────────────────────────
TOTAL_TX=$((TX_PER_SEC * DURATION))
SLEEP_MS=$(echo "scale=6; 1 / $TX_PER_SEC" | bc)

SUCCESS=0
FAILED=0
START_TS=$(date +%s%N)

echo "🚀 Sending $TOTAL_TX transactions over ${DURATION}s..."
echo "   (one dot = 10 txs)"
echo ""

COUNTER=0
END_TS=$(( $(date +%s) + DURATION ))
BATCH_PIDS=()

while [[ $(date +%s) -lt $END_TS ]]; do
  # pick random from/to
  FROM_IDX=$((RANDOM % NUM_ACCOUNTS))
  TO_IDX=$(( (FROM_IDX + 1 + RANDOM % (NUM_ACCOUNTS - 1)) % NUM_ACCOUNTS ))
  FROM_NAME="bstress-acct-$FROM_IDX"
  TO_ADDR="${ADDRS[$TO_IDX]}"

  (
    OUT=$(vitacoind tx bank send "$FROM_NAME" "$TO_ADDR" "1000${DENOM}" \
      --chain-id "$CHAIN_ID" \
      --node "$RPC" \
      --gas $GAS \
      --fees "$FEES" \
      --keyring-backend test \
      --yes --broadcast-mode async \
      --output json 2>&1)
    CODE=$(echo "$OUT" | jq -r '.code // 0' 2>/dev/null || echo "1")
    if [[ "$CODE" == "0" ]]; then
      echo "OK"
    else
      echo "FAIL"
    fi
  ) &
  BATCH_PIDS+=($!)

  COUNTER=$((COUNTER + 1))
  if (( COUNTER % 10 == 0 )); then
    printf "."
  fi

  # Rate limiting
  sleep "$SLEEP_MS" 2>/dev/null || true
done

echo ""
echo ""
echo "⏳ Waiting for in-flight transactions to complete..."

# Collect results
for pid in "${BATCH_PIDS[@]}"; do
  RESULT=$(wait "$pid" && echo "OK" || echo "FAIL") 2>/dev/null || RESULT="FAIL"
  if [[ "$RESULT" == "OK" ]]; then
    SUCCESS=$((SUCCESS + 1))
  else
    FAILED=$((FAILED + 1))
  fi
done

END_TS_NS=$(date +%s%N)
ELAPSED_NS=$(( END_TS_NS - START_TS ))
ELAPSED_S=$(echo "scale=2; $ELAPSED_NS / 1000000000" | bc)
ACTUAL_TOTAL=$((SUCCESS + FAILED))
TPS=$(echo "scale=2; $ACTUAL_TOTAL / $ELAPSED_S" | bc 2>/dev/null || echo "N/A")
FAIL_RATE=0
if [[ $ACTUAL_TOTAL -gt 0 ]]; then
  FAIL_RATE=$(echo "scale=1; $FAILED * 100 / $ACTUAL_TOTAL" | bc)
fi

# ── Summary ───────────────────────────────────────────────────────────────────
echo "╔═══════════════════════════════════════════════════╗"
echo "║   ⚡ VitaCoin Bash Stress Test — Results Summary   ║"
echo "╠═══════════════════════════════════════════════════╣"
printf "║  Mode:           %-32s║\n" "$MODE"
printf "║  Target TPS:     %-32d║\n" "$TX_PER_SEC"
printf "║  Elapsed:        %-31s s║\n" "$ELAPSED_S"
echo "╠═══════════════════════════════════════════════════╣"
printf "║  Total Sent:     %-32d║\n" "$ACTUAL_TOTAL"
printf "║  Successful:     %-32d║\n" "$SUCCESS"
printf "║  Failed:         %-32d║\n" "$FAILED"
printf "║  Fail Rate:      %-31s%%║\n" "$FAIL_RATE"
echo "╠═══════════════════════════════════════════════════╣"
printf "║  Actual TPS:     %-32s║\n" "$TPS"
echo "╚═══════════════════════════════════════════════════╝"
