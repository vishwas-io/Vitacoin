<div align="center">

# ⛓️ VITA Blockchain

<img src="https://img.shields.io/badge/status-in%20development-yellow?style=for-the-badge" alt="Status"/>
<img src="https://img.shields.io/badge/phase-3%20in%20progress-blue?style=for-the-badge" alt="Phase"/>
<img src="https://img.shields.io/badge/progress-75%25-green?style=for-the-badge" alt="Progress"/>
<img src="https://img.shields.io/badge/tests-fixing-orange?style=for-the-badge" alt="Tests"/>

**Next-Generation Blockchain Payment Infrastructure**

*Powering VITACOIN cryptocurrency and VITAPAY payment services*

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21.13-00ADD8.svg?logo=go)](https://golang.org)
[![Cosmos SDK](https://img.shields.io/badge/cosmos--sdk-v0.50.3-5064EA.svg)](https://github.com/cosmos/cosmos-sdk)
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
│  ├─ Custom Cosmos SDK v0.50.3 blockchain                   │
│  ├─ CometBFT Proof-of-Stake consensus                      │
│  ├─ Production-ready: 7,550+ LOC, 38%+ test coverage      │
│  ├─ 5-second block time, instant finality                  │
│  └─ IBC-enabled for cross-chain communication              │
│        │                                                     │
│        ├─► 🪙 LAYER 2: VITACOIN (VITA Token)                │
│        │   ├─ Native cryptocurrency on VITA Blockchain     │
│        │   ├─ 1 Billion VITA total supply                  │
│        │   ├─ 0.1% protocol fee per transaction            │
│        │   ├─ Fee distribution: 50% validators, 25% burn, 25% treasury
│        │   ├─ Merchant tier system with fee discounts      │
│        │   └─ Deflationary tokenomics with burn cap        │
│        │                                                     │
│        └─► 💳 LAYER 3: VITAPAY (Payment Service) - Planned │
│            ├─ Mobile wallet app (React Native)             │
│            ├─ Merchant payment gateway (Go API)            │
│            ├─ Merchant dashboard (Next.js)                 │
│            ├─ E-commerce plugins (Shopify, WooCommerce)    │
│            └─ User-friendly interface for VITACOIN         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 📊 Current Status (April 2026)

### Development Progress

```
Phase 1: Foundation Setup          ████████████████████ 100% ✅
Phase 2: Custom Module             ████████████████████  98% ✅
Phase 3: Fee System & Treasury     ███████████████░░░░░  75% 🚧  ← ACTIVE
Phase 4: Staking System            ░░░░░░░░░░░░░░░░░░░░   0% 📋
Phase 5: Governance                ░░░░░░░░░░░░░░░░░░░░   0% 📋
Phase 6: IBC Integration           ░░░░░░░░░░░░░░░░░░░░   0% 📋
Phase 7: VITAPAY Mobile Wallet     ░░░░░░░░░░░░░░░░░░░░   0% 📋
Phase 8: VITAPAY Gateway           ░░░░░░░░░░░░░░░░░░░░   0% 📋
Phase 9: Mainnet Launch            ░░░░░░░░░░░░░░░░░░░░   0% 🎯
```

### Technical Metrics (April 4, 2026)

| Metric | Value |
|--------|-------|
| **Production Code** | 7,550+ LOC |
| **Test Code** | 1,900+ LOC |
| **Test Coverage** | 38%+ (target: 50%+) |
| **Keeper Functions** | 80+ implemented |
| **gRPC Endpoints** | 20+ query endpoints |
| **Go Version** | 1.21.13 (installed 2026-04-04) |
| **Build Status** | ✅ Compiles clean |
| **Binary Size** | 44.9 MB (vitacoind) |

### Phase 3 Active Work (2026-04-04)

| Fix | Status |
|-----|--------|
| Go 1.21.13 installed & in PATH | ✅ Done |
| Proto fields (params.pb.go) — fields 12-18 marshal/unmarshal | ✅ Done |
| Fee split validation (100 → 1.0 denominator) | ✅ Done |
| TreasurySpendProposal implements govtypes.Content | ✅ Done |
| Treasury proposal handler unblocked | ✅ Done |
| UpdateMerchant min-stake + name validation | ✅ Done |
| RegisterMerchant name length validation | ✅ Done |
| Keeper test MockBankKeeper + MockAccountKeeper | ✅ Done |
| Test suite error string alignment | 🚧 In Progress |
| gRPC pagination nil fix | 📋 Next |
| Rate limiting (stub → real) | 📋 Next |
| `make test` all green ≥50% coverage | 📋 Next |

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21+ (`which go || /usr/local/go/bin/go`)
- Git, Make

### Build VITA Blockchain

```bash
git clone https://github.com/esspron/vitacoin.git
cd vitacoin/vitacoin
export PATH=$PATH:/usr/local/go/bin
make build
./build/vitacoind version
```

### Run Tests

```bash
cd vitacoin
export PATH=$PATH:/usr/local/go/bin
make test                      # all tests
go test ./x/vitacoin/...       # module only (faster)
make test-cover                # coverage report
```

---

## 🏗️ Architecture

### Custom VITACOIN Module

```
x/vitacoin/
├── keeper/                    # State management (3,190+ LOC)
│   ├── keeper.go             # CRUD operations
│   ├── msg_server.go         # Transaction handlers
│   ├── fees.go               # Fee system (370+ LOC)
│   ├── treasury.go           # Treasury management (550+ LOC)
│   ├── treasury_proposals.go # Gov proposal handler ✅ unblocked
│   ├── fee_state.go          # Fee statistics
│   └── grpc_query.go         # gRPC queries (pagination: 📋 fixing)
├── types/                     # Data structures (16,855+ LOC)
│   ├── params.pb.go          # Proto-generated (fields 1-18 ✅)
│   ├── genesis.pb.go         # Payment/Merchant structs
│   ├── treasury_types.go     # TreasurySpendProposal ✅
│   ├── params.go             # DefaultParams + Validate
│   └── msgs.go               # Message definitions
└── module.go                  # Module lifecycle
```

---

## 🪙 VITACOIN - Cryptocurrency

### Token Specifications

| Property | Value |
|----------|-------|
| **Name** | VITACOIN |
| **Symbol** | VITA |
| **Total Supply** | 1,000,000,000 VITA (1 Billion) |
| **Decimals** | 18 |
| **Smallest Unit** | avita (1 VITA = 10¹⁸ avita) |
| **Block Time** | ~6 seconds |
| **Finality** | Instant (BFT consensus) |

### Fee Distribution (On-Chain)

```
VITACOIN Transaction Fee: 0.1%

├─ 50% → Validators (x/distribution)
├─ 25% → Burned Forever (cap: 500M VITA)
└─ 25% → Treasury (governance-controlled)
```

### Merchant Tier System

| Tier | Stake Required | Fee Discount |
|------|---------------|--------------|
| Bronze | 10,000 VITA | 0% |
| Silver | 50,000 VITA | 25% |
| Gold | 100,000 VITA | 50% |
| Platinum | 100,000+ VITA | 75% |

---

## 💳 VITAPAY - Payment Service

> **Status**: Planning Phase — Target Q2 2026

**Components:**
- 📱 Mobile Wallet (React Native) — iOS + Android
- 🏪 Payment Gateway (Go/Gin) — Merchant API
- 📊 Merchant Dashboard (Next.js) — Analytics
- 🔌 E-commerce Plugins — Shopify, WooCommerce

---

## 🛣️ Roadmap

### ✅ Phase 1: Foundation (Complete — Oct 2025)
- Go 1.21 + Cosmos SDK v0.50.3
- Protobuf definitions + CI/CD pipeline
- Build automation (Makefile)

### ✅ Phase 2: Custom Module (98% — Oct 2025)
- Keeper (3,190+ LOC), 8 message types, 10+ gRPC queries
- Validation system (700+ LOC), test suite (1,900+ LOC)
- Module integrated in app.go

### 🚧 Phase 3: Fee System & Treasury (75% — Active)

- [x] Fee collection & escrow (370+ LOC)
- [x] Fee distribution 50/25/25 split
- [x] Burn mechanism with 500M VITA cap
- [x] Treasury system (1,450+ LOC, 9 queries)
- [x] TreasurySpendProposal → govtypes.Content ✅
- [x] Proto params fields 12-18 serialization ✅
- [x] Go 1.21.13 environment fix ✅
- [ ] gRPC pagination nil fix (4 endpoints)
- [ ] Rate limiting (stub → implementation)
- [ ] `make test` all green — coverage ≥50%
- [ ] Genesis & vesting setup

### 📋 Phase 4: Staking System (Target: Q2 2026)
- Advanced validator mechanics
- Liquid staking (stVITA derivative token)
- Delegation + unbonding queue

### 📋 Phase 5: Governance (Target: Q2 2026)
- Proposal system, voting, treasury spending

### 📋 Phase 6: IBC Integration (Target: Q3 2026)
- Cross-chain transfers, relayer infrastructure

### 📋 Phase 7: VITAPAY Mobile Wallet (Q2 2026)
- React Native, HD wallets, QR scanning, biometrics

### 📋 Phase 8: VITAPAY Payment Gateway (Q2 2026)
- Merchant API, webhooks, SDKs (JS/Python/PHP)

### 🎯 Phase 9: Mainnet Launch (August 2026)
- Security audit → testnet → 100+ validators → genesis ceremony

---

## 🔐 Security

**DO NOT** open public issues for vulnerabilities.  
Email: security@vita-blockchain.network (coming soon)

**Features:**
- Multi-layer validation (proto → basic → business logic)
- Bech32 address checksums
- Emergency pause controls (governance)
- Burn cap enforcement (500M VITA max)
- Byzantine Fault Tolerance (33% attacker tolerance)

---

## 📁 Repository Structure

```
vitacoin/  (monorepo)
├── docs/                   # Documentation
├── vitacoin/               # ⛓️ Blockchain (Go + Cosmos SDK)
│   ├── x/vitacoin/        # Custom module
│   ├── app/               # App setup
│   ├── cmd/vitacoind/     # Daemon CLI
│   └── proto/             # Protobuf definitions
└── vitapay/               # 💳 Payment Service (planned)
    ├── mobile-wallet/
    ├── payment-gateway/
    └── merchant-dashboard/
```

---

## 🔧 Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.21.13 |
| Framework | Cosmos SDK v0.50.3 |
| Consensus | CometBFT v0.38 (PoS) |
| APIs | gRPC (9090) + REST (1317) |
| Mobile | React Native (planned) |
| Backend | Go + Gin (planned) |
| Frontend | Next.js 14 (planned) |

---

## 🤝 Contributing

1. Fork → `git checkout -b feature/xyz`
2. Code → `make test` (all green required)
3. Commit → `git commit -m "fix/feat: description"`
4. PR → wait for review

**Guidelines**: Tests required, `make lint` must pass, no secrets committed.

---

## 📜 License

Apache License 2.0 — see [LICENSE](LICENSE)

---

<div align="center">

**Built with ❤️ by the VITA Blockchain Team**

*VITA Blockchain • VITACOIN • VITAPAY*

**Last Updated**: April 4, 2026 | **Phase**: 3 (75% Complete) | **Mainnet Target**: August 2026

[⬆ Back to Top](#️-vita-blockchain)

</div>
