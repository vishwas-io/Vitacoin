# VITACOIN - Getting Started Guide

## 🎯 Overview

This guide will walk you through the complete development process for VITACOIN, from setup to deployment.

---

## 📋 Prerequisites

### Required Software
- **Go**: Version 1.21 or higher
- **Make**: Build automation
- **Git**: Version control
- **Protocol Buffers**: v3.x for proto generation
- **Docker** (optional): For containerization

### Recommended Knowledge
- Basic Go programming
- Understanding of blockchain concepts
- Familiarity with command line
- Basic understanding of Cosmos SDK (we'll teach you)

---

## 🚀 Quick Start (5 Steps)

### Step 1: Verify Prerequisites
```bash
# Check Go version
go version  # Should be 1.21+

# Check Make
make --version

# Check Git
git version

# Check Protobuf Compiler
protoc --version  # Should be 3.x+
```

### Step 2: Install Dependencies
```bash
# Install protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Install protoc-gen-go-grpc
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Install protoc-gen-gocosmos
go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@latest

# Verify installation
which protoc-gen-go
which protoc-gen-gocosmos
```

### Step 3: Update Dependencies
```bash
# We'll run this command to update go.mod
go mod tidy
```

### Step 4: Generate Proto Code
```bash
# We'll create the script, then run:
make proto-gen
```

### Step 5: Build & Test
```bash
# Build the blockchain
make build

# Run tests
make test

# Install binary
make install
```

---

## 📚 Development Workflow

### Phase-by-Phase Development

We'll follow this systematic approach:

#### **Phase 1: Foundation (Week 1-2)**
1. ✅ Create documentation (DONE!)
2. Update go.mod with dependencies
3. Create Makefile
4. Setup proto generation
5. Create proto files

#### **Phase 2: Core Implementation (Week 3-6)**
1. Implement types package
2. Implement keeper package
3. Wire app.go
4. Create CLI commands
5. Test core functionality

#### **Phase 3: Advanced Features (Week 7-10)**
1. Liquid staking
2. Time-locked vaults
3. Reward pools
4. Fee burning
5. Advanced testing

#### **Phase 4: Integration (Week 11-12)**
1. IBC setup
2. Smart contracts (CosmWasm)
3. DeFi primitives
4. NFT support

#### **Phase 5: Testing & Audit (Week 13-16)**
1. Comprehensive testing
2. Security audit
3. Bug fixes
4. Performance optimization

#### **Phase 6: Deployment (Week 17-20)**
1. Testnet deployment
2. Community testing
3. Mainnet preparation
4. Launch!

---

## 🛠️ Project Structure Explained

```
vitacoin/
├── proto/                    # Protocol Buffer definitions
│   └── vitacoin/v1/         # VITA module protos
│       ├── genesis.proto    # Genesis state
│       ├── params.proto     # Module parameters
│       ├── query.proto      # Query service
│       ├── tx.proto         # Transaction messages
│       ├── vault.proto      # Vault types
│       └── pool.proto       # Pool types
│
├── x/vitacoin/              # Custom VITA module
│   ├── keeper/              # State management
│   ├── types/               # Type definitions
│   ├── client/cli/          # CLI commands
│   └── module.go            # Module setup
│
├── app/                     # Application setup
│   ├── app.go              # Main app wiring
│   ├── encoding.go         # Encoding config
│   └── params.go           # Global parameters
│
├── cmd/vitacoind/          # Command line binary
│   ├── main.go             # Entry point
│   └── cmd/                # Subcommands
│
├── tests/                   # Test suites
│   ├── unit/               # Unit tests
│   ├── integration/        # Integration tests
│   └── e2e/                # End-to-end tests
│
├── scripts/                 # Build scripts
│   ├── protocgen.sh        # Proto generation
│   └── build.sh            # Build scripts
│
├── docs/                    # Documentation
│   ├── DEVELOPMENT_ROADMAP.md  # This roadmap
│   ├── ARCHITECTURE.md         # Technical architecture
│   └── TODO.md                 # Task tracker
│
├── go.mod                   # Go dependencies
├── go.sum                   # Dependency checksums
├── Makefile                # Build automation
└── README.md               # Project overview
```

---

## 📖 Key Concepts to Understand

### 1. **Cosmos SDK Architecture**
- **Modules**: Self-contained functionality (auth, bank, staking)
- **Keepers**: State management within modules
- **Messages**: User actions (send tokens, delegate)
- **Queries**: Read state (get balance, validator info)
- **Events**: Notifications of state changes

### 2. **Protocol Buffers**
- Define data structures
- Language-agnostic
- Generate Go code automatically
- Used for APIs and storage

### 3. **ABCI (Application Blockchain Interface)**
- Interface between blockchain (CometBFT) and app
- Handles consensus, networking, mempool
- Your app focuses on business logic

### 4. **State Management**
- **IAVL Tree**: Merkle tree for state storage
- **Store Keys**: Identify different storage spaces
- **KV Store**: Key-value storage per module

### 5. **Transaction Lifecycle**
- User creates and signs transaction
- Broadcast to network
- Validation (CheckTx)
- Consensus (block proposal)
- Execution (DeliverTx)
- Finalization (commit)

---

## 🎓 Learning Resources

### Official Documentation
- **Cosmos SDK**: https://docs.cosmos.network
- **CometBFT**: https://docs.cometbft.com
- **IBC**: https://ibc.cosmos.network
- **CosmWasm**: https://docs.cosmwasm.com

### Tutorials
- Cosmos SDK Tutorials: https://tutorials.cosmos.network
- Building a blockchain: https://github.com/cosmos/sdk-tutorials

### Community
- Discord: Cosmos Network Discord
- Forum: https://forum.cosmos.network
- Twitter: @cosmos

---

## 💡 Development Tips

### 1. **Start Simple**
- Get basic functionality working first
- Add advanced features incrementally
- Test each feature thoroughly

### 2. **Use Existing Modules**
- Don't reinvent the wheel
- Cosmos SDK has many built-in modules
- Study existing module implementations

### 3. **Test Everything**
- Write tests as you code
- Test edge cases
- Test error conditions
- Use simulation testing

### 4. **Document as You Go**
- Comment complex logic
- Update documentation
- Create examples
- Write guides

### 5. **Security First**
- Validate all inputs
- Check permissions
- Handle errors properly
- Consider attack vectors

---

## 🔧 Common Commands

### Development
```bash
# Update dependencies
go mod tidy

# Generate proto code
make proto-gen

# Build binary
make build

# Install binary
make install

# Run tests
make test

# Run linter
make lint

# Format code
make format
```

### Running the Blockchain
```bash
# Initialize node
vitacoind init mynode --chain-id vitacoin-1

# Add key
vitacoind keys add mykey

# Add genesis account
vitacoind add-genesis-account mykey 1000000000uvita

# Create genesis transaction
vitacoind gentx mykey 100000000uvita --chain-id vitacoin-1

# Collect genesis txs
vitacoind collect-gentxs

# Start node
vitacoind start
```

### Interacting with the Chain
```bash
# Check balance
vitacoind query bank balances <address>

# Send tokens
vitacoind tx bank send <from> <to> 1000uvita --chain-id vitacoin-1

# Query validators
vitacoind query staking validators

# Delegate
vitacoind tx staking delegate <validator> 1000uvita --from mykey
```

---

## 🐛 Troubleshooting

### Common Issues

#### 1. **"protoc: command not found"**
```bash
# Install protoc
# macOS:
brew install protobuf

# Ubuntu/Debian:
sudo apt install protobuf-compiler

# Or download from: https://github.com/protocolbuffers/protobuf/releases
```

#### 2. **"cannot find package"**
```bash
# Run go mod tidy
go mod tidy
go mod download
```

#### 3. **"proto file not found"**
```bash
# Make sure proto imports are correct
# Check GOPATH and module cache
go env GOPATH
```

#### 4. **Build errors**
```bash
# Clean and rebuild
make clean
make build
```

---

## 📊 Progress Tracking

We'll use three documents to track progress:

1. **DEVELOPMENT_ROADMAP.md**: Overall project plan
2. **TODO.md**: Current sprint tasks
3. **ARCHITECTURE.md**: Technical specifications

Update these as we complete tasks!

---

## 🎯 Next Steps

**Right now, we need to**:
1. Update `go.mod` with proper dependencies
2. Create the `Makefile`
3. Setup proto generation scripts
4. Create proto files
5. Generate Go code

**Let's start with Step 1!** 🚀

Ready to begin implementation? Just say "let's start" and I'll begin with updating the go.mod file!

---

## 💬 Questions?

Throughout this process:
- Ask questions anytime
- Request explanations for any concept
- We'll go step-by-step
- I'll explain what each piece does

---

**Version**: 1.0.0  
**Created**: October 15, 2025  
**Status**: Ready to Start Development! 🎉
