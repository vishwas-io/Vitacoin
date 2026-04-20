# VitaCoin ⚡

> A payment-focused Layer 1 blockchain built on Cosmos SDK — fast, fee-efficient, IBC-native.

[![Build](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/vishwas-io/VITACOIN)
[![Tests](https://img.shields.io/badge/tests-97%20passing-brightgreen)](https://github.com/vishwas-io/VITACOIN)
[![Coverage](https://img.shields.io/badge/coverage-70.1%25-green)](https://github.com/vishwas-io/VITACOIN)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue)](https://golang.org)
[![Cosmos SDK](https://img.shields.io/badge/cosmos--sdk-v0.50.15-blueviolet)](https://github.com/cosmos/cosmos-sdk)
[![Testnet](https://img.shields.io/badge/testnet-LIVE-brightgreen)](https://explorer.vitacoin.network)
[![Mainnet](https://img.shields.io/badge/mainnet-August%202026-orange)](https://vitacoin.network)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](../LICENSE)
[![Discord](https://img.shields.io/badge/discord-join-7289da)](https://discord.gg/9JsRPwDzg)

## What is VitaCoin?

VitaCoin is a Cosmos SDK-based blockchain designed for real-world payments. It features:

- **Low fees** — 0.1% per transaction, split 40% burn / 40% validators / 20% treasury
- **IBC-native** — interoperable with 100+ Cosmos chains on day 1
- **Staking & governance** — fully functional validator network with on-chain governance
- **VITAPAY** — a mobile wallet + payment gateway built on top
- **Live testnet** — chain-id `vitacoin-testnet-2`, 3 validators signing

## Quick Start (One-liner)

```bash
curl -s https://vitacoin.network/setup.sh | bash
```

Or download the pre-built binary from [GitHub Releases → v0.1.0-testnet](https://github.com/vishwas-io/VITACOIN/releases/tag/v0.1.0-testnet).

## Testnet Endpoints

| Service | URL |
|---------|-----|
| RPC | `https://rpc.vitacoin.network` |
| REST API | `https://api.vitacoin.network` |
| Explorer | `https://explorer.vitacoin.network` |
| Faucet | `https://faucet.vitacoin.network` |

**Chain ID:** `vitacoin-testnet-2`  
**Address prefix:** `vita1` (bech32)  
**Denom:** `uvita` (1 VITA = 1,000,000 uvita)

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  VitaCoin Chain                  │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │x/vitacoin│  │x/staking │  │x/governance  │  │
│  │(fee/burn)│  │ ✅ Done  │  │  ✅ Done     │  │
│  └──────────┘  └──────────┘  └──────────────┘  │
│                                                  │
│  ┌────────────────────────────────────────────┐ │
│  │              IBC ✅ Done                   │ │
│  └────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘
         ↕                         ↕
┌─────────────────┐    ┌──────────────────────────┐
│  VITAPAY Mobile │    │  VITAPAY Gateway (Go)    │
│  (React Native) │    │  REST/gRPC payment API   │
│   ✅ Complete   │    │     ✅ Complete           │
└─────────────────┘    └──────────────────────────┘
```

## Phase Status

| Phase | Description            | Status        |
|-------|------------------------|---------------|
| 1     | Core chain & scaffold  | ✅ Complete   |
| 2     | Custom module          | ✅ Complete   |
| 3     | Fee system & treasury  | ✅ Complete   |
| 4     | Staking system         | ✅ Complete   |
| 5     | Governance             | ✅ Complete   |
| 6     | IBC integration        | ✅ Complete   |
| 7     | VITAPAY Mobile wallet  | ✅ Complete   |
| 8     | VITAPAY Payment gateway| ✅ Complete   |
| 9     | Mainnet launch         | 🗓 Aug 2026   |

## Project Stats

- **97 tests passing** — 70.1% coverage
- **35,929+ lines of Go code**
- **3 validators** actively signing on testnet
- **Cosmos SDK v0.50.15** + **CometBFT v0.38**

## Build & Run

### Prerequisites
- Go 1.21+
- Make

### Build from source
```bash
cd vitacoin
make build
```

### Run tests
```bash
cd vitacoin
make test
```

### Run local devnet
```bash
cd vitacoin
make init
make start
```

### Connect to testnet
```bash
# Add testnet to config
vitacoind config chain-id vitacoin-testnet-2
vitacoind config node https://rpc.vitacoin.network:443

# Check status
vitacoind status

# Get testnet tokens from faucet
curl -X POST https://faucet.vitacoin.network/claim \
  -H "Content-Type: application/json" \
  -d '{"address": "vita1YOUR_ADDRESS_HERE"}'
```

## Fee Distribution

Every transaction on VitaCoin pays a 0.1% protocol fee, distributed as:

| Recipient  | Share |
|------------|-------|
| 🔥 Burn    | 40%   |
| Validators | 40%   |
| Treasury   | 20%   |

## Documentation

- [Mainnet Launch Guide](docs/mainnet-launch.md)
- [Tokenomics](docs/tokenomics.md)
- [Exchange Listing Checklist](docs/exchange-listing.md)
- [Phase Docs](../docs/phases/README.md)

## Community

- **Website:** [vitacoin.network](https://vitacoin.network)
- **Discord:** [discord.gg/9JsRPwDzg](https://discord.gg/9JsRPwDzg)
- **GitHub:** [github.com/vishwas-io/VITACOIN](https://github.com/vishwas-io/VITACOIN)
- **Explorer:** [explorer.vitacoin.network](https://explorer.vitacoin.network)

## Security

This is a **PUBLIC** repository. Never commit:
- Private keys or mnemonics
- `priv_validator_key.json` / `node_key.json`
- `.env` files or API keys

## License

Apache 2.0 — see LICENSE

---

Built by [Vishwas Verma](https://github.com/vishwas-io) & Nova ⚡
