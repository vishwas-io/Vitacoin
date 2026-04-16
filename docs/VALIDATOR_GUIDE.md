# VitaCoin Testnet — Validator Onboarding Guide

**Chain ID:** `vitacoin-testnet-1`  
**Bond denom:** `uvita` (1 VITA = 1,000,000 uvita)  
**Minimum self-delegation:** 10,000 VITA (10,000,000,000 uvita)

---

## 1. Prerequisites

### Hardware (minimum)
| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 4 cores | 8 cores |
| RAM | 8 GB | 16 GB |
| Disk | 100 GB SSD | 500 GB NVMe |
| Network | 100 Mbps | 1 Gbps |

### OS
Ubuntu 22.04 LTS (or any systemd-based Linux distro)

### Go 1.24
```bash
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc
go version  # should print go1.24.x
```

### Other dependencies
```bash
sudo apt-get update && sudo apt-get install -y git make build-essential curl jq
```

---

## 2. Build from Source

```bash
git clone https://github.com/esspron/VITACOIN.git
cd VITACOIN/vitacoin

CGO_ENABLED=0 go build \
  -mod=readonly \
  -tags "netgo" \
  -ldflags '-extldflags "-static"' \
  -o build/vitacoind \
  ./cmd/vitacoind

sudo mv build/vitacoind /usr/local/bin/vitacoind
vitacoind version
```

---

## 3. Initialize Node

```bash
# Replace <moniker> with your validator name
vitacoind init <moniker> --chain-id vitacoin-testnet-1

# Example:
vitacoind init my-validator --chain-id vitacoin-testnet-1
```

This creates `~/.vitacoind/` with `config/` and `data/` directories.

---

## 4. Download Genesis

```bash
curl -s http://rpc.vitacoin.network/genesis \
  | jq '.result.genesis' \
  > ~/.vitacoind/config/genesis.json

# Verify chain ID
jq '.chain_id' ~/.vitacoind/config/genesis.json
# Expected: "vitacoin-testnet-1"
```

---

## 5. Configure Node

### Persistent peers
```bash
PEERS="1c8b402728761d888c0ecf28c84edeee597c65eb@34.93.188.116:26656,3cc90df00febacc76146188e25c48bed892ea9ce@34.93.51.248:26656,3a14a2a189d15d3c9929b1111e0b8b10225af361@34.93.37.70:26656"

sed -i "s|^persistent_peers *=.*|persistent_peers = \"$PEERS\"|" \
  ~/.vitacoind/config/config.toml
```

### Minimum gas prices
```bash
sed -i 's|^minimum-gas-prices *=.*|minimum-gas-prices = "0.025uvita"|' \
  ~/.vitacoind/config/app.toml
```

### (Optional) Enable Prometheus metrics
```bash
sed -i 's|^prometheus *=.*|prometheus = true|' \
  ~/.vitacoind/config/config.toml
```

### Verify config
```bash
grep "persistent_peers" ~/.vitacoind/config/config.toml
grep "minimum-gas-prices" ~/.vitacoind/config/app.toml
```

---

## 6. Start Node and Sync

### Run directly (for testing)
```bash
vitacoind start
```

### Run as systemd service (recommended for production)

Create `/etc/systemd/system/vitacoind.service`:

```ini
[Unit]
Description=VitaCoin Node
After=network-online.target
Wants=network-online.target

[Service]
User=<your-user>
ExecStart=/usr/local/bin/vitacoind start
Restart=on-failure
RestartSec=5s
LimitNOFILE=65535
StandardOutput=journal
StandardError=journal
SyslogIdentifier=vitacoind

[Install]
WantedBy=multi-user.target
```

Replace `<your-user>` with the Linux user running the node.

```bash
sudo systemctl daemon-reload
sudo systemctl enable vitacoind
sudo systemctl start vitacoind

# Check status
sudo systemctl status vitacoind

# Follow logs
journalctl -u vitacoind -f
```

### Check sync status
```bash
# Wait until catching_up = false
curl -s http://localhost:26657/status | jq '.result.sync_info.catching_up'

# Current block height
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'
```

Do not create the validator until `catching_up` returns `false`.

---

## 7. Create Wallet and Get Testnet VITA

### Create wallet
```bash
vitacoind keys add <wallet-name>
# Save the mnemonic securely — this is shown only once
```

Note your address:
```bash
vitacoind keys show <wallet-name> -a
# Output: cosmos1...
```

