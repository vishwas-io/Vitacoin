# VITACOIN MAINNET EXECUTION PLAYBOOK

> Generated: 2026-04-09 | Chain: vitacoin-1 | SDK: v0.50.15 | Binary: vitacoind
> Repo: github.com/esspron/VITACOIN | Target: August 2026

---

## PHASE 0 — PRE-FLIGHT VALIDATION

### 0.1 — minimum-gas-prices

```bash
# Verify config
grep "minimum-gas-prices" ~/.vitacoin/config/app.toml
```
**Expected:** `minimum-gas-prices = "0.025uvita"`

```bash
# Verify on-chain
curl -s http://localhost:1317/cosmos/base/node/v1beta1/config | jq .minimum_gas_price
```
**Expected:** `"0.025uvita"`

### 0.2 — Denom consistency (uvita only)

```bash
# Must return 0 results
grep -rn '"avita"' vitacoin/x/vitacoin/ --include="*.go" | grep -v _test.go
```
**Expected:** No output (0 lines)

```bash
# Verify bond denom on-chain
curl -s http://localhost:1317/cosmos/staking/v1beta1/params | jq .params.bond_denom
```
**Expected:** `"uvita"`

### 0.3 — Fee split (40% burn / 40% validators / 20% treasury)

```bash
# Verify in params.go
grep -A3 "FeeBurnPercent\|FeeValidatorPercent\|FeeTreasuryPercent" vitacoin/x/vitacoin/types/params.go | grep -v "^--$"
```
**Expected:**
```
FeeBurnPercent:         math.LegacyNewDecWithPrec(40, 2),  // 40% burn
FeeValidatorPercent:    math.LegacyNewDecWithPrec(40, 2),  // 40% to validators
FeeTreasuryPercent:     math.LegacyNewDecWithPrec(20, 2),  // 20% to treasury
```

### 0.4 — Block gas limit

```bash
grep "max_gas" ~/.vitacoin/config/config.toml
```
**Expected:** Contains `max_gas = 100000000` (in [consensus.params.block] section of genesis)

```bash
curl -s http://localhost:26657/consensus_params | jq .result.consensus_params.block.max_gas
```
**Expected:** `"100000000"`

### 0.5 — Min self-delegation

```bash
curl -s http://localhost:1317/cosmos/staking/v1beta1/validators | jq '.validators[0].min_self_delegation'
```
**Expected:** `"10000000000"` (10,000 VITA = 10,000 × 1,000,000 uvita)

### 0.6 — Rate limiting

```bash
curl -s http://localhost:1317/vitacoin/vitacoin/params | jq .params
```
**Expected:** Contains `MinBlocksBetweenTx > 0` (value = 1)

### 0.7 — Inflation range

```bash
curl -s http://localhost:1317/cosmos/mint/v1beta1/params | jq .params
```
**Expected:**
```json
{
  "inflation_max": "0.100000000000000000",
  "inflation_min": "0.080000000000000000",
  "inflation_rate_change": "0.060000000000000000",
  "goal_bonded": "0.670000000000000000"
}
```

---

## PHASE 1 — CODE & CONFIG FIXES

### 1.1 — Fix denom: avita → uvita (21 files)

```bash
cd /home/gcp-instance-20260311/.openclaw/workspace-vitacoin-code
```

#### File: `vitacoin/x/vitacoin/types/keys.go`
```
BEFORE (line 20):
BondDenom = "avita"

AFTER:
BondDenom = "uvita"
```

#### File: `vitacoin/x/vitacoin/keeper/fees.go`
```bash
sed -i 's/"avita"/"uvita"/g' vitacoin/x/vitacoin/keeper/fees.go
```
Lines affected: 66, 102, 151, 261, 281, 294

#### File: `vitacoin/x/vitacoin/keeper/treasury.go`
```bash
sed -i 's/"avita"/"uvita"/g' vitacoin/x/vitacoin/keeper/treasury.go
```
Lines affected: 109, 226, 227, 391

#### File: `vitacoin/x/vitacoin/keeper/treasury_proposals.go`
```bash
sed -i 's/"avita"/"uvita"/g' vitacoin/x/vitacoin/keeper/treasury_proposals.go
```
Lines affected: 106, 157, 158

#### File: `vitacoin/x/vitacoin/keeper/fee_state.go`
```bash
sed -i 's/"avita"/"uvita"/g' vitacoin/x/vitacoin/keeper/fee_state.go
```
Line affected: 299

#### File: `vitacoin/x/vitacoin/keeper/msg_server.go`
```bash
sed -i 's/"avita"/"uvita"/g' vitacoin/x/vitacoin/keeper/msg_server.go
```
Lines affected: 94, 102, 185 (line 452 is commented out — fix anyway)

#### File: `vitacoin/x/vitacoin/types/ibc_types.go`
```bash
sed -i 's/"avita"/"uvita"/g' vitacoin/x/vitacoin/types/ibc_types.go
```
Line affected: 47 (comment only — fix for consistency)

#### Bulk command (all at once):
```bash
find vitacoin/x/vitacoin/ -name "*.go" ! -name "*_test.go" -exec sed -i 's/"avita"/"uvita"/g' {} +
```

#### Verify:
```bash
grep -rn '"avita"' vitacoin/x/vitacoin/ --include="*.go" | grep -v _test.go
# Expected: 0 results
```

### 1.2 — Fix fee split (40/40/20)

#### File: `vitacoin/x/vitacoin/types/params.go`

```
BEFORE (lines 26-28):
FeeBurnPercent:         math.LegacyNewDecWithPrec(25, 2),  // 25% burn
FeeValidatorPercent:    math.LegacyNewDecWithPrec(50, 2),  // 50% to validators
FeeTreasuryPercent:     math.LegacyNewDecWithPrec(25, 2),  // 25% to treasury

AFTER:
FeeBurnPercent:         math.LegacyNewDecWithPrec(40, 2),  // 40% burn
FeeValidatorPercent:    math.LegacyNewDecWithPrec(40, 2),  // 40% to validators
FeeTreasuryPercent:     math.LegacyNewDecWithPrec(20, 2),  // 20% to treasury
```

### 1.3 — Fix params.go denomination scale

The current params use 1e18 scale (avita). Change to 1e6 scale (uvita).

#### File: `vitacoin/x/vitacoin/types/params.go`

```
BEFORE (line 11):
oneVITA := math.NewInt(1000000000000000000) // 1e18 avita = 1 VITA

AFTER:
oneVITA := math.NewInt(1000000) // 1e6 uvita = 1 VITA
```

