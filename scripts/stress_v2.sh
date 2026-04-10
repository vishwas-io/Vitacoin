#!/bin/bash
# VitaCoin Stress Test v2 — Direct CLI broadcast, proper sequence management
# Usage: stress_v2.sh <num_accounts> <txs_per_account> <label>
# Runs on testnet VM only

set -euo pipefail

VBIN="/home/gcp-instance-20260311/vitacoind"
VHOME="/home/gcp-instance-20260311/.vitacoin-testnet"
CHAIN="vitacoin-testnet-1"
DEST="cosmos1rh5vuyeh2dm4afm6hzldrj368s27pq6wvsn6td"  # faucet
REST="http://localhost:1317"

NUM_ACCOUNTS=${1:-3}
TXS_PER_ACCOUNT=${2:-10}
LABEL=${3:-"stress"}
TOTAL=$((NUM_ACCOUNTS * TXS_PER_ACCOUNT))

echo "============================================"
echo "VitaCoin Stress Test v2"
echo "Accounts: $NUM_ACCOUNTS | Txs/acct: $TXS_PER_ACCOUNT | Total: $TOTAL"
echo "Label: $LABEL"
echo "Started: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo "============================================"

# Get block height before
HEIGHT_BEFORE=$(curl -s http://localhost:26657/status | python3 -c "import json,sys; print(json.load(sys.stdin)['result']['sync_info']['latest_block_height'])")
echo "Block before: $HEIGHT_BEFORE"

# Phase 1: Pre-sign all txs
echo ""
echo "=== PHASE 1: Pre-signing $TOTAL txs ==="
SIGN_START=$(date +%s%N)
SIGNED=0
SIGN_ERRORS=0

mkdir -p /tmp/stress_txs

for i in $(seq 0 $((NUM_ACCOUNTS - 1))); do
    ACCT="stress${i}"
    ADDR=$(${VBIN} keys show ${ACCT} -a --keyring-backend=test --home=${VHOME} 2>/dev/null)
    
    # Get account info
    ACCT_INFO=$(curl -s "${REST}/cosmos/auth/v1beta1/accounts/${ADDR}")
    ACCT_NUM=$(echo "$ACCT_INFO" | python3 -c "import json,sys; a=json.load(sys.stdin).get('account',{}); a=a.get('base_account',a); print(a.get('account_number',0))")
    SEQ=$(echo "$ACCT_INFO" | python3 -c "import json,sys; a=json.load(sys.stdin).get('account',{}); a=a.get('base_account',a); print(a.get('sequence',0))")
    
    echo "  ${ACCT} (${ADDR}): acct=${ACCT_NUM} seq=${SEQ}"
    
    for j in $(seq 0 $((TXS_PER_ACCOUNT - 1))); do
        CUR_SEQ=$((SEQ + j))
        AMOUNT=$(( (RANDOM % 100) + 1 ))
        NOTE="${LABEL}-${ACCT}-s${CUR_SEQ}-t$(date +%s%N)"
        TX_FILE="/tmp/stress_txs/${ACCT}_${CUR_SEQ}.json"
        
        # Generate unsigned tx
        ${VBIN} tx send ${ACCT} ${DEST} ${AMOUNT}uvita \
            --chain-id=${CHAIN} \
            --fees=2000uvita \
            --gas=80000 \
            --note="${NOTE}" \
            --sequence=${CUR_SEQ} \
            --account-number=${ACCT_NUM} \
            --generate-only \
            --keyring-backend=test \
            --keyring-dir=${VHOME} \
            --output=json \
            > /tmp/stress_txs/unsigned_${ACCT}_${CUR_SEQ}.json 2>/dev/null
        
        if [ $? -ne 0 ]; then
            SIGN_ERRORS=$((SIGN_ERRORS + 1))
            continue
        fi
        
        # Sign it
        ${VBIN} tx sign /tmp/stress_txs/unsigned_${ACCT}_${CUR_SEQ}.json \
            --from=${ACCT} \
            --chain-id=${CHAIN} \
            --sequence=${CUR_SEQ} \
            --account-number=${ACCT_NUM} \
            --offline \
            --keyring-backend=test \
            --keyring-dir=${VHOME} \
            --output=json \
            > ${TX_FILE} 2>/dev/null
        
        if [ $? -eq 0 ]; then
            SIGNED=$((SIGNED + 1))
        else
            SIGN_ERRORS=$((SIGN_ERRORS + 1))
        fi
    done
done

SIGN_END=$(date +%s%N)
SIGN_MS=$(( (SIGN_END - SIGN_START) / 1000000 ))
echo ""
echo "Pre-sign complete: ${SIGNED} signed, ${SIGN_ERRORS} errors, ${SIGN_MS}ms"

if [ $SIGNED -eq 0 ]; then
    echo "ERROR: No txs signed. Aborting."
    exit 1
fi

# Phase 2: Broadcast all signed txs via REST as fast as possible
echo ""
echo "=== PHASE 2: Broadcasting $SIGNED txs via REST ==="
BROADCAST_START=$(date +%s%N)
SUCCESS=0
FAIL=0
ERRORS=""

for TX_FILE in /tmp/stress_txs/stress*.json; do
    [ -f "$TX_FILE" ] || continue
    
    # Encode tx for REST broadcast
    TX_BYTES=$(python3 -c "
import json, base64, subprocess, sys
tx = json.load(open('${TX_FILE}'))
# Use encode command to get amino bytes
result = subprocess.run(
    ['${VBIN}', 'tx', 'encode', '${TX_FILE}', '--output=json'],
    capture_output=True, text=True
)
if result.returncode == 0:
    print(result.stdout.strip().strip('\"'))
else:
    sys.exit(1)
" 2>/dev/null)
    
    if [ -z "$TX_BYTES" ]; then
        # Fallback: broadcast via CLI
        RESULT=$(${VBIN} tx broadcast ${TX_FILE} --broadcast-mode=sync --output=json --home=${VHOME} 2>/dev/null || echo '{"code":99}')
        CODE=$(echo "$RESULT" | python3 -c "import json,sys; print(json.load(sys.stdin).get('code',99))" 2>/dev/null || echo "99")
    else
        # Broadcast via REST
        RESULT=$(curl -s -X POST "${REST}/cosmos/tx/v1beta1/txs" \
            -H "Content-Type: application/json" \
            -d "{\"tx_bytes\":\"${TX_BYTES}\",\"mode\":\"BROADCAST_MODE_SYNC\"}" 2>/dev/null)
        CODE=$(echo "$RESULT" | python3 -c "import json,sys; r=json.load(sys.stdin).get('tx_response',{}); print(r.get('code',99))" 2>/dev/null || echo "99")
    fi
    
    if [ "$CODE" = "0" ]; then
        SUCCESS=$((SUCCESS + 1))
    else
        FAIL=$((FAIL + 1))
        if [ $FAIL -le 5 ]; then
            ERRORS="${ERRORS}\n  Code ${CODE}: $(basename ${TX_FILE})"
        fi
    fi
done

BROADCAST_END=$(date +%s%N)
BROADCAST_MS=$(( (BROADCAST_END - BROADCAST_START) / 1000000 ))
BROADCAST_SEC=$(python3 -c "print(f'{${BROADCAST_MS}/1000:.2f}')")

echo "Broadcast complete: ${SUCCESS} ok, ${FAIL} failed, ${BROADCAST_MS}ms"
if [ -n "$ERRORS" ]; then
    echo "First errors:${ERRORS}"
fi

# Phase 3: Wait for inclusion and measure
echo ""
echo "=== PHASE 3: Waiting for tx inclusion ==="
sleep 15

HEIGHT_AFTER=$(curl -s http://localhost:26657/status | python3 -c "import json,sys; print(json.load(sys.stdin)['result']['sync_info']['latest_block_height'])")
BLOCKS=$((HEIGHT_AFTER - HEIGHT_BEFORE))

# Count txs in blocks
TOTAL_TXS_IN_BLOCKS=0
for h in $(seq $((HEIGHT_BEFORE + 1)) ${HEIGHT_AFTER}); do
    BLOCK_TXS=$(curl -s "http://localhost:26657/block?height=${h}" | python3 -c "import json,sys; d=json.load(sys.stdin); print(len(d['result']['block']['data'].get('txs',[])))" 2>/dev/null || echo 0)
    TOTAL_TXS_IN_BLOCKS=$((TOTAL_TXS_IN_BLOCKS + BLOCK_TXS))
done

# Calculate TPS
if [ "$BROADCAST_MS" -gt 0 ]; then
    BROADCAST_TPS=$(python3 -c "print(f'{${SUCCESS} / (${BROADCAST_MS}/1000):.1f}')")
else
    BROADCAST_TPS="N/A"
fi

CHAIN_TPS=$(python3 -c "
blocks=${BLOCKS}
txs=${TOTAL_TXS_IN_BLOCKS}
if blocks > 0:
    avg_block_time = 5.0  # ~5s per block
    print(f'{txs / (blocks * avg_block_time):.1f}')
else:
    print('N/A')
")

# Check validator signatures
LATEST=$(curl -s http://localhost:26657/status | python3 -c "import json,sys; print(json.load(sys.stdin)['result']['sync_info']['latest_block_height'])")
SIGS=$(curl -s "http://localhost:26657/commit?height=${LATEST}" | python3 -c "
import json
d = json.load(open('/dev/stdin'))
sigs = d['result']['signed_header']['commit']['signatures']
active = sum(1 for s in sigs if s.get('block_id_flag') == 2)
print(f'{active}/{len(sigs)}')
" 2>/dev/null || echo "?/?")

# Resource usage
RESOURCES=$(ps aux | grep "vitacoind start" | grep -v grep | head -1 | awk '{print "CPU:", $3"%", "MEM:", $4"%", "RSS:", $6/1024, "MB"}')

echo ""
echo "============================================"
echo "STRESS TEST RESULTS: ${LABEL}"
echo "============================================"
echo "Time:           $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo "Accounts:       ${NUM_ACCOUNTS}"
echo "Txs/account:    ${TXS_PER_ACCOUNT}"
echo "Total target:   ${TOTAL}"
echo "---"
echo "Signed:         ${SIGNED}"
echo "Broadcast OK:   ${SUCCESS}"
echo "Broadcast fail: ${FAIL}"
echo "Sign time:      ${SIGN_MS}ms"
echo "Broadcast time: ${BROADCAST_SEC}s"
echo "Broadcast TPS:  ${BROADCAST_TPS}"
echo "---"
echo "Blocks:         ${HEIGHT_BEFORE} → ${HEIGHT_AFTER} (${BLOCKS} blocks)"
echo "Txs in blocks:  ${TOTAL_TXS_IN_BLOCKS}"
echo "Chain TPS:      ${CHAIN_TPS}"
echo "Validators:     ${SIGS} signing"
echo "Resources:      ${RESOURCES}"
echo "============================================"

# Cleanup
rm -rf /tmp/stress_txs

echo "DONE"
