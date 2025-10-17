# 🗺️ VITACOIN Quick Reference Guide

## 📚 Documentation Map

```
VITACOIN Documentation Suite
│
├── 📖 README.md
│   └── Start here! Project overview, quick start, basic usage
│
├── 🚀 GETTING_STARTED.md
│   └── Complete beginner guide, setup instructions, learning path
│
├── 🗺️ DEVELOPMENT_ROADMAP.md
│   └── Full project plan (17 phases), all features, timeline
│
├── 🏗️ ARCHITECTURE.md
│   └── Technical deep dive, system design, implementation details
│
├── ✅ TODO.md
│   └── Current tasks, progress tracking, next actions
│
└── 📋 PROJECT_SUMMARY.md
    └── High-level overview, decisions, what we've built

🎯 YOU ARE HERE → QUICK_REFERENCE.md
    └── Visual guide, cheat sheets, quick lookup
```

---

## 🎯 Project at a Glance

| Aspect | Details |
|--------|---------|
| **Name** | VITACOIN |
| **Symbol** | VITA |
| **Type** | Advanced Cosmos SDK Blockchain |
| **Supply** | 1,000,000,000 VITA |
| **Purpose** | DeFi + Smart Contracts + NFTs + IBC |
| **Status** | 📝 Documentation Complete → 🔨 Ready to Build |

---

## 🏗️ Development Phases Quick View

```
Week 1-2   ⚡ Foundation
Week 3-6   🏗️ Core Module
Week 7-10  💎 Advanced Features
Week 11-12 🌐 IBC & Smart Contracts
Week 13-16 🔒 Security & Testing
Week 17-20 🧪 Testnet
Week 21+   🚀 Mainnet Launch
```

---

## 🎨 Feature Set Overview

```
┌─────────────────────────────────────────┐
│         VITACOIN FEATURES               │
├─────────────────────────────────────────┤
│                                         │
│  🏦 CORE BLOCKCHAIN                     │
│  ├─ Token Transfers                     │
│  ├─ Account Management                  │
│  ├─ Transaction History                 │
│  └─ Fee Management                      │
│                                         │
│  🥩 STAKING & VALIDATION                │
│  ├─ Proof of Stake                      │
│  ├─ Validator Management                │
│  ├─ Delegation/Undelegation             │
│  ├─ Reward Distribution                 │
│  └─ Slashing Protection                 │
│                                         │
│  💧 LIQUID STAKING (Advanced)           │
│  ├─ Stake → Get stVITA                  │
│  ├─ Auto-Compounding                    │
│  ├─ DeFi Composability                  │
│  └─ Instant Liquidity                   │
│                                         │
│  🔒 TIME-LOCKED VAULTS (Advanced)       │
│  ├─ 1-12 Month Locks                    │
│  ├─ Enhanced Rewards (1x-2x)            │
│  ├─ Early Withdrawal Penalties          │
│  └─ Vault Management                    │
│                                         │
│  🎁 REWARD POOLS (Advanced)             │
│  ├─ Custom Pool Creation                │
│  ├─ Multi-Token Rewards                 │
│  ├─ Time-Based Distribution             │
│  └─ Performance Bonuses                 │
│                                         │
│  🔥 FEE BURNING (Advanced)              │
│  ├─ Deflationary Mechanism              │
│  ├─ Automatic Burn %                    │
│  ├─ Supply Tracking                     │
│  └─ Value Appreciation                  │
│                                         │
│  🗳️ GOVERNANCE                          │
│  ├─ Proposal Creation                   │
│  ├─ Voting System                       │
│  ├─ Parameter Changes                   │
│  └─ Treasury Management                 │
│                                         │
│  🌐 IBC (Inter-Blockchain)              │
│  ├─ Cross-Chain Transfers               │
│  ├─ Multi-Chain DeFi                    │
│  ├─ Channel Management                  │
│  └─ Packet Relay                        │
│                                         │
│  📜 SMART CONTRACTS (CosmWasm)          │
│  ├─ Contract Deployment                 │
│  ├─ CW20 Tokens                         │
│  ├─ CW721 NFTs                          │
│  └─ Custom Logic                        │
│                                         │
│  💰 DeFi PRIMITIVES                     │
│  ├─ AMM Pools (Swap)                    │
│  ├─ Lending/Borrowing                   │
│  ├─ Yield Farming                       │
│  └─ Derivatives                         │
│                                         │
│  🎨 NFT SYSTEM                          │
│  ├─ NFT Minting                         │
│  ├─ NFT Trading                         │
│  ├─ NFT Staking                         │
│  └─ Marketplace                         │
│                                         │
└─────────────────────────────────────────┘
```

---

## 🔑 Key Concepts Cheat Sheet

### Transaction Types