```
BEFORE (line 14):
MinGasPrice:             math.LegacyNewDecWithPrec(1, 3), // 0.001 avita

AFTER:
MinGasPrice:             math.LegacyNewDecWithPrec(25, 3), // 0.025 uvita
```

```
BEFORE (line 29):
MinProtocolFee:         math.NewInt(1000000000000000),      // 0.001 VITA (1e15 avita)

AFTER:
MinProtocolFee:         math.NewInt(1000),                   // 0.001 VITA (1000 uvita)
```

```
BEFORE (line 30):
MaxProtocolFee:         math.NewInt(100).Mul(oneVITA),      // 100 VITA

AFTER:
MaxProtocolFee:         math.NewInt(100000000),              // 100 VITA (100 * 1e6 uvita)
```

```
BEFORE (line 31):
BurnCapSupply:          math.NewInt(500000000).Mul(oneVITA), // 500M VITA minimum supply

AFTER:
BurnCapSupply:          math.NewInt(500000000000000),        // 500M VITA (500M * 1e6 uvita)
```

### 1.4 — Enable rate limiting

#### File: `vitacoin/x/vitacoin/keeper/msg_server_validation.go`

Verify `GetMinBlocksBetweenTx` default. Set in genesis params:

```json
"vitacoin": {
  "params": {
    "min_blocks_between_tx": 1
  }
}
```

### 1.5 — Build binary

```bash
cd /home/gcp-instance-20260311/.openclaw/workspace-vitacoin-code
export PATH=$PATH:/usr/local/go/bin

# Run tests
go test ./vitacoin/x/vitacoin/keeper/... -count=1 -v 2>&1 | tail -5
# Expected: ok ... (PASS)

# Security scan
git diff HEAD -U0 | grep -E "(eyJ[A-Za-z0-9]{30,}|sk-[A-Za-z0-9]{20,}|[a-f0-9]{64})" | grep "^\+"
# Expected: no output

# Build
go build -o vitacoind -ldflags "-X github.com/cosmos/cosmos-sdk/version.Version=v1.0.0 -X github.com/cosmos/cosmos-sdk/version.Commit=$(git rev-parse HEAD)" ./vitacoin/cmd/vitacoind

# Verify
./vitacoind version
# Expected: v1.0.0

# Tag
git add -A
git commit -m "fix(mainnet): denom uvita, fee split 40/40/20, gas price, rate limit"
git tag -a v1.0.0 -m "VitaCoin v1.0.0 — Mainnet Release"
git push origin main --tags
```

### 1.6 — Config: app.toml

```toml
# ~/.vitacoin/config/app.toml

minimum-gas-prices = "0.025uvita"

pruning = "default"
pruning-keep-recent = "362880"
pruning-interval = "10"

[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"
max-open-connections = 1000

[grpc]
enable = true
address = "0.0.0.0:9090"

[state-sync]
snapshot-interval = 1000
snapshot-keep-recent = 2

[telemetry]
enabled = true
prometheus-retention-time = 600
```

### 1.7 — Config: config.toml

```toml
# ~/.vitacoin/config/config.toml

[p2p]
laddr = "tcp://0.0.0.0:26656"
external_address = "tcp://<PUBLIC_IP>:26656"
max_num_inbound_peers = 100
max_num_outbound_peers = 30
pex = true
seed_mode = false  # true ONLY for seed nodes
addr_book_strict = false  # false for testnet, true for mainnet
seeds = "<seed1_id>@<seed1_ip>:26656,<seed2_id>@<seed2_ip>:26656"

[rpc]
laddr = "tcp://127.0.0.1:26657"  # bind to localhost, expose via nginx
cors_allowed_origins = []  # empty for validators, ["*"] for public RPC

[mempool]
max_txs_bytes = 1073741824
max_tx_bytes = 1048576
size = 5000
cache_size = 10000

[consensus]
timeout_propose = "3s"
timeout_prevote = "1s"
timeout_precommit = "1s"
timeout_commit = "5s"

[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
```

### 1.8 — Genesis parameter overrides (genesis.json)

```bash
# After vitacoind init, patch genesis:
python3 << 'PYEOF'
import json

with open("config/genesis.json") as f:
    g = json.load(f)

# Fix bond denom globally
raw = json.dumps(g).replace('"stake"', '"uvita"')
g = json.loads(raw)

# Consensus block params
g["consensus"]["params"]["block"]["max_gas"] = "100000000"
g["consensus"]["params"]["block"]["max_bytes"] = "22020096"

# Staking
g["app_state"]["staking"]["params"]["unbonding_time"] = "1814400s"
g["app_state"]["staking"]["params"]["max_validators"] = 100
g["app_state"]["staking"]["params"]["min_commission_rate"] = "0.050000000000000000"

# Slashing
g["app_state"]["slashing"]["params"]["signed_blocks_window"] = "10000"
g["app_state"]["slashing"]["params"]["min_signed_per_window"] = "0.050000000000000000"
g["app_state"]["slashing"]["params"]["slash_fraction_double_sign"] = "0.050000000000000000"
g["app_state"]["slashing"]["params"]["slash_fraction_downtime"] = "0.000100000000000000"
g["app_state"]["slashing"]["params"]["downtime_jail_duration"] = "600s"

# Mint (8-10% inflation)
g["app_state"]["mint"]["params"]["inflation_max"] = "0.100000000000000000"
g["app_state"]["mint"]["params"]["inflation_min"] = "0.080000000000000000"
g["app_state"]["mint"]["params"]["inflation_rate_change"] = "0.060000000000000000"
g["app_state"]["mint"]["params"]["goal_bonded"] = "0.670000000000000000"
g["app_state"]["mint"]["params"]["blocks_per_year"] = "6311520"
g["app_state"]["mint"]["minter"]["inflation"] = "0.090000000000000000"

# Governance
g["app_state"]["gov"]["params"]["min_deposit"] = [{"denom": "uvita", "amount": "10000000000"}]
g["app_state"]["gov"]["params"]["max_deposit_period"] = "259200s"
g["app_state"]["gov"]["params"]["voting_period"] = "432000s"
g["app_state"]["gov"]["params"]["quorum"] = "0.334000000000000000"
g["app_state"]["gov"]["params"]["threshold"] = "0.500000000000000000"
g["app_state"]["gov"]["params"]["veto_threshold"] = "0.334000000000000000"
g["app_state"]["gov"]["params"]["expedited_voting_period"] = "86400s"
g["app_state"]["gov"]["params"]["expedited_threshold"] = "0.667000000000000000"
g["app_state"]["gov"]["params"]["expedited_min_deposit"] = [{"denom": "uvita", "amount": "50000000000"}]

# Distribution
g["app_state"]["distribution"]["params"]["community_tax"] = "0.020000000000000000"

# Crisis
g["app_state"]["crisis"]["constant_fee"] = {"denom": "uvita", "amount": "1000000000000"}

with open("config/genesis.json", "w") as f:
    json.dump(g, f, indent=2)

print("Genesis patched successfully")
PYEOF
```

