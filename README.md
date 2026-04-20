<div align="center">

# ⛓️ VITA Blockchain

<img src="https://img.shields.io/badge/status-testnet%20LIVE-brightgreen-brightgreen?style=for-the-badge" alt="Status"/>
<img src="https://img.shields.io/badge/phases-9%2F9%20complete-success?style=for-the-badge" alt="Phase"/>
<img src="https://img.shields.io/badge/mainnet-August%202026-blue?style=for-the-badge" alt="Mainnet"/>
<img src="https://img.shields.io/badge/tests-97%20passing-success?style=for-the-badge" alt="Tests"/>

**Next-Generation Blockchain Payment Infrastructure**

*Powering VITACOIN cryptocurrency and VITAPAY payment services*

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21.13-00ADD8.svg?logo=go)](https://golang.org)
[![Cosmos SDK](https://img.shields.io/badge/cosmos--sdk-v0.50.15-5064EA.svg)](https://github.com/cosmos/cosmos-sdk)
[![CometBFT](https://img.shields.io/badge/cometbft-v0.38-red.svg)](https://github.com/cometbft/cometbft)

[🚀 Quick Start](#-quick-start) •
[📖 Documentation](#-documentation) •
[🏗️ Architecture](#️-architecture) •
[💳 VITAPAY](#-vitapay---payment-service) •
[🪙 VITACOIN](#-vitacoin---cryptocurrency) •
[🛣️ Roadmap](#️-roadmap)

</div>

---

## 🌟 What is VITA Blockchain?

**VITA Blockchain** is a production-ready blockchain platform built on Cosmos SDK that powers a complete payment ecosystem. It combines blockchain technology, VITACOIN cryptocurrency, and VITAPAY payment tools to revolutionize global transactions.

### The Three-Layer System

```
┌─────────────────────────────────────────────────────────────┐
│                    VITA BLOCKCHAIN                           │
│                   The Foundation Layer                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ⛓️  LAYER 1: VITA Blockchain Platform                      │
│  ├─ Custom Cosmos SDK v0.50.15 blockchain                   │
│  ├─ CometBFT Proof-of-Stake consensus                      │
│  ├─ 35,929+ LOC, 70%+ test coverage                       │
│  ├─ 5-second block time, instant finality                  │
│  └─ IBC-enabled for cross-chain communication              │
│        │                                                     │
│        ├─► 🪙 LAYER 2: VITACOIN (VITA Token)                │
│        │   ├─ Native cryptocurrency on VITA Blockchain     │
│        │   ├─ 1 Billion VITA total supply                  │
│        │   ├─ 0.1% protocol fee per transaction            │
│        │   ├─ Fee distribution: 40% burn, 40% validators, 20% treasury
│        │   ├─ Merchant tier system with fee discounts      │
│        │   └─ Deflationary tokenomics with burn cap        │
│        │                                                     │
│        └─► 💳 LAYER 3: VITAPAY (Payment Service)           │
│            ├─ Mobile wallet app (React Native / Expo) ✅   │
│            ├─ Merchant payment gateway (Go/Gin) ✅         │
│            ├─ QR-code payment flow (vitapay:// URI) ✅     │
│            ├─ CosmJS transaction signing ✅                 │
│            └─ Cloud Run deployment ready ✅                 │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 📊 Current Status (April 2026)

### Development Progress

```
Phase 1: Foundation Setup          ████████████████████ 100% ✅
Phase 2: Custom Module             ████████████████████ 100% ✅
Phase 3: Fee System & Treasury     ████████████████████ 100% ✅
Phase 4: Staking System            ████████████████████ 100% ✅
Phase 5: Governance                ████████████████████ 100% ✅
Phase 6: IBC Integration           ████████████████████ 100% ✅
Phase 7: VITAPAY Mobile Wallet     ████████████████████ 100% ✅
Phase 8: VITAPAY Payment Gateway   ████████████████████ 100% ✅
Phase 9: Mainnet Prep & Docs       ████████████████████ 100% ✅
```

**ALL CODE PHASES COMPLETE. Public testnet is LIVE with 3 validators.**

### Technical Metrics

| Metric | Value |
|---|---|
| Blockchain LOC | 35,929+ |
| Mobile Wallet LOC | 1,665 |
| Payment Gateway LOC | 1,214 |
| **Total LOC** | **~35,929** |
| Keeper Tests Passing | 97 |
| Test Coverage | 70.1% |
| Keeper Functions | 80+ |
| gRPC Endpoints | 20+ |
| Total Commits | 28+ |

---

## 🏗️ Architecture

```
vitacoin/ (monorepo)
├── vitacoin/                    ← Blockchain (Go + Cosmos SDK v0.50.15)
│   ├── x/vitacoin/              ← Custom module
│   │   ├── keeper/              ← 80+ keeper functions
│   │   │   ├── keeper.go        ← Core + merchant registry
│   │   │   ├── fee.go           ← Fee collection, treasury, burn
│   │   │   ├── validator.go     ← Validator registry + slash/jail
│   │   │   ├── rewards.go       ← Staking rewards + distribution
│   │   │   ├── liquid_staking.go← stVITA liquid staking derivative
│   │   │   ├── governance.go    ← Proposals, voting, tally, execution
│   │   │   └── ibc.go           ← Cross-chain VITA transfers
│   │   ├── types/               ← All types, params, messages
│   │   └── module.go            ← Module wiring + EndBlocker
│   ├── app/                     ← App wiring (Cosmos SDK v0.50.15)
│   └── cmd/vitacoind/           ← Node daemon CLI
├── vitapay-mobile/              ← React Native / Expo mobile wallet
│   └── src/
│       ├── screens/             ← Home, Send, Receive, Stake, Pay, History
│       ├── lib/                 ← CosmJS wallet, tx signing
│       └── constants/           ← Chain config, endpoints
└── vitapay-gateway/             ← Go/Gin payment gateway
    ├── handlers/                ← Merchant register, payment lifecycle
    ├── blockchain/              ← On-chain tx verification client
    ├── middleware/              ← JWT auth, CORS
    └── Dockerfile               ← Cloud Run ready
```

---

## 🪙 VITACOIN — Cryptocurrency

| Property | Value |
|---|---|
| Ticker | VITA |
| Base Denom | uvita (1 VITA = 1,000,000 uvita) |
| Total Supply | 1,000,000,000 VITA |
| Protocol Fee | 0.1% per transaction |
| Fee Split | 40% burn / 40% validators / 20% treasury |
| Consensus | CometBFT Proof-of-Stake |
| Block Time | ~5 seconds |
| Unbonding Period | 21 days |
| Max Validators | 100 |
| Min Validator Bond | 10,000 VITA |
| Staking APR | 10% |
| Liquid Staking | stVITA (exchange rate: totalDelegated/stVITASupply) |
| IBC Port | vitacoin |
| IBC Version | vitacoin-1 |

### Tokenomics

| Allocation | % | Amount |
|---|---|---|
| Community & Ecosystem | 40% | 400M VITA |
| Team & Advisors (4yr vesting) | 20% | 200M VITA |
| Treasury | 15% | 150M VITA |
| Public Sale | 15% | 150M VITA |
| Validators & Staking | 10% | 100M VITA |

---

## 💳 VITAPAY — Payment Service

### Mobile Wallet (React Native / Expo)
- HD wallet: generate / import via 24-word mnemonic
- Secure key storage: iOS Keychain / Android Keystore (`expo-secure-store`)
- Send VITA, view balances, transaction history
- QR code scanner for merchant payments
- QR code generator for receiving payments
- Staking: delegate, claim rewards, validator list
- Payment URI format: `vitapay://pay?to={addr}&amount={amt}&denom={denom}&memo={id}&expires={ts}`

### Payment Gateway (Go/Gin)
- `POST /api/v1/merchant/register` — register merchant on-chain
- `POST /api/v1/payment/create` — create payment request
- `GET  /api/v1/payment/:id` — poll payment status
- `POST /api/v1/payment/:id/confirm` — verify on-chain tx
- Webhook system: fire-and-retry (3 attempts, exponential backoff)
- JWT auth middleware
- Dockerfile + Cloud Run ready

---

## 🏛️ Governance

- **Proposal types:** text, param_change, treasury_spend
- **Quorum:** 33.4% of staked VITA
- **Threshold:** 50% Yes (of non-abstain votes)
- **Veto threshold:** 33.4%
- **Deposit period:** 2 weeks
- **Voting period:** 1 week
- Executed on-chain via `EndBlockerGovernance`

---

## 🌐 IBC Integration

- Port ID: `vitacoin`
- Version: `vitacoin-1`
- Transfer: escrow (source) / mint (destination)
- Full packet lifecycle: send → receive → ack / timeout
- `IBCModule` interface implemented in `module.go`

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- `make`

### Build

```bash
git clone https://github.com/vishwas-io/VITACOIN
cd VITACOIN/vitacoin

export PATH=$PATH:/usr/local/go/bin
make build
# Binary: ./build/vitacoind
```

### Run Tests

```bash
cd vitacoin
go test -timeout 120s ./x/vitacoin/keeper/
# Expected: 97 tests, PASS, coverage: 70.1%
```

### Initialize a Local Testnet Node

```bash
# Init chain
./build/vitacoind init my-node --chain-id vitacoin-testnet-2

# Add test key
./build/vitacoind keys add validator --keyring-backend test

# Add genesis account
./build/vitacoind genesis add-genesis-account $(./build/vitacoind keys show validator -a --keyring-backend test) 1000000000uvita,1000000stake

# Create genesis validator
./build/vitacoind genesis gentx validator 1000000stake --chain-id vitacoin-testnet-2 --keyring-backend test
./build/vitacoind genesis collect-gentxs

# Start node
./build/vitacoind start
```

### Connect to Public Testnet (Live April 15, 2026)

```
Chain ID:  vitacoin-testnet-2
RPC:       https://rpc.vitacoin.network
REST:      https://api.vitacoin.network
Explorer:  https://explorer.vitacoin.network
Faucet:    https://faucet.vitacoin.network

### One-Line Setup (Linux)

```bash
curl -s https://vitacoin.network/setup.sh | bash
```

Downloads binary, genesis, configures peers, creates systemd service. No Go required.

### Validator Guide

See [docs/VALIDATOR_GUIDE.md](docs/VALIDATOR_GUIDE.md) for full validator setup instructions.

### Discord

Join the community: [discord.gg/9JsRPwDzg](https://discord.gg/9JsRPwDzg)
```

---

## 🛣️ Roadmap

| Phase | Status | Date |
|---|---|---|
| Phase 1: Foundation | ✅ Complete | Oct 2025 |
| Phase 2: Custom Module | ✅ Complete | Oct 2025 |
| Phase 3: Fee System & Treasury | ✅ Complete | Apr 5, 2026 |
| Phase 4: Staking System | ✅ Complete | Apr 6, 2026 |
| Phase 5: Governance | ✅ Complete | Apr 6, 2026 |
| Phase 6: IBC Integration | ✅ Complete | Apr 7, 2026 |
| Phase 7: VITAPAY Mobile Wallet | ✅ Complete | Apr 7, 2026 |
| Phase 8: VITAPAY Payment Gateway | ✅ Complete | Apr 7, 2026 |
| Phase 9: Mainnet Prep Docs | ✅ Complete | Apr 7, 2026 |
| **Public Testnet Launch** | ✅ **Live — April 15, 2026** | |
| External Security Audit | 📋 Planned | May–Jun 2026 |
| Exchange Listings (Osmosis, Gate.io) | 📋 Planned | Jun–Jul 2026 |
| VITAPAY Beta (100 merchants) | 📋 Planned | Jul 2026 |
| **Mainnet Launch** | 🎯 **August 2026** | |

---

## 📖 Documentation

| Doc | Description |
|---|---|
| [docs/mainnet-launch.md](vitacoin/docs/mainnet-launch.md) | Mainnet launch guide |
| [docs/tokenomics.md](vitacoin/docs/tokenomics.md) | Full tokenomics spec |
| [docs/exchange-listing.md](vitacoin/docs/exchange-listing.md) | Exchange listing checklist |
| [scripts/init-node.sh](vitacoin/scripts/init-node.sh) | Node initialization script |
| [scripts/genesis-validate.sh](vitacoin/scripts/genesis-validate.sh) | Genesis validation script |

---

## 🔒 Security

- All keeper state uses typed KV store keys (no raw string keys)
- Fee validation: amounts checked before any state mutation
- Governance: quorum + veto thresholds enforced in tally
- IBC: escrow account verified before packet send
- Mobile wallet: keys stored in OS secure enclave only (never in app state)
- Gateway: no secrets hardcoded — all config via environment variables

> ⚠️ **This repo is PUBLIC.** Never commit private keys, validator keys, mnemonics, or `.env` files.

Pre-mainnet external audit planned with Halborn / Trail of Bits / OtterSec.

---

## 🤝 Contributing

VitaCoin is currently in pre-mainnet development. Testnet is live! Join as a validator:

- **Validator interest:** Open an issue with your organization name and infrastructure details
- **Bug reports:** Open a GitHub issue
- **Security vulnerabilities:** Contact privately before disclosure

---

## 📄 License

Apache 2.0 — see [LICENSE](LICENSE)

---

<div align="center">

**Built with ⚡ by Vishwas Verma**

[vitacoin.network](https://vitacoin.network) • [Discord](https://discord.gg/9JsRPwDzg) • [GitHub](https://github.com/vishwas-io/VITACOIN)

</div>
