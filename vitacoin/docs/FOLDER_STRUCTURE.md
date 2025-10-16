# 📁 VITACOIN Ecosystem - Folder Structure

This document explains the **monorepo structure** that houses both VITACOIN (cryptocurrency) and VITAPAY (payment network).

## Why a Monorepo?

**Benefits:**
- ✅ Easier development initially (both projects tightly coupled)
- ✅ Shared code and utilities
- ✅ Simplified dependency management
- ✅ Faster iteration and testing
- ✅ Single source of truth

**Future:** Can be split into separate repos once stable.

---

## 📂 Complete Structure

```
vitacoin/  (root repository)
│
├── README.md                     # 🌟 ECOSYSTEM OVERVIEW (start here!)
├── LICENSE                       # Apache 2.0 license
├── .gitignore                    # Git ignore rules
├── .github/                      # GitHub workflows, issue templates
│
├── docs/                         # 📚 ALL DOCUMENTATION
│   ├── README.md                # Documentation index
│   ├── VITACOIN.md              # 🪙 Cryptocurrency guide (blockchain, staking, etc.)
│   ├── FOLDER_STRUCTURE.md      # 📁 This file!
│   │
│   ├── architecture/             # Technical architecture
│   │   ├── ARCHITECTURE.md
│   │   ├── DEV_SETUP.md
│   │   └── SECURITY.md
│   │
│   ├── development/              # Developer guides
│   │   ├── GETTING_STARTED.md
│   │   ├── QUICK_REFERENCE.md
│   │   └── COMPILATION_FIX_SUMMARY.md
│   │
│   └── project/                  # Project management
│       ├── VITAPAY.md           # 💳 Payment network guide
│       ├── MOBILE_APP.md        # Mobile wallet specs
│       ├── PROJECT_SUMMARY.md   # High-level overview
│       ├── DEVELOPMENT_ROADMAP.md
│       ├── TODO.md
│       └── REVENUE_AUTOMATION_STRATEGY.md
│
├── vitacoin/                     # 🪙 THE CRYPTOCURRENCY BLOCKCHAIN
│   ├── README.md                # VITACOIN-specific docs
│   ├── go.mod                   # Go dependencies
│   ├── Makefile                 # Build commands
│   ├── config.yml               # Chain configuration
│   ├── buf.yaml                 # Protobuf config
│   │
│   ├── app/                     # Cosmos SDK app setup
│   │   ├── ante.go             # Transaction preprocessing
│   │   ├── app.go              # Main application
│   │   ├── encoding.go         # Encoding config
│   │   ├── genesis.go          # Genesis state
│   │   └── params.go           # Network parameters
│   │
│   ├── cmd/                     # Command-line interface
│   │   └── vitacoind/
│   │       ├── main.go         # Entry point
│   │       └── cmd/            # CLI commands
│   │
│   ├── proto/                   # Protocol buffer definitions
│   │   └── vitacoin/
│   │       └── v1/
│   │           ├── tx.proto    # Transaction messages
│   │           ├── query.proto # Query messages
│   │           └── genesis.proto
│   │
│   ├── x/                       # Cosmos SDK modules
│   │   └── vitacoin/           # Custom VITACOIN module
│   │       ├── keeper/         # State management
│   │       ├── types/          # Data types
│   │       ├── module.go       # Module definition
│   │       └── genesis.go      # Genesis handling
│   │
│   ├── testutil/                # Testing utilities
│   └── build/                   # Compiled binaries
│       └── vitacoind           # Blockchain node binary
│
├── vitapay/                      # 💳 THE PAYMENT NETWORK
│   ├── README.md                # VITAPAY-specific docs
│   │
│   ├── mobile-wallet/           # 📱 Customer mobile app
│   │   ├── README.md
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   │
│   │   ├── src/                # React Native source
│   │   │   ├── App.tsx
│   │   │   ├── screens/        # App screens
│   │   │   ├── components/     # UI components
│   │   │   ├── store/          # Redux store
│   │   │   ├── services/       # Business logic
│   │   │   └── utils/          # Helper functions
│   │   │
│   │   ├── ios/                # iOS-specific
│   │   │   └── VITAPAYWallet/
│   │   │
│   │   └── android/            # Android-specific
│   │       └── app/
│   │
│   ├── payment-gateway/         # 🌐 Merchant payment API
│   │   ├── README.md
│   │   ├── go.mod
│   │   ├── Makefile
│   │   │
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go    # API server entry point
│   │   │
│   │   ├── internal/
│   │   │   ├── api/           # HTTP handlers
│   │   │   ├── services/      # Business logic
│   │   │   ├── models/        # Data models
│   │   │   ├── repository/    # Database access
│   │   │   └── blockchain/    # VITACOIN client
│   │   │
│   │   ├── migrations/        # Database migrations
│   │   └── config/            # Configuration
│   │
│   ├── merchant-dashboard/      # 🖥️ Web dashboard
│   │   ├── README.md
│   │   ├── package.json
│   │   ├── next.config.js
│   │   │
│   │   ├── src/
│   │   │   ├── pages/         # Next.js pages
│   │   │   ├── components/    # React components
│   │   │   ├── hooks/         # Custom hooks
│   │   │   └── lib/           # Utilities
│   │   │
│   │   └── public/            # Static assets
│   │
│   └── shared/                  # 🔧 Shared VITAPAY utilities
│       └── vitacoin-client/    # TypeScript SDK for VITACOIN
│           ├── README.md
│           ├── package.json
│           ├── src/
│           │   ├── connection.ts
│           │   ├── transactions.ts
│           │   └── wallet.ts
│           └── tests/
│
├── shared/                       # 🔗 SHARED BETWEEN BOTH PROJECTS
│   ├── types/                   # Common TypeScript/Go types
│   │   ├── payment.ts
│   │   ├── transaction.ts
│   │   └── wallet.ts
│   │
│   └── utils/                   # Common utilities
│       ├── crypto.ts
│       ├── validation.ts
│       └── formatting.ts
│
├── scripts/                      # 🛠️ BUILD & DEPLOYMENT SCRIPTS
│   ├── build-vitacoin.sh        # Build blockchain
│   ├── build-vitapay.sh         # Build payment network
│   ├── test-all.sh              # Run all tests
│   ├── localnet.sh              # Start local testnet
│   ├── protocgen.sh             # Generate protobuf
│   └── migrate/                 # Migration scripts
│
├── go/                           # 🗑️ TO BE REMOVED (Go source)
└── github-setup.sh               # GitHub repository setup
```