---

## PHASE 2 — INFRASTRUCTURE SETUP

### 2.1 — Node Inventory

| Role | Count | Machine Type | Disk | Notes |
|---|---|---|---|---|
| Validator | 3-5 | e2-standard-4 (4 CPU, 16GB) | 500GB SSD | Behind sentry nodes |
| Seed | 2 | e2-medium (1 CPU, 4GB) | 100GB SSD | `seed_mode = true` |
| Archive/RPC | 1 | e2-standard-4 (4 CPU, 16GB) | 1TB SSD | `pruning = nothing` |
| Sentry | 2 per validator | e2-standard-2 (2 CPU, 8GB) | 200GB SSD | Public-facing |

### 2.2 — Provision VM (repeat per node)

```bash
# Variables — change per node
NODE_NAME="vitacoin-validator-1"  # or seed-1, archive-1, sentry-1
ZONE="asia-south1-c"
MACHINE="e2-standard-4"  # e2-medium for seed
DISK="500"  # 100 for seed, 1000 for archive

gcloud compute instances create $NODE_NAME \
  --zone=$ZONE \
  --machine-type=$MACHINE \
  --boot-disk-size=${DISK}GB \
  --boot-disk-type=pd-ssd \
  --image-family=debian-12 \
  --image-project=debian-cloud \
  --tags=vitacoin-node \
  --metadata=startup-script='#!/bin/bash
    apt-get update && apt-get install -y git make gcc jq curl wget
    # Install Go 1.21
    wget -q https://go.dev/dl/go1.21.13.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.21.13.linux-amd64.tar.gz
    echo "export PATH=\$PATH:/usr/local/go/bin" >> /etc/profile
  '
```

### 2.3 — Firewall Rules

```bash
# Validator nodes (behind sentry — only allow sentry IPs)
gcloud compute firewall-rules create vitacoin-validator-p2p \
  --allow tcp:26656 \
  --source-ranges="<sentry1_ip>/32,<sentry2_ip>/32" \
  --target-tags=vitacoin-validator

# Seed + Sentry nodes (public P2P)
gcloud compute firewall-rules create vitacoin-public-p2p \
  --allow tcp:26656 \
  --source-ranges="0.0.0.0/0" \
  --target-tags=vitacoin-seed,vitacoin-sentry

# RPC/API node (public)
gcloud compute firewall-rules create vitacoin-public-rpc \
  --allow tcp:80,tcp:443,tcp:26657,tcp:1317,tcp:9090 \
  --source-ranges="0.0.0.0/0" \
  --target-tags=vitacoin-rpc

# Prometheus metrics (internal only)
gcloud compute firewall-rules create vitacoin-metrics \
  --allow tcp:26660,tcp:9100 \
  --source-ranges="10.0.0.0/8" \
  --target-tags=vitacoin-node
```

### 2.4 — Install Binary (run on each node)

```bash
export PATH=$PATH:/usr/local/go/bin

git clone https://github.com/esspron/VITACOIN.git
cd VITACOIN
git checkout v1.0.0

go build -o vitacoind ./vitacoin/cmd/vitacoind
sudo mv vitacoind /usr/local/bin/
vitacoind version
# Expected: v1.0.0
```

### 2.5 — Install Cosmovisor (run on each node)

```bash
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest
sudo mv ~/go/bin/cosmovisor /usr/local/bin/

# Setup directory structure
export DAEMON_NAME=vitacoind
export DAEMON_HOME=$HOME/.vitacoin

mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
mkdir -p $DAEMON_HOME/cosmovisor/upgrades

cp /usr/local/bin/vitacoind $DAEMON_HOME/cosmovisor/genesis/bin/vitacoind
```

### 2.6 — systemd Service (Cosmovisor)

```bash
sudo tee /etc/systemd/system/vitacoind.service << 'EOF'
[Unit]
Description=VitaCoin Node (Cosmovisor)
After=network-online.target
Wants=network-online.target

[Service]
User=gcp-instance-20260311
Environment="DAEMON_NAME=vitacoind"
Environment="DAEMON_HOME=/home/gcp-instance-20260311/.vitacoin"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"
Environment="DAEMON_LOG_BUFFER_SIZE=512"
Environment="UNSAFE_SKIP_BACKUP=false"
ExecStart=/usr/local/bin/cosmovisor run start --home /home/gcp-instance-20260311/.vitacoin
Restart=always
RestartSec=5
LimitNOFILE=65535
LimitNPROC=65535
MemoryMax=14G
CPUQuota=350%

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable vitacoind
```

### 2.7 — Seed Node Config Overrides

```bash
# On seed nodes ONLY:
sed -i 's/seed_mode = false/seed_mode = true/' ~/.vitacoin/config/config.toml
sed -i 's/max_num_inbound_peers = 100/max_num_inbound_peers = 500/' ~/.vitacoin/config/config.toml
sed -i 's/max_num_outbound_peers = 30/max_num_outbound_peers = 100/' ~/.vitacoin/config/config.toml

# Disable API/gRPC on seed nodes
sed -i 's/enable = true/enable = false/' ~/.vitacoin/config/app.toml
```

### 2.8 — Archive Node Config Overrides

```bash
# On archive node ONLY:
sed -i 's/pruning = "default"/pruning = "nothing"/' ~/.vitacoin/config/app.toml

# Enable state sync snapshots for other nodes
sed -i 's/snapshot-interval = 0/snapshot-interval = 1000/' ~/.vitacoin/config/app.toml
sed -i 's/snapshot-keep-recent = 2/snapshot-keep-recent = 5/' ~/.vitacoin/config/app.toml

# Public RPC
sed -i 's|laddr = "tcp://127.0.0.1:26657"|laddr = "tcp://0.0.0.0:26657"|' ~/.vitacoin/config/config.toml
sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = ["*"]/' ~/.vitacoin/config/config.toml
```

---

## PHASE 3 — VALIDATOR ONBOARDING (GENESIS)