| Transaction | Command Example | Purpose |
|-------------|-----------------|---------|
| **Send** | `vitacoind tx bank send` | Transfer VITA |
| **Delegate** | `vitacoind tx staking delegate` | Stake to validator |
| **Undelegate** | `vitacoind tx staking unbond` | Unstake from validator |
| **Redelegate** | `vitacoind tx staking redelegate` | Move stake |
| **Vote** | `vitacoind tx gov vote` | Vote on proposal |
| **Create Vault** | `vitacoind tx vitacoin create-vault` | Lock tokens |
| **Liquid Stake** | `vitacoind tx vitacoin liquid-stake` | Get stVITA |

### Query Types

| Query | Command Example | Returns |
|-------|-----------------|---------|
| **Balance** | `vitacoind query bank balances` | Token balances |
| **Validators** | `vitacoind query staking validators` | All validators |
| **Delegations** | `vitacoind query staking delegations` | Your stakes |
| **Proposals** | `vitacoind query gov proposals` | All proposals |
| **Vaults** | `vitacoind query vitacoin vaults` | Your vaults |

---

## 💾 File Structure Cheat Sheet

### Where to Find What

```
vitacoin/
│
├── 📄 Documentation (Read These First)
│   ├── README.md              ← Project overview
│   ├── GETTING_STARTED.md     ← Setup guide
│   ├── DEVELOPMENT_ROADMAP.md ← Full plan
│   ├── ARCHITECTURE.md        ← Technical details
│   ├── TODO.md                ← Current tasks
│   ├── PROJECT_SUMMARY.md     ← High-level summary
│   └── QUICK_REFERENCE.md     ← This file
│
├── 🔧 Proto Definitions (API & Types)
│   └── proto/vitacoin/v1/
│       ├── genesis.proto      ← Initial state
│       ├── params.proto       ← Parameters
│       ├── query.proto        ← Query API
│       ├── tx.proto           ← Transaction API
│       ├── vault.proto        ← Vault types
│       └── pool.proto         ← Pool types
│
├── 🎯 Custom Module (Our Code)
│   └── x/vitacoin/
│       ├── keeper/            ← State management
│       │   ├── keeper.go      ← Core keeper
│       │   ├── msg_server.go  ← Handle transactions
│       │   ├── query_server.go← Handle queries
│       │   ├── vault.go       ← Vault logic
│       │   ├── pool.go        ← Pool logic
│       │   └── liquid_staking.go ← Liquid staking
│       │
│       ├── types/             ← Type definitions
│       │   ├── codec.go       ← Encoding
│       │   ├── keys.go        ← Storage keys
│       │   ├── errors.go      ← Error types
│       │   ├── events.go      ← Event types
│       │   ├── genesis.go     ← Genesis types
│       │   ├── params.go      ← Parameters
│       │   ├── msgs.go        ← Message types
│       │   └── expected_keepers.go ← Interfaces
│       │
│       ├── client/cli/        ← Command line
│       │   ├── query.go       ← Query commands
│       │   └── tx.go          ← Transaction commands
│       │
│       └── module.go          ← Module definition
│
├── 🏗️ Application (Wire Everything)
│   └── app/
│       ├── app.go             ← Main app setup
│       ├── encoding.go        ← Encoding config
│       └── params.go          ← Global params
│
├── 💻 Command Line Binary
│   └── cmd/vitacoind/
│       ├── main.go            ← Entry point
│       └── cmd/
│           ├── root.go        ← Root command
│           ├── genesis.go     ← Genesis commands
│           ├── init.go        ← Init commands
│           └── config.go      ← Config commands
│
├── 🧪 Tests
│   └── tests/
│       ├── unit/              ← Unit tests
│       ├── integration/       ← Integration tests
│       └── e2e/               ← End-to-end tests
│
├── 📜 Build & Scripts
│   ├── Makefile               ← Build commands
│   └── scripts/
│       ├── protocgen.sh       ← Generate proto code
│       └── build.sh           ← Build scripts
│
└── ⚙️ Configuration
    ├── go.mod                 ← Dependencies
    └── config.yml             ← Chain config
```

---

## 🎯 Development Workflow

### Daily Workflow
```
1. Check TODO.md
   ↓
2. Pick a task
   ↓
3. Read relevant ARCHITECTURE section
   ↓
4. Implement feature
   ↓
5. Write tests
   ↓
6. Update TODO.md
   ↓
7. Commit & repeat
```

### When Stuck
```
1. Check GETTING_STARTED.md troubleshooting
   ↓
2. Review ARCHITECTURE.md for that component
   ↓
3. Look at similar Cosmos SDK modules
   ↓
4. Check Cosmos SDK documentation
   ↓
5. Ask for help
```

---

## 🔧 Common Commands Reference