---

## 🎯 Key Folders Explained

### Root Level

| Folder/File | Purpose | Who Uses It |
|------------|---------|-------------|
| `README.md` | **Ecosystem overview** - explains both projects | Everyone (start here!) |
| `docs/` | **All documentation** for both projects | Developers, users, contributors |
| `vitacoin/` | **Blockchain code** - the cryptocurrency | Blockchain developers, validators |
| `vitapay/` | **Payment network code** - user-facing apps | App developers, merchants |
| `shared/` | **Shared utilities** used by both projects | All developers |
| `scripts/` | **Build scripts** for automation | DevOps, CI/CD |

### `docs/` - Documentation

**Purpose**: Single source of truth for all documentation.

**Structure**:
```
docs/
├── README.md              # Navigation hub
├── VITACOIN.md           # Complete cryptocurrency guide
├── FOLDER_STRUCTURE.md   # This file
│
├── architecture/          # How the system works
├── development/           # Developer guides
└── project/              # Project management
    ├── VITAPAY.md        # Payment network guide
    └── MOBILE_APP.md     # Wallet app specs
```

**When to update**:
- New feature added → update relevant guide
- Architecture change → update ARCHITECTURE.md
- New API endpoint → update VITAPAY.md
- Roadmap shift → update DEVELOPMENT_ROADMAP.md

### `vitacoin/` - Cryptocurrency Blockchain

**Purpose**: The blockchain infrastructure - validators, consensus, token economics.

**Key Files**:
- `vitacoin/README.md` - Blockchain-specific docs
- `vitacoin/cmd/vitacoind/` - Node software binary
- `vitacoin/x/vitacoin/` - Custom blockchain module
- `vitacoin/proto/` - Protocol buffer definitions

**Who works here**:
- Blockchain developers
- Validator operators
- Protocol designers

**Tech stack**: Go, Cosmos SDK, CometBFT

### `vitapay/` - Payment Network

**Purpose**: User-facing applications and merchant tools.

