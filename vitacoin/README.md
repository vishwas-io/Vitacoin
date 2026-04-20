# VitaCoin ⚡

> A payment-focused Layer 1 blockchain built on Cosmos SDK — fast, fee-efficient, IBC-native.

[![Build](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/vishwas-io/VITACOIN)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue)](https://golang.org)
[![Chain](https://img.shields.io/badge/chain-vitacoin--1-purple)](https://vitacoin.network)
[![Mainnet](https://img.shields.io/badge/mainnet-August%202026-orange)](https://vitacoin.network)

## What is VitaCoin?

VitaCoin is a Cosmos SDK-based blockchain designed for real-world payments. It features:

- **Low fees** via an on-chain fee system with treasury management
- **Staking rewards** for validators and delegators
- **IBC-native** — interoperable with 100+ Cosmos chains on day 1
- **VITAPAY** — a mobile wallet + payment gateway built on top

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  VitaCoin Chain                  │
│                                                  │
│  ┌─────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ x/vitacoin│  │ x/staking │  │ x/governance  │  │
│  │ (fee/burn)│  │ (Phase 4) │  │   (Phase 5)   │  │
│  └─────────┘  └──────────┘  └───────────────┘  │
│                                                  │
│  ┌──────────────────────────────────────────┐   │
│  │           IBC (Phase 6)                  │   │
│  └──────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
         ↕                         ↕
┌─────────────────┐    ┌────────────────────────┐
│  VITAPAY Mobile │    │  VITAPAY Gateway (Go)  │
│  (React Native) │    │  REST/gRPC payment API │
│    (Phase 7)    │    │       (Phase 8)        │
└─────────────────┘    └────────────────────────┘
```

## Current Status

| Phase | Description          | Status      |
|-------|----------------------|-------------|
| 1     | Core chain           | ✅ Complete  |
| 2     | Module scaffold      | ✅ Complete  |
| 3     | Fee System/Treasury  | 🔄 75% done  |
| 4     | Staking System       | ⏳ Pending   |
| 5     | Governance           | ⏳ Pending   |
| 6     | IBC Integration      | ⏳ Pending   |
| 7     | VITAPAY Mobile       | ⏳ Pending   |
| 8     | VITAPAY Gateway      | ⏳ Pending   |
| 9     | Mainnet Launch       | 🗓 Aug 2026  |

## Build & Run

### Prerequisites
- Go 1.21+
- Make

### Build
```bash
cd vitacoin
make build
```

### Test
```bash
cd vitacoin
make test
```

### Run local node (devnet)
```bash
cd vitacoin
make init
make start
```

### Initialize a validator node
```bash
bash scripts/init-node.sh my-validator
```

## Documentation

- [Mainnet Launch Guide](docs/mainnet-launch.md)
- [Tokenomics](docs/tokenomics.md)
- [Exchange Listing Checklist](docs/exchange-listing.md)

## Website

**[vitacoin.network](https://vitacoin.network)** — live project status, roadmap, investor info

## Security

This is a PUBLIC repository. Never commit:
- Private keys or mnemonics
- `priv_validator_key.json` / `node_key.json`
- `.env` files or API keys

## License

MIT — see LICENSE

---

Built by Vishwas Verma & Nova ⚡