### 3.1 — Initialize Genesis Coordinator Node

```bash
# Run on genesis coordinator machine (can be archive node)
CHAIN_ID="vitacoin-1"
MONIKER="genesis-coordinator"

vitacoind init $MONIKER --chain-id $CHAIN_ID --home ~/.vitacoin
```

### 3.2 — Patch Genesis

Run the Python script from Phase 1.8 to set all genesis parameters.

### 3.3 — Add Genesis Accounts

```bash
# Team (vesting: 4yr, 1yr cliff)
# Cliff timestamp = genesis_time + 365 days = Unix timestamp
# End timestamp = genesis_time + 4*365 days = Unix timestamp
GENESIS_TIME_UNIX=1788300000  # TBD — set to actual genesis time

CLIFF=$(( GENESIS_TIME_UNIX + 31536000 ))   # +1 year
END=$(( GENESIS_TIME_UNIX + 126144000 ))     # +4 years

vitacoind genesis add-genesis-account cosmos1<team_addr> 150000000000000uvita \
  --vesting-amount 150000000000000uvita \
  --vesting-start-time $GENESIS_TIME_UNIX \
  --vesting-end-time $END \
  --home ~/.vitacoin

# Treasury (module account — added automatically via InitGenesis)
# No manual genesis account needed; treasury module registers itself

# Ecosystem (3yr linear vesting)
ECO_END=$(( GENESIS_TIME_UNIX + 94608000 ))  # +3 years
vitacoind genesis add-genesis-account cosmos1<ecosystem_addr> 150000000000000uvita \
  --vesting-amount 150000000000000uvita \
  --vesting-start-time $GENESIS_TIME_UNIX \
  --vesting-end-time $ECO_END \
  --home ~/.vitacoin

# Liquidity (no vesting)
vitacoind genesis add-genesis-account cosmos1<liquidity_addr> 50000000000000uvita \
  --home ~/.vitacoin

# Validator genesis accounts (one per validator)
vitacoind genesis add-genesis-account cosmos1<validator1_addr> 50000000000000uvita --home ~/.vitacoin
vitacoind genesis add-genesis-account cosmos1<validator2_addr> 50000000000000uvita --home ~/.vitacoin
vitacoind genesis add-genesis-account cosmos1<validator3_addr> 50000000000000uvita --home ~/.vitacoin
# ... repeat for each validator
```

### 3.4 — Validator gentx (each validator runs on their machine)

```bash
MONIKER="my-validator"
CHAIN_ID="vitacoin-1"
BOND_AMOUNT="10000000000uvita"  # 10,000 VITA minimum

# Generate key (SAVE MNEMONIC)
vitacoind keys add validator --keyring-backend file --home ~/.vitacoin

# Create gentx
vitacoind genesis gentx validator $BOND_AMOUNT \
  --chain-id $CHAIN_ID \
  --moniker "$MONIKER" \
  --commission-rate 0.10 \
  --commission-max-rate 0.20 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation "10000000000" \
  --website "https://example.com" \
  --details "Description of validator" \
  --keyring-backend file \
  --home ~/.vitacoin

# Output file:
ls ~/.vitacoin/config/gentx/
# Expected: gentx-<node_id>.json

# Validator submits this file to genesis coordinator (via PR or secure transfer)
```

### 3.5 — Collect gentx (genesis coordinator)

```bash
# Copy all gentx files to ~/.vitacoin/config/gentx/
# Then:

vitacoind genesis collect-gentxs --home ~/.vitacoin
# Expected: "Genesis transaction written to..."

vitacoind genesis validate-genesis --home ~/.vitacoin
# Expected: "File at /.../.vitacoin/config/genesis.json is a valid genesis file"
```

### 3.6 — Distribute Final Genesis

```bash
# Compute checksum
sha256sum ~/.vitacoin/config/genesis.json
# Expected: <hash>  genesis.json

# Publish
cp ~/.vitacoin/config/genesis.json /path/to/repo/genesis/mainnet/genesis.json
git add genesis/mainnet/genesis.json
git commit -m "genesis: vitacoin-1 mainnet genesis"
git push origin main

# All validators download:
curl -o ~/.vitacoin/config/genesis.json \
  https://raw.githubusercontent.com/esspron/VITACOIN/main/genesis/mainnet/genesis.json

# Verify checksum
sha256sum ~/.vitacoin/config/genesis.json
# Must match published hash
```

### 3.7 — Configure Peers (all nodes)

```bash
# Get node IDs (each validator runs):
vitacoind cometbft show-node-id --home ~/.vitacoin
# Example output: a1b2c3d4e5f6...

# Set seeds (point to seed nodes)
SEEDS="<seed1_id>@<seed1_ip>:26656,<seed2_id>@<seed2_ip>:26656"
sed -i "s/seeds = \"\"/seeds = \"$SEEDS\"/" ~/.vitacoin/config/config.toml

# Set persistent peers (validators set sentry nodes)
PEERS="<sentry1_id>@<sentry1_ip>:26656,<sentry2_id>@<sentry2_ip>:26656"
sed -i "s/persistent_peers = \"\"/persistent_peers = \"$PEERS\"/" ~/.vitacoin/config/config.toml

# Set external address
sed -i "s/external_address = \"\"/external_address = \"tcp:\/\/<PUBLIC_IP>:26656\"/" ~/.vitacoin/config/config.toml
```

---

## PHASE 4 — TESTNET FINAL STABILIZATION (2-3 WEEKS)

### 4.1 — Load Test: Transaction Flood