**Sub-folders**:

#### `vitapay/mobile-wallet/`
- **What**: Customer mobile app (iOS & Android)
- **Tech**: React Native, TypeScript
- **Users**: Consumers making payments

#### `vitapay/payment-gateway/`
- **What**: Merchant API for accepting payments
- **Tech**: Go, PostgreSQL, Redis
- **Users**: E-commerce sites, merchants

#### `vitapay/merchant-dashboard/`
- **What**: Web portal for merchants
- **Tech**: Next.js, TypeScript, Tailwind
- **Users**: Business owners managing payments

#### `vitapay/shared/vitacoin-client/`
- **What**: TypeScript SDK for interacting with VITACOIN blockchain
- **Tech**: TypeScript, CosmJS
- **Users**: All VITAPAY components

### `shared/` - Cross-Project Utilities

**Purpose**: Code used by BOTH vitacoin and vitapay.

**Contents**:
- `shared/types/` - Common data types
- `shared/utils/` - Helper functions (crypto, validation, formatting)

**Example**:
```typescript
// Used by both mobile wallet and payment gateway
import { validateAddress } from '@shared/utils/validation';

if (!validateAddress(userInput)) {
  throw new Error('Invalid VITA address');
}
```

### `scripts/` - Automation Scripts

**Purpose**: Build, test, and deploy automation.

**Key Scripts**:
- `build-vitacoin.sh` - Compile blockchain
- `build-vitapay.sh` - Build all VITAPAY apps
- `test-all.sh` - Run entire test suite
- `localnet.sh` - Start local testnet for development
- `protocgen.sh` - Generate protobuf code

---

## 🔀 Navigation Guide

### "I want to understand the ecosystem"
```
Start: README.md (root)
Then: docs/VITACOIN.md + docs/project/VITAPAY.md
```

### "I want to run the blockchain"
```
Go to: vitacoin/
Read: vitacoin/README.md
Run: cd vitacoin && make install && make localnet
```

### "I want to build the mobile app"
```
Go to: vitapay/mobile-wallet/
Read: vitapay/mobile-wallet/README.md
Run: cd vitapay/mobile-wallet && npm install && npm run ios
```

### "I want to integrate merchant payments"
```
Go to: vitapay/payment-gateway/
Read: vitapay/payment-gateway/README.md
API Docs: docs/project/VITAPAY.md
```

### "I want to contribute"
```
Read: CONTRIBUTING.md (coming soon)
Check: docs/project/TODO.md
Pick a task and start!
```

---

## 🚀 Development Workflow

### Setting Up Full Environment

```bash
# 1. Clone repository
git clone https://github.com/esspron/vitacoin
cd vitacoin

# 2. Build blockchain
cd vitacoin
make install
make localnet

# 3. Setup payment gateway (in new terminal)
cd ../vitapay/payment-gateway
go mod download
make migrate-up
make run

# 4. Setup mobile wallet (in new terminal)
cd ../vitapay/mobile-wallet
npm install
npm run ios  # or npm run android

# 5. Setup dashboard (in new terminal)
cd ../vitapay/merchant-dashboard
npm install
npm run dev
```

### Making Changes

**Blockchain Changes**:
```bash
cd vitacoin/
# Edit code in x/vitacoin/
make test
make build
./build/vitacoind start
```

**Mobile App Changes**:
```bash
cd vitapay/mobile-wallet/
# Edit code in src/
npm test
npm run ios
```

**Payment Gateway Changes**:
```bash
cd vitapay/payment-gateway/
# Edit code in internal/
go test ./...
make run
```

---

## 📦 Where to Put New Code

### New Blockchain Feature
```
Add to: vitacoin/x/vitacoin/
Example: vitacoin/x/vitacoin/keeper/staking.go
```

### New Mobile App Screen
```
Add to: vitapay/mobile-wallet/src/screens/
Example: vitapay/mobile-wallet/src/screens/SettingsScreen.tsx
```

### New API Endpoint
```
Add to: vitapay/payment-gateway/internal/api/handlers/
Example: vitapay/payment-gateway/internal/api/handlers/refunds.go
```

### New Dashboard Page
```
Add to: vitapay/merchant-dashboard/src/pages/
Example: vitapay/merchant-dashboard/src/pages/analytics.tsx
```