### Request testnet VITA from faucet
```bash
ADDR=$(vitacoind keys show <wallet-name> -a)

curl -X POST http://34.93.188.116:8889/faucet \
  -H "Content-Type: application/json" \
  -d "{\"address\": \"$ADDR\"}"
```

### Verify balance
```bash
vitacoind query bank balances $ADDR --node http://rpc.vitacoin.network:26657
```

You need at least **10,000,000,000 uvita** (10,000 VITA) to create a validator.

---

## 8. Create Validator

Wait until node is fully synced (`catching_up: false`), then:

```bash
ADDR=$(vitacoind keys show <wallet-name> -a)
PUBKEY=$(vitacoind tendermint show-validator)

vitacoind tx staking create-validator \
  --amount 10000000000uvita \
  --from <wallet-name> \
  --pubkey "$PUBKEY" \
  --moniker "<your-moniker>" \
  --chain-id vitacoin-testnet-1 \
  --commission-rate 0.05 \
  --commission-max-rate 0.20 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1 \
  --gas auto \
  --gas-adjustment 1.4 \
  --fees 5000uvita \
  --node http://rpc.vitacoin.network:26657 \
  --broadcast-mode sync \
  -y
```

### Get your validator address
```bash
vitacoind keys show <wallet-name> --bech val -a
# Output: cosmosvaloper1...
```

---

## 9. Verify on Explorer

Open: [http://explorer.vitacoin.network](http://explorer.vitacoin.network)

- Navigate to **Validators**
- Search for your moniker or `cosmosvaloper1...` address
- Confirm status is `BOND_STATUS_BONDED`

You can also query via CLI:
```bash
vitacoind query staking validator <cosmosvaloper1...> \
  --node http://rpc.vitacoin.network:26657
```

Or check all validators:
```bash
vitacoind query staking validators \
  --node http://rpc.vitacoin.network:26657 \
  | jq '.validators[] | {moniker: .description.moniker, status: .status, tokens: .tokens}'
```

---

## 10. Troubleshooting

### Node not connecting to peers
```bash
# Check if p2p port is open
ss -tlnp | grep 26656

# Check firewall
sudo ufw allow 26656/tcp
sudo ufw allow 26657/tcp

# Verify peers are set
grep persistent_peers ~/.vitacoind/config/config.toml
```

### Wrong genesis file
```bash
# Re-download genesis
curl -s http://rpc.vitacoin.network/genesis \
  | jq '.result.genesis' \
  > ~/.vitacoind/config/genesis.json

# Check hash matches network
curl -s http://rpc.vitacoin.network/genesis | jq '.result.genesis' | sha256sum
```

### Validator not in active set
The testnet may have a limited validator set size. Check:
```bash
vitacoind query staking params --node http://rpc.vitacoin.network:26657 \
  | jq '.max_validators'
```

If the set is full, you'll need to stake more than the lowest-bonded validator.

### `out of gas` error on create-validator
Increase gas adjustment:
```bash
--gas-adjustment 1.6 --gas auto
```

### Node crashes with `panic: failed to load latest version`
Data corruption — reset and resync:
```bash
vitacoind tendermint unsafe-reset-all
sudo systemctl start vitacoind
```

### Check node logs for errors
```bash
journalctl -u vitacoind -n 100 --no-pager | grep -E "ERR|PANIC|Error"
```

### Faucet not responding
The faucet is rate-limited. Wait ~60 seconds and retry. Each request drips a fixed amount; you may need multiple requests to reach 10,000 VITA.

---

## Useful Endpoints

| Service | URL |
|---------|-----|
| RPC | http://rpc.vitacoin.network:26657 |
| REST API | http://api.vitacoin.network:1317 |
| Explorer | http://explorer.vitacoin.network |
| Faucet | http://34.93.188.116:8889 |

### Swagger / API docs
```
http://api.vitacoin.network:1317/swagger/
```

---

## Key Files

| File | Purpose |
|------|---------|
| `~/.vitacoind/config/config.toml` | P2P, RPC, consensus config |
| `~/.vitacoind/config/app.toml` | App-level config (gas, pruning, API) |
| `~/.vitacoind/config/genesis.json` | Chain genesis |
| `~/.vitacoind/config/priv_validator_key.json` | **Validator signing key — back up securely** |
| `~/.vitacoind/config/node_key.json` | Node identity key |

> ⚠️ Never share or expose `priv_validator_key.json`. Back it up offline before doing anything else.