```bash
# Script: send 100 txs/second for 10 minutes
# Run from a test machine with funded accounts

cat > loadtest.sh << 'LOADEOF'
#!/bin/bash
RPC="http://rpc.vitacoin.network:26657"
HOME_DIR="$HOME/.vitacoin-loadtest"
CHAIN_ID="vitacoin-testnet-1"

# Create 10 funded accounts
for i in $(seq 1 10); do
  vitacoind keys add "loadtest-$i" --keyring-backend test --home $HOME_DIR 2>/dev/null
done

# Fund them (run from validator)
VALIDATOR_ADDR=$(vitacoind keys show validator -a --keyring-backend test --home ~/.vitacoin-testnet)
for i in $(seq 1 10); do
  ADDR=$(vitacoind keys show "loadtest-$i" -a --keyring-backend test --home $HOME_DIR)
  vitacoind tx bank send $VALIDATOR_ADDR $ADDR 1000000000uvita \
    --chain-id $CHAIN_ID --keyring-backend test \
    --home ~/.vitacoin-testnet --yes --fees 5000uvita \
    --sequence $((i-1)) 2>/dev/null &
done
wait
sleep 10

# Flood: 100 tx/s for 600 seconds
SENT=0
START=$(date +%s)
while [ $(($(date +%s) - START)) -lt 600 ]; do
  for i in $(seq 1 10); do
    for j in $(seq 1 10); do
      ADDR_FROM=$(vitacoind keys show "loadtest-$i" -a --keyring-backend test --home $HOME_DIR)
      vitacoind tx bank send $ADDR_FROM $VALIDATOR_ADDR 1uvita \
        --chain-id $CHAIN_ID --keyring-backend test \
        --home $HOME_DIR --yes --fees 5000uvita \
        --node $RPC --broadcast-mode async 2>/dev/null &
      SENT=$((SENT+1))
    done
  done
  sleep 1
  echo "Sent: $SENT txs ($(( $(date +%s) - START ))s elapsed)"
done

echo "Total sent: $SENT"
LOADEOF
chmod +x loadtest.sh
```

**Pass condition:** Chain continues producing blocks. Block time stays <10s. No panics in logs.

```bash
# Monitor during test
watch -n 5 'curl -s http://rpc.vitacoin.network/status | jq .result.sync_info.latest_block_height'

# Check for errors
ssh vitacoin-testnet "sudo journalctl -u vitacoind --since '10 min ago' | grep -ciE 'panic|fatal'"
# Expected: 0
```

### 4.2 — Governance Test

```bash
NODE="http://rpc.vitacoin.network:26657"
HOME="$HOME/.vitacoin-testnet"
CHAIN="vitacoin-testnet-1"

# Submit text proposal
vitacoind tx gov submit-proposal \
  --type text \
  --title "Test Proposal: Update Min Commission" \
  --description "This is a test governance proposal to verify the full lifecycle." \
  --deposit 10000000uvita \
  --from validator \
  --chain-id $CHAIN \
  --keyring-backend test \
  --home $HOME \
  --node $NODE \
  --yes --fees 50000uvita

sleep 10

# Query proposal
curl -s http://api.vitacoin.network/cosmos/gov/v1/proposals | jq '.proposals[-1].id'
# Expected: "1" (or next sequential ID)
PROP_ID=1

# Vote YES
vitacoind tx gov vote $PROP_ID yes \
  --from validator \
  --chain-id $CHAIN \
  --keyring-backend test \
  --home $HOME \
  --node $NODE \
  --yes --fees 50000uvita

sleep 10

# Verify vote recorded
curl -s "http://api.vitacoin.network/cosmos/gov/v1/proposals/$PROP_ID/votes" | jq .
# Expected: vote with option VOTE_OPTION_YES

# Wait for voting period (testnet: reduce to 600s for testing)
# Then check final status
curl -s "http://api.vitacoin.network/cosmos/gov/v1/proposals/$PROP_ID" | jq .proposal.status
# Expected: "PROPOSAL_STATUS_PASSED"
```

**Pass condition:** Proposal transitions: DEPOSIT → VOTING → PASSED.

### 4.3 — Validator Failure Simulation

```bash
# Stop one validator (if multi-validator testnet)
ssh vitacoin-validator-2 "sudo systemctl stop vitacoind"

# Monitor chain continues
for i in $(seq 1 12); do
  HEIGHT=$(curl -s http://rpc.vitacoin.network/status | jq -r .result.sync_info.latest_block_height)
  echo "$(date +%H:%M:%S) Block: $HEIGHT"
  sleep 10
done
# Pass: blocks keep incrementing

# Check slashing after window
curl -s http://api.vitacoin.network/cosmos/slashing/v1beta1/signing_infos | \
  jq '.info[] | select(.missed_blocks_counter > "0")'
# Expected: stopped validator shows missed blocks

# Restart and unjail
ssh vitacoin-validator-2 "sudo systemctl start vitacoind"
sleep 30
ssh vitacoin-validator-2 "vitacoind tx slashing unjail --from validator --chain-id vitacoin-testnet-1 --keyring-backend test --home ~/.vitacoin-testnet --yes --fees 50000uvita"

# Verify unjailed
curl -s http://api.vitacoin.network/cosmos/staking/v1beta1/validators | \
  jq '.validators[] | {moniker: .description.moniker, jailed: .jailed}'
# Expected: all validators jailed=false
```

**Pass condition:** Chain survives validator downtime. Validator successfully unjails.

### 4.4 — Spam Attack Simulation

```bash
# Send 1000 zero-fee txs (should be rejected if min gas price is set)
for i in $(seq 1 1000); do
  vitacoind tx bank send validator cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a 1uvita \
    --chain-id vitacoin-testnet-1 --keyring-backend test \
    --home ~/.vitacoin-testnet --yes --fees 0uvita \
    --broadcast-mode async 2>&1 | grep -c "insufficient fees" &
done
wait

# Check: all should fail with insufficient fees
# Pass: grep shows "insufficient fees" for every tx
```

**Pass condition:** All zero-fee transactions rejected. Chain unaffected.

---

## PHASE 5 — MAINNET PREPARATION

### 5.1 — Freeze Binary

```bash
# Final build with version tag
cd /home/gcp-instance-20260311/.openclaw/workspace-vitacoin-code

git tag -a v1.0.0 -m "VitaCoin v1.0.0 — Mainnet Release"
git push origin v1.0.0

# Build reproducible binary
go build -o vitacoind \
  -ldflags "-X github.com/cosmos/cosmos-sdk/version.Version=v1.0.0 \
            -X github.com/cosmos/cosmos-sdk/version.Commit=$(git rev-parse v1.0.0) \
            -X github.com/cosmos/cosmos-sdk/version.BuildTags=netgo,ledger" \
  ./vitacoin/cmd/vitacoind

# Publish checksum
sha256sum vitacoind > vitacoind-v1.0.0-linux-amd64.sha256
# Publish both binary and checksum to GitHub Releases
```

### 5.2 — Freeze Genesis

```bash
# Final genesis (after all gentx collected)
sha256sum ~/.vitacoin/config/genesis.json > genesis-vitacoin-1.sha256

# Publish to repo + website
echo "Chain ID: vitacoin-1"
echo "Genesis SHA256: $(cat genesis-vitacoin-1.sha256 | cut -d' ' -f1)"
echo "Genesis Time: 2026-08-XX T15:00:00Z"
echo "Binary Version: v1.0.0"
echo "Binary SHA256: $(cat vitacoind-v1.0.0-linux-amd64.sha256 | cut -d' ' -f1)"
```