### Setup Commands
```bash
# Install dependencies
go mod tidy

# Generate proto code
make proto-gen

# Build binary
make build

# Install binary
make install
```

### Node Commands
```bash
# Initialize node
vitacoind init <moniker> --chain-id vitacoin-1

# Add key
vitacoind keys add <key-name>

# Start node
vitacoind start
```

### Transaction Commands
```bash
# Send tokens
vitacoind tx bank send <from> <to> <amount> --chain-id vitacoin-1

# Delegate
vitacoind tx staking delegate <validator> <amount> --from <key>

# Vote
vitacoind tx gov vote <proposal-id> yes --from <key>
```

### Query Commands
```bash
# Check balance
vitacoind query bank balances <address>

# List validators
vitacoind query staking validators

# Check proposal
vitacoind query gov proposal <id>
```

---

## 📊 Token Economics Quick Reference

### Supply Breakdown
```
Total: 1,000,000,000 VITA
│
├── 40% (400M) → Staking Rewards (10 years)
├── 30% (300M) → Genesis Allocation
├── 20% (200M) → Ecosystem Development
└── 10% (100M) → Governance Reserve
```

### Reward Multipliers
```
Vault Duration     Multiplier
─────────────────────────────
1 month       →    1.0x
3 months      →    1.2x
6 months      →    1.5x
12 months     →    2.0x
```

### Inflation
```
Base: 7% per year
Range: 3% - 10%
Adjustment: Monthly
Target Bonded: 67%
```

---

## 🎓 Learning Path

### Week 1: Foundations
- [ ] Read README.md
- [ ] Read GETTING_STARTED.md
- [ ] Setup development environment
- [ ] Learn Go basics (if needed)
- [ ] Understand blockchain basics

### Week 2: Cosmos SDK
- [ ] Study Cosmos SDK architecture
- [ ] Read ARCHITECTURE.md
- [ ] Understand modules
- [ ] Learn about keepers
- [ ] Understand ABCI

### Week 3-4: Implementation
- [ ] Follow DEVELOPMENT_ROADMAP.md
- [ ] Implement core features
- [ ] Write tests
- [ ] Debug and iterate

### Week 5+: Advanced
- [ ] Add advanced features
- [ ] Optimize performance
- [ ] Security review
- [ ] Documentation

---

## 🚀 Current Status

```
Phase 1: Documentation  ✅ 100% Complete
├── README.md                    ✅
├── GETTING_STARTED.md           ✅
├── DEVELOPMENT_ROADMAP.md       ✅
├── ARCHITECTURE.md              ✅
├── TODO.md                      ✅
├── PROJECT_SUMMARY.md           ✅
└── QUICK_REFERENCE.md           ✅

Phase 2: Foundation Setup  ⏳ 0% Complete
├── go.mod update                ⏳
├── Makefile creation            ⏳
├── Proto generation script      ⏳
├── Proto files                  ⏳
└── Generated Go code            ⏳

Overall Progress: 1/17 phases (6%)
```

---

## 💡 Pro Tips

### Development
- ✅ Read documentation first
- ✅ Understand before coding
- ✅ Write tests as you go
- ✅ Commit frequently
- ✅ Document your code

### Debugging
- ✅ Check logs first
- ✅ Use debugger
- ✅ Add print statements
- ✅ Test in isolation
- ✅ Ask for help when stuck

### Testing
- ✅ Unit test everything
- ✅ Test edge cases
- ✅ Test error conditions
- ✅ Integration test
- ✅ Manual testing

---

## 📞 Quick Help

### "I want to..."
- **Understand the project** → Read README.md
- **Get started coding** → Read GETTING_STARTED.md
- **See all features** → Read DEVELOPMENT_ROADMAP.md
- **Understand architecture** → Read ARCHITECTURE.md
- **Know what to do next** → Read TODO.md
- **See high-level overview** → Read PROJECT_SUMMARY.md
- **Quick lookup** → Read QUICK_REFERENCE.md (this file)

### "I'm stuck on..."
- **Setup** → GETTING_STARTED.md → Troubleshooting
- **Architecture** → ARCHITECTURE.md → Relevant section
- **Implementation** → DEVELOPMENT_ROADMAP.md → Phase details
- **Commands** → QUICK_REFERENCE.md → Commands section

---

## 🎯 Next Action

**Ready to start building?**

Say **"let's start implementing"** and we'll begin with:
1. Updating go.mod
2. Creating Makefile  
3. Setting up proto generation
4. Creating proto files
5. Generating Go code

Then we'll move into implementing the actual VITACOIN module!

---

<div align="center">

**VITACOIN Quick Reference**

Version 1.0.0 | October 15, 2025

[⬆ Back to Top](#-vitacoin-quick-reference-guide)

</div>