### Shared Utility Function
```
Add to: shared/utils/
Example: shared/utils/address-formatter.ts
```

### Documentation
```
Add to: docs/
Example: docs/development/NEW_FEATURE_GUIDE.md
```

---

## 🔄 When to Split the Monorepo?

**Consider splitting when:**
- ✅ Both projects are stable and production-ready
- ✅ VITACOIN has external developers building on it
- ✅ VITAPAY needs separate versioning
- ✅ Different teams own each project
- ✅ Independent release cycles needed

**How to split**:
1. Create new repo: `vitacoin-blockchain`
2. Create new repo: `vitapay-network`
3. Extract respective folders
4. Publish VITACOIN as npm/go package
5. Update VITAPAY to import published package
6. Archive monorepo or keep as meta-repo

---

## 🛠️ CI/CD Structure

```yaml
# .github/workflows/vitacoin.yml
name: VITACOIN Blockchain
on:
  push:
    paths:
      - 'vitacoin/**'
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Test blockchain
        run: cd vitacoin && make test

# .github/workflows/vitapay.yml
name: VITAPAY Payment Network
on:
  push:
    paths:
      - 'vitapay/**'
jobs:
  test-mobile:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v2
      - name: Test mobile wallet
        run: cd vitapay/mobile-wallet && npm test
  
  test-gateway:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Test payment gateway
        run: cd vitapay/payment-gateway && go test ./...
```

---

## 📝 Quick Reference

### README Hierarchy
```
README.md                              # Ecosystem overview
├── vitacoin/README.md                 # Blockchain docs
├── vitapay/README.md                  # Payment network docs
│   ├── mobile-wallet/README.md        # Wallet app docs
│   ├── payment-gateway/README.md      # Gateway API docs
│   └── merchant-dashboard/README.md   # Dashboard docs
└── docs/README.md                     # Documentation index
```

### Documentation Locations
| Topic | File |
|-------|------|
| Cryptocurrency guide | `docs/VITACOIN.md` |
| Payment network guide | `docs/project/VITAPAY.md` |
| Mobile wallet specs | `docs/project/MOBILE_APP.md` |
| Architecture | `docs/architecture/ARCHITECTURE.md` |
| Development setup | `docs/architecture/DEV_SETUP.md` |
| Roadmap | `docs/project/DEVELOPMENT_ROADMAP.md` |
| TODO list | `docs/project/TODO.md` |

### Code Locations
| Component | Path |
|-----------|------|
| Blockchain node | `vitacoin/cmd/vitacoind/` |
| Custom module | `vitacoin/x/vitacoin/` |
| Mobile wallet | `vitapay/mobile-wallet/src/` |
| Payment API | `vitapay/payment-gateway/internal/` |
| Merchant dashboard | `vitapay/merchant-dashboard/src/` |
| Shared SDK | `vitapay/shared/vitacoin-client/` |

---

## 🤝 Contributing Guidelines

### Before Adding Files
1. **Check if folder exists** for your component
2. **Read the README** in that folder
3. **Follow the structure** already established
4. **Update documentation** if adding new features

### Naming Conventions
- **Folders**: lowercase with hyphens (`mobile-wallet`, `payment-gateway`)
- **Go files**: lowercase with underscores (`payment_handler.go`)
- **TypeScript files**: PascalCase for components (`PaymentScreen.tsx`), camelCase for utilities (`validation.ts`)
- **Documentation**: UPPERCASE with underscores (`DEVELOPMENT_ROADMAP.md`)

### Git Workflow
```bash
# Create feature branch
git checkout -b feature/payment-gateway-webhooks

# Make changes in appropriate folder
cd vitapay/payment-gateway
# ... make changes ...

# Commit with clear scope
git commit -m "vitapay(gateway): add webhook retry logic"

# Push and create PR
git push origin feature/payment-gateway-webhooks
```

---

## 📞 Questions?

**Structure unclear?** Open an issue: [GitHub Issues](https://github.com/esspron/vitacoin/issues)  
**Can't find something?** Check `docs/README.md` for full documentation index  
**Want to contribute?** Read `CONTRIBUTING.md` (coming soon)

---

**Last Updated**: October 16, 2025  
**Version**: 1.0.0 (Monorepo Structure Established)