### 5.3 — Distribute Persistent Peers

```bash
# Collect from all seed + sentry nodes
cat > peers.txt << 'PEERS'
# Seed Nodes
<seed1_id>@<seed1_ip>:26656
<seed2_id>@<seed2_ip>:26656

# Sentry Nodes (public)
<sentry1_id>@<sentry1_ip>:26656
<sentry2_id>@<sentry2_ip>:26656
<sentry3_id>@<sentry3_ip>:26656
<sentry4_id>@<sentry4_ip>:26656
PEERS

# Publish to repo
cp peers.txt genesis/mainnet/peers.txt
git add genesis/mainnet/ && git commit -m "mainnet: peers list" && git push
```

### 5.4 — Validator Pre-Launch Checklist (EACH VALIDATOR)

```bash
# 1. Binary version
vitacoind version
# Expected: v1.0.0

# 2. Genesis hash
sha256sum ~/.vitacoin/config/genesis.json
# Must match published hash

# 3. Seeds configured
grep "^seeds" ~/.vitacoin/config/config.toml
# Must contain seed node addresses

# 4. External address set
grep "external_address" ~/.vitacoin/config/config.toml
# Must contain tcp://<public_ip>:26656

# 5. Cosmovisor installed
cosmovisor version
ls ~/.vitacoin/cosmovisor/genesis/bin/vitacoind
# Must exist

# 6. Systemd service enabled
sudo systemctl is-enabled vitacoind
# Expected: enabled

# 7. Firewall port open
nc -zv <own_public_ip> 26656
# Expected: Connection succeeded

# 8. Disk space
df -h / | awk 'NR==2{print $4}'
# Expected: >400GB available

# 9. Validator key backed up
ls ~/.vitacoin/config/priv_validator_key.json
# Exists AND backed up securely offline

# 10. Keyring accessible
vitacoind keys show validator --keyring-backend file --home ~/.vitacoin
# Must show address
```

---

## PHASE 6 — MAINNET LAUNCH (BLOCK 0)

### 6.1 — T-4 Hours: Start Nodes

```bash
# Order: seed nodes → sentry nodes → validator nodes → archive node

# SEED NODES (start first)
ssh seed-1 "sudo systemctl start vitacoind"
ssh seed-2 "sudo systemctl start vitacoind"
sleep 30

# SENTRY NODES
ssh sentry-1 "sudo systemctl start vitacoind"
ssh sentry-2 "sudo systemctl start vitacoind"
ssh sentry-3 "sudo systemctl start vitacoind"
ssh sentry-4 "sudo systemctl start vitacoind"
sleep 30

# VALIDATOR NODES
ssh validator-1 "sudo systemctl start vitacoind"
ssh validator-2 "sudo systemctl start vitacoind"
ssh validator-3 "sudo systemctl start vitacoind"
# ... all validators
sleep 30

# ARCHIVE/RPC NODE
ssh archive-1 "sudo systemctl start vitacoind"
```

### 6.2 — T-2 Hours: Verify Peering

```bash
# Check each node sees peers
for node in seed-1 seed-2 sentry-1 sentry-2 validator-1 validator-2 validator-3 archive-1; do
  PEERS=$(ssh $node "curl -s http://localhost:26657/net_info | jq .result.n_peers")
  echo "$node: $PEERS peers"
done
# All nodes must show >= 1 peer
```

### 6.3 — T-0: Genesis Time

```bash
# Nodes are waiting for genesis time. At genesis_time:
# CometBFT automatically starts consensus

# Monitor block 1 (run on archive node)
watch -n 1 'curl -s http://localhost:26657/status | jq "{height: .result.sync_info.latest_block_height, time: .result.sync_info.latest_block_time}"'

# BLOCK 1 MUST APPEAR WITHIN 60 SECONDS OF GENESIS TIME
# If not → check logs: sudo journalctl -u vitacoind -f
```

### 6.4 — T+2 Minutes: First Block Verification

```bash
# Block height
curl -s http://rpc.vitacoin.network/status | jq .result.sync_info.latest_block_height
# Expected: "1" or higher

# Validator set
curl -s http://rpc.vitacoin.network/validators | jq '.result.validators | length'
# Expected: number of genesis validators

# All validators signing
curl -s http://api.vitacoin.network/cosmos/slashing/v1beta1/signing_infos | \
  jq '[.info[] | select(.missed_blocks_counter == "0")] | length'
# Expected: equals total validators

# Chain ID correct
curl -s http://api.vitacoin.network/cosmos/base/tendermint/v1beta1/node_info | \
  jq .default_node_info.network
# Expected: "vitacoin-1"
```

### 6.5 — T+5 Minutes: Functional Test

```bash
# Send first mainnet transaction
vitacoind tx bank send <team_addr> <test_addr> 1000000uvita \
  --chain-id vitacoin-1 \
  --keyring-backend file \
  --fees 50000uvita \
  --yes

# Verify tx in block
sleep 10
vitacoind q tx <txhash> --home ~/.vitacoin
# Expected: code: 0 (success)
```

### 6.6 — T+5 Minutes: Public Announcement

```bash
# Update website
cd /home/gcp-instance-20260311/.openclaw/workspace-vitacoin/vitacoin-web
# Update status banner to "MAINNET LIVE"
vercel --prod --yes

# Explorer pointing to mainnet
# Update explorer config chain_name = "vitacoin", api = "https://api.vitacoin.network"

# Announce
echo "🚀 VitaCoin mainnet is LIVE! Chain ID: vitacoin-1 | Block 1 produced at $(date -u)"
```

---

## PHASE 7 — POST-LAUNCH MONITORING

### 7.1 — Install Prometheus (monitoring server)

```bash
# On a dedicated monitoring VM or the archive node

# Install Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.51.0/prometheus-2.51.0.linux-amd64.tar.gz
tar xzf prometheus-2.51.0.linux-amd64.tar.gz
sudo mv prometheus-2.51.0.linux-amd64/prometheus /usr/local/bin/

cat > /etc/prometheus/prometheus.yml << 'PROMEOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alerts.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']

scrape_configs:
  - job_name: 'vitacoin-validators'
    static_configs:
      - targets:
        - '<validator1_ip>:26660'
        - '<validator2_ip>:26660'
        - '<validator3_ip>:26660'
        labels:
          role: validator

  - job_name: 'vitacoin-seeds'
    static_configs:
      - targets:
        - '<seed1_ip>:26660'
        - '<seed2_ip>:26660'
        labels:
          role: seed

  - job_name: 'vitacoin-archive'
    static_configs:
      - targets:
        - '<archive_ip>:26660'
        labels:
          role: archive

  - job_name: 'node-exporter'
    static_configs:
      - targets:
        - '<validator1_ip>:9100'
        - '<validator2_ip>:9100'
        - '<validator3_ip>:9100'
        - '<archive_ip>:9100'
PROMEOF
```

### 7.2 — Alert Rules

```bash
cat > /etc/prometheus/alerts.yml << 'ALERTEOF'
groups:
  - name: vitacoin-critical
    rules:
      - alert: ChainHalted
        expr: increase(cometbft_consensus_height[2m]) == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "VitaCoin chain halted — no new blocks in 2 minutes"

      - alert: ValidatorMissedBlocks
        expr: cometbft_consensus_missing_validators > 0
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "{{ $value }} validators missing from consensus"

      - alert: NodeDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Node {{ $labels.instance }} is DOWN"

      - alert: HighMempoolSize
        expr: cometbft_mempool_size > 5000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Mempool has {{ $value }} pending transactions"

      - alert: LowPeerCount
        expr: cometbft_p2p_peers < 3
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Node {{ $labels.instance }} has only {{ $value }} peers"

      - alert: DiskSpaceLow
        expr: node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} < 0.15
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Disk space below 15% on {{ $labels.instance }}"

      - alert: HighMemoryUsage
        expr: node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes < 0.10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Memory below 10% available on {{ $labels.instance }}"

      - alert: ValidatorJailed
        expr: increase(cometbft_consensus_validator_missed_blocks[10m]) > 50
        labels:
          severity: critical
        annotations:
          summary: "Validator approaching jail threshold"
ALERTEOF
```

### 7.3 — Install Grafana

```bash
# Install
sudo apt-get install -y apt-transport-https software-properties-common
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
echo "deb https://packages.grafana.com/oss/deb stable main" | sudo tee /etc/apt/sources.list.d/grafana.list
sudo apt-get update && sudo apt-get install -y grafana

sudo systemctl enable --now grafana-server

# Access: http://<monitoring_ip>:3000 (admin/admin)
# Import dashboard ID: 11036 (Cosmos Validator)
# Add Prometheus data source: http://localhost:9090
```

### 7.4 — Telegram Alert Bot

```bash
# Install Alertmanager
wget https://github.com/prometheus/alertmanager/releases/download/v0.27.0/alertmanager-0.27.0.linux-amd64.tar.gz
tar xzf alertmanager-0.27.0.linux-amd64.tar.gz
sudo mv alertmanager-0.27.0.linux-amd64/alertmanager /usr/local/bin/

cat > /etc/alertmanager/alertmanager.yml << 'AMEOF'
global:
  resolve_timeout: 5m

route:
  receiver: 'telegram'
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h

receivers:
  - name: 'telegram'
    telegram_configs:
      - bot_token: '<TELEGRAM_BOT_TOKEN>'
        chat_id: 850369190
        parse_mode: 'HTML'
        message: |
          🚨 <b>{{ .GroupLabels.alertname }}</b>
          {{ range .Alerts }}
          {{ .Annotations.summary }}
          {{ end }}
AMEOF

sudo systemctl enable --now alertmanager
```

### 7.5 — Health Check Commands (Daily Operations)

```bash
# Quick status
curl -s http://rpc.vitacoin.network/status | jq '{
  height: .result.sync_info.latest_block_height,
  time: .result.sync_info.latest_block_time,
  catching_up: .result.sync_info.catching_up
}'

# Peer count
curl -s http://rpc.vitacoin.network/net_info | jq .result.n_peers

# All validators status
curl -s http://api.vitacoin.network/cosmos/staking/v1beta1/validators | \
  jq '.validators[] | {moniker: .description.moniker, status: .status, jailed: .jailed, tokens: .tokens}'

# Missed blocks
curl -s http://api.vitacoin.network/cosmos/slashing/v1beta1/signing_infos | \
  jq '.info[] | select(.missed_blocks_counter != "0") | {address: .address, missed: .missed_blocks_counter}'

# Supply check
curl -s http://api.vitacoin.network/cosmos/bank/v1beta1/supply | jq .supply

# Community pool
curl -s http://api.vitacoin.network/cosmos/distribution/v1beta1/community_pool | jq .pool

# Active proposals
curl -s http://api.vitacoin.network/cosmos/gov/v1/proposals | jq '.proposals[] | select(.status == "PROPOSAL_STATUS_VOTING_PERIOD")'
```

---

## PHASE 8 — FAILURE & RECOVERY PLAYBOOK

### 8.1 — Chain Halt Recovery

```bash
# STEP 1: Identify scope
# Check which validators are online
for node in validator-1 validator-2 validator-3; do
  HEIGHT=$(ssh $node "curl -s http://localhost:26657/status 2>/dev/null | jq -r .result.sync_info.latest_block_height" 2>/dev/null || echo "UNREACHABLE")
  echo "$node: block $HEIGHT"
done

# STEP 2: Check consensus state
curl -s http://rpc.vitacoin.network/dump_consensus_state | jq .result.round_state.height_vote_set

# STEP 3: Restart failed validators
ssh <failed_validator> "sudo systemctl restart vitacoind"
sleep 15
ssh <failed_validator> "curl -s http://localhost:26657/status | jq .result.sync_info"

# STEP 4: Verify chain resumes
watch -n 2 'curl -s http://rpc.vitacoin.network/status | jq .result.sync_info.latest_block_height'
# Blocks must increment

# STEP 5: If chain does NOT resume (corrupted state)
# On each validator:
sudo systemctl stop vitacoind
vitacoind cometbft unsafe-reset-all --home ~/.vitacoin --keep-addr-book
# Download snapshot from archive node or use state sync
sudo systemctl start vitacoind
```

### 8.2 — Validator Crash Recovery

```bash
# STEP 1: Restart
sudo systemctl restart vitacoind
sleep 15

# STEP 2: Check if syncing
curl -s http://localhost:26657/status | jq .result.sync_info.catching_up
# If true, wait for sync. If false, proceed.

# STEP 3: Check if jailed
VALOPER=$(vitacoind keys show validator --bech val -a --keyring-backend file --home ~/.vitacoin)
curl -s http://localhost:1317/cosmos/staking/v1beta1/validators/$VALOPER | jq .validator.jailed
# If true:

# STEP 4: Unjail (wait for jail duration to pass — 600s)
vitacoind tx slashing unjail \
  --from validator \
  --chain-id vitacoin-1 \
  --keyring-backend file \
  --home ~/.vitacoin \
  --fees 50000uvita \
  --yes

# STEP 5: Verify active
curl -s http://localhost:1317/cosmos/staking/v1beta1/validators/$VALOPER | jq '{status: .validator.status, jailed: .validator.jailed}'
# Expected: status = BOND_STATUS_BONDED, jailed = false
```

### 8.3 — Cosmovisor Upgrade Flow

```bash
# STEP 1: Submit upgrade proposal via governance
vitacoind tx gov submit-proposal software-upgrade "v1.1.0" \
  --title "Upgrade to v1.1.0" \
  --description "Bug fixes and improvements" \
  --upgrade-height <target_height> \
  --upgrade-info '{"binaries":{"linux/amd64":"https://github.com/esspron/VITACOIN/releases/download/v1.1.0/vitacoind-v1.1.0-linux-amd64?checksum=sha256:<hash>"}}' \
  --deposit 10000000000uvita \
  --from validator \
  --chain-id vitacoin-1 \
  --keyring-backend file \
  --fees 50000uvita \
  --yes

# STEP 2: Vote
vitacoind tx gov vote <proposal_id> yes \
  --from validator \
  --chain-id vitacoin-1 \
  --keyring-backend file \
  --fees 50000uvita \
  --yes

# STEP 3: Prepare new binary (all validators, before upgrade height)
mkdir -p ~/.vitacoin/cosmovisor/upgrades/v1.1.0/bin

# Option A: Download pre-built
wget -O ~/.vitacoin/cosmovisor/upgrades/v1.1.0/bin/vitacoind \
  https://github.com/esspron/VITACOIN/releases/download/v1.1.0/vitacoind-v1.1.0-linux-amd64
chmod +x ~/.vitacoin/cosmovisor/upgrades/v1.1.0/bin/vitacoind

# Option B: Build from source
cd VITACOIN && git checkout v1.1.0
go build -o ~/.vitacoin/cosmovisor/upgrades/v1.1.0/bin/vitacoind ./vitacoin/cmd/vitacoind

# STEP 4: At upgrade height
# Cosmovisor detects upgrade plan → stops old binary → backs up data → starts new binary
# Monitor:
sudo journalctl -u vitacoind -f
# Expected: "upgrade needed... applying upgrade v1.1.0... starting new binary"

# STEP 5: Verify upgrade applied
vitacoind version
# Expected: v1.1.0

curl -s http://localhost:1317/cosmos/upgrade/v1beta1/applied_plan/v1.1.0 | jq .height
# Expected: the upgrade height
```

### 8.4 — Emergency Halt (Critical Bug)

```bash
# IMMEDIATE: Stop all validators
for node in validator-1 validator-2 validator-3; do
  ssh $node "sudo systemctl stop vitacoind" &
done
wait

# Announce in war room
echo "🚨 EMERGENCY HALT — Chain stopped at block $(curl -s http://rpc.vitacoin.network/status | jq -r .result.sync_info.latest_block_height)"

# Fix binary
cd VITACOIN && git checkout -b hotfix/critical-bug
# ... apply fix ...
go build -o vitacoind ./vitacoin/cmd/vitacoind
git tag v1.0.1

# Deploy to all validators
for node in validator-1 validator-2 validator-3; do
  scp vitacoind $node:~/.vitacoin/cosmovisor/genesis/bin/vitacoind
done

# Coordinated restart
for node in seed-1 seed-2 sentry-1 sentry-2 validator-1 validator-2 validator-3 archive-1; do
  ssh $node "sudo systemctl start vitacoind" &
done
wait

# Verify chain resumes
sleep 30
curl -s http://rpc.vitacoin.network/status | jq .result.sync_info.latest_block_height
```

---

## FINAL SECTION — GO / NO-GO CRITERIA

### STRICT LAUNCH CONDITIONS

| # | Condition | Verification Command | Required Value |
|---|---|---|---|
| 1 | **Minimum 5 validators live** | `curl -s api/cosmos/staking/v1beta1/validators \| jq '[.validators[] \| select(.status=="BOND_STATUS_BONDED")] \| length'` | `>= 5` |
| 2 | **No critical bugs** | `gosec ./vitacoin/...` | 0 HIGH findings |
| 3 | **Testnet uptime >14 days** | Check first block time vs now | `>= 14 days` |
| 4 | **Zero chain halts in last 7 days** | Monitoring logs | 0 halts |
| 5 | **All tests pass** | `go test ./vitacoin/x/vitacoin/keeper/... -count=1` | `ok` (PASS) |
| 6 | **Load test passed** | Phase 4.1 results | Blocks <10s under 100 tx/s |
| 7 | **Governance test passed** | Phase 4.2 results | Proposal lifecycle complete |
| 8 | **Validator failure test passed** | Phase 4.3 results | Chain survived, unjail worked |
| 9 | **Spam rejection confirmed** | Phase 4.4 results | Zero-fee txs rejected |
| 10 | **Genesis checksum distributed** | All validators confirm SHA256 match | 100% match |
| 11 | **Cosmovisor installed on all validators** | `ssh <node> cosmovisor version` | Returns version |
| 12 | **Monitoring + alerts live** | Trigger test alert | Alert received in Telegram |
| 13 | **No `avita` references in code** | `grep -rn '"avita"' vitacoin/x/ \| grep -v _test.go \| wc -l` | `0` |
| 14 | **Binary version tagged** | `vitacoind version` on all nodes | `v1.0.0` |
| 15 | **Explorer live** | `curl -s explorer.vitacoin.network` | HTTP 200 |

### DECISION MATRIX

```
IF any condition 1-9 FAILS  → DO NOT LAUNCH. Fix and retest.
IF any condition 10-15 FAILS → DELAY 48 HOURS. Fix and reverify.
IF ALL conditions PASS       → PROCEED TO LAUNCH.
```

### SIGN-OFF REQUIRED

```
[ ] Vishwas Verma (Founder)     — Date: ___________
[ ] Nova (CTO/DevOps)           — Date: ___________
[ ] Security Auditor            — Date: ___________
[ ] Lead Validator Operator     — Date: ___________
```

---

**DO NOT LAUNCH IF ANY CONDITION FAILS.**

---

*Document version: 1.0 | Generated: 2026-04-09 | Next review: before mainnet genesis ceremony*
