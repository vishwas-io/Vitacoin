# VITACOIN Ecosystem - Copilot Instructions

## Project Overview

This is a **monorepo** containing two interconnected projects:
- **VITACOIN**: A Cosmos SDK blockchain (the cryptocurrency layer)
- **VITAPAY**: Payment network applications (mobile wallet, merchant gateway, dashboard)

Both projects share a single `go.mod` at the repository root (`/Blockchain Project/go.mod`).

## Critical Architecture Patterns

### Monorepo Structure
```
/Blockchain Project/
├── go.mod                    # Shared Go module root
├── vitacoin/
│   ├── vitacoin/            # Blockchain code (Cosmos SDK)
│   │   ├── x/vitacoin/      # Custom blockchain module
│   │   ├── app/             # Application wiring
│   │   ├── cmd/vitacoind/   # CLI binary
│   │   └── proto/           # Protocol buffers
│   ├── vitapay/             # Payment network apps
│   └── docs/                # All documentation
```

**Key Insight**: Import paths must include the full monorepo structure from root:
```go
// Correct:
import "github.com/vitacoin/vitacoin/vitacoin/vitacoin/x/vitacoin/types"

// Wrong:
import "github.com/vitacoin/vitacoin/x/vitacoin/types"
```

### Build System

**Always build from project root** (not from `vitacoin/vitacoin/`):
```bash
cd "/Users/vishwasverma/Downloads/Blockchain Project"
go build -mod=readonly -o vitacoin/vitacoin/build/vitacoind ./vitacoin/vitacoin/cmd/vitacoind
```

The `Makefile` in `vitacoin/vitacoin/` handles this by using `PROJECT_ROOT`:
```makefile
PROJECT_ROOT := $(shell cd ../.. && pwd)
# Then: cd "$(PROJECT_ROOT)" && go build ...
```

**Critical Make Commands**:
- `make build` - Compile the blockchain binary
- `make proto-gen` - Regenerate protobuf code (run from project root)
- `make test` - Run tests
- `make install` - Install binary to `$GOPATH/bin`

### Protocol Buffers Workflow

**Proto generation is special** - must run `scripts/protocgen.sh` from project root:
1. Proto definitions in `vitacoin/vitacoin/proto/vitacoin/v1/*.proto`
2. Script generates `.pb.go` files in `vitacoin/vitacoin/x/vitacoin/types/`
3. **Never edit** generated `*.pb.go` files - they get overwritten
4. Custom logic goes in separate files like `types/validation.go`, `types/msgs.go`

### Cosmos SDK Module Architecture

VITACOIN uses standard Cosmos SDK v0.50.3 patterns:

**Module Structure** (`x/vitacoin/`):
- `types/` - Data structures, messages, queries (mostly generated from proto)
- `keeper/` - State management and business logic (**this is where custom code lives**)
- `module.go` - Module registration and lifecycle hooks
- `client/cli/` - CLI command definitions

**State Management Pattern**:
```go
// Keeper methods follow this pattern:
func (k Keeper) CreateVault(ctx sdk.Context, msg *types.MsgCreateVault) error {
    // 1. Validate inputs
    // 2. Load state from store
    // 3. Execute business logic
    // 4. Save state to store
    // 5. Emit events
    return nil
}
```

**Store Keys** are defined in `types/keys.go`:
```go
const (
    VaultPrefix       = "vault:"
    PoolPrefix        = "pool:"
    // Keys define how data is organized in the IAVL tree
)
```

## Development Workflows

### Adding a New Proto Message

1. Add to `proto/vitacoin/v1/tx.proto` (transactions) or `query.proto` (queries)
2. Run `make proto-gen` from `vitacoin/vitacoin/`
3. Implement handler in `keeper/msg_server.go` or `keeper/query_server.go`
4. Add validation in `types/validation.go`
5. Register in `types/codec.go` if needed

### Adding a New Keeper Method

1. Define in `keeper/keeper.go` or create new file like `keeper/vault.go`
2. Access state via `k.storeService.OpenKVStore(ctx)`
3. Use consistent error handling: `return types.ErrInvalidVault.Wrap("reason")`
4. Emit events: `ctx.EventManager().EmitEvent(...)`
5. Write tests in matching `*_test.go` file

### Wiring App.go

The `app/app.go` file (658 lines) is the central wiring point:
- Module keepers are initialized in order (dependencies matter!)
- `ModuleBasics` registers CLI commands and REST endpoints
- `BeginBlockers` and `EndBlockers` define execution order
- **Don't modify unless adding new modules** - this is delicate

### Testing Strategy

```bash
# Unit tests - fast, test individual functions
cd vitacoin/vitacoin && make test-unit

# Run specific package
go test ./x/vitacoin/keeper/...

# With coverage
make test-cover  # Generates coverage.html
```

Tests go alongside code:
- `keeper/vault.go` → `keeper/vault_test.go`
- Use `testutil/` for test helpers and fixtures

## Common Gotchas

### Import Path Issues
If you see "package not found" errors, verify the import path includes the full monorepo path. The go.mod is at `/Blockchain Project/go.mod`, not in `vitacoin/`.

### Proto Generation Failures
- Ensure you're running from project root
- Check that `protoc`, `protoc-gen-gocosmos`, and `protoc-gen-grpc-gateway` are installed
- Generated files land in `x/vitacoin/types/` - don't edit them manually

### Build Failures After go.mod Changes
```bash
# From project root:
go mod tidy
go mod download
cd vitacoin/vitacoin && make build
```

### Spaces in Directory Paths
The project lives in `/Users/vishwasverma/Downloads/Blockchain Project/` (note spaces).
Always quote paths in shell commands and Makefiles.

## Project-Specific Conventions

### Module Naming
- Package: `vitacoin` (lowercase)
- Store prefix: `vitacoin` 
- CLI namespace: `vitacoind` (daemon pattern)
- Types prefix: `types.Msg*`, `types.Query*`

### Error Handling
Use sentinel errors from `types/errors.go`:
```go
var (
    ErrInvalidVault = errors.Register(ModuleName, 1101, "invalid vault")
    // Error codes 1000-1999 reserved for vitacoin module
)
```

### Event Patterns
Events use this structure (see `types/events.go`):
```go
const (
    EventTypeCreateVault = "create_vault"
    AttributeKeyVaultID  = "vault_id"
)

// Emit like:
sdk.NewEvent(
    types.EventTypeCreateVault,
    sdk.NewAttribute(types.AttributeKeyVaultID, vaultID),
)
```

### Fee Distribution Logic
VITACOIN implements transparent fee distribution (0.1% transaction fee):
- 50% → Validators
- 25% → Burned (deflationary)
- 25% → Treasury

This is **not yet implemented** but is a critical feature for Phase 2+.

## Key Features & Differentiators

### VITACOIN (Blockchain)
**Token**: VITA (1 billion total supply)
- Denomination: `uvita` (1 VITA = 10^18 uvita)
- Chain ID: `vitacoin-1` (mainnet), `vitacoin-local` (testnet)

**Unique Features** (to be implemented):
1. **Time-Locked Vaults** - Users lock tokens for rewards (1/3/6/12 months with multipliers 1x-2x)
2. **Liquid Staking** - Stake VITA, receive stVITA derivative tokens
3. **Custom Reward Pools** - Create pools for specific staking programs
4. **Transparent Fee Burning** - 25% of all fees burned on-chain, visible to all
5. **E-commerce Focus** - Built specifically for payment use cases, not general DeFi

**Token Economics**:
- Genesis allocation: 300M (30%) - team, advisors, early supporters
- Staking rewards: 400M (40%) - distributed over 10 years
- Ecosystem development: 200M (20%) - grants, partnerships, VITAPAY
- Governance reserve: 100M (10%) - community treasury

**Staking Parameters**:
- Target bonded ratio: 67%
- Unbonding period: 21 days
- Dynamic inflation: 3-10% annually
- Expected APR: ~7%

### VITAPAY (Payment Network)
**Purpose**: Make VITA cryptocurrency easy to use for everyday payments

**Components**:
1. **Mobile Wallet** - React Native app for iOS/Android
   - Non-custodial (users control keys)
   - QR code scanning
   - Contact management
   - Biometric security
   
2. **Payment Gateway** - Go API for merchants
   - Accept VITA on websites
   - Webhook notifications
   - Blockchain monitoring
   - KYC/AML integration
   
3. **Merchant Dashboard** - Next.js web portal
   - Transaction analytics
   - API key management
   - Revenue reporting
   - Webhook configuration

4. **E-commerce Plugins**
   - WordPress/WooCommerce
   - Shopify
   - Magento, PrestaShop

**Value Proposition**:
- **For Merchants**: 0.1% fee vs 2-3% traditional processors (97% savings)
- **For Customers**: Fast (5-second finality), private, no card data shared
- **For Developers**: Simple integration, clear API docs, multiple SDKs

## Current Development Phase

### VITACOIN Blockchain (Phase 1 → Phase 2)

**Phase 1 - Foundation (✅ 100% Complete)**:
- ✅ Proto definitions (4 files: genesis, params, query, tx)
- ✅ Generated .pb.go files (360KB+ code)
- ✅ Module structure (keeper, types, module.go)
- ✅ App.go wired with all Cosmos SDK modules
- ✅ Binary compiles (`./build/vitacoind` - 35MB)
- ✅ Basic CLI commands (help, init, export-genesis)
- ✅ Monorepo structure with proper import paths

**Phase 2 - Custom Module Implementation (🚧 Starting Now - 3 weeks)**:
- Week 3-4: Module structure (keeper methods, types validation, genesis)
- Week 5: Transaction handlers (MsgUpdateParams, validation, tests)
- **Next Steps**: Implement keeper methods in `x/vitacoin/keeper/`
- **Files to Create**: `keeper/vault.go`, `keeper/pool.go`, `keeper/msg_server.go`

**Future Phases** (See `vitacoin/vitacoin/TODO.md` for full roadmap):
- Phase 3: Token economics & fee distribution (0.1% fee, 50/25/25 split)
- Phase 4: Staking system (21-day unbonding, 67% target bonded)
- Phase 5: Governance (14-day voting, 40% quorum)
- Phase 6: IBC integration
- Phase 7-20: Security, CLI, APIs, DevOps, infrastructure
- **Mainnet Launch**: August 2026

### VITAPAY Payment Network (Phase 0 - Planning)

**Current Status (10% Complete)**:
- ✅ Folder structure created (`mobile-wallet`, `payment-gateway`, `merchant-dashboard`)
- ✅ README files written for each component
- 🚧 Technical specifications in progress
- ⏳ Wireframes and UI/UX design pending

**Upcoming Development** (Q2 2026 onwards):
- Phase 1: Mobile Wallet MVP (React Native, 8 weeks)
- Phase 3: Payment Gateway API (Go, 6 weeks)
- Phase 5: Merchant Dashboard (Next.js, 6 weeks)
- Phase 7-13: E-commerce plugins, DevOps, security, fiat integration
- **Public Launch**: Q2 2027

### Integration Points
- VITAPAY depends on VITACOIN mainnet (April 2026)
- Shared SDK in `vitapay/shared/vitacoin-client/`
- Cross-project dependencies managed in root `go.mod`

## Essential Commands Reference

```bash
# Navigate to blockchain
cd vitacoin/vitacoin

# Build blockchain
make build

# Test the binary
./build/vitacoind --help

# Initialize a test node
./build/vitacoind init mynode --chain-id vitacoin-local

# Development cycle
make proto-gen    # After proto changes
make build        # Compile
make test         # Verify
```

## Documentation Structure

Primary docs to consult:
- `vitacoin/README.md` - Ecosystem overview (start here!)
- `vitacoin/docs/VITACOIN.md` - Cryptocurrency guide
- `vitacoin/docs/architecture/ARCHITECTURE.md` - Technical deep dive
- `vitacoin/docs/development/GETTING_STARTED.md` - Developer onboarding
- `vitacoin/docs/FOLDER_STRUCTURE.md` - Navigation guide

For VITAPAY (payment network):
- `vitacoin/docs/project/VITAPAY.md` - Payment network overview
- `vitacoin/vitapay/mobile-wallet/` - React Native wallet app
- `vitacoin/vitapay/payment-gateway/` - Merchant API (Go)

## Integration Points

### VITACOIN ↔ VITAPAY Communication
- VITAPAY apps use CosmJS to interact with VITACOIN blockchain
- Shared SDK in `vitacoin/vitapay/shared/vitacoin-client/`
- Mobile wallet reads blockchain state via gRPC/REST
- Payment gateway monitors blockchain events for confirmations

### Future: IBC Integration
The blockchain uses Cosmos SDK with IBC support for cross-chain transfers.
IBC code is already wired in `app/app.go` but not yet configured.

## AI Assistant Guidelines

When working with this codebase:
1. **Always check import paths** - they must include the full monorepo structure
2. **Never edit `.pb.go` files** - modify the `.proto` source and regenerate
3. **Keeper methods are the core business logic** - most implementation work happens there
4. **Read existing code patterns** - see `keeper/params.go` for a simple example
5. **Test as you go** - write `*_test.go` files alongside implementation
6. **Check phase docs** - `docs/development/PHASE1_BUILD_SUCCESS.md` explains current state
7. **Makefile is your friend** - use it for all build operations, not direct `go` commands
8. **When in doubt, check Cosmos SDK docs** - this follows standard patterns from cosmos/cosmos-sdk v0.50.3

## Task Tracking System

This project uses a hierarchical TODO system:

**Master TODO**: `vitacoin/docs/project/TODO.md`
- High-level milestones and progress tracking
- Coordination between VITACOIN and VITAPAY teams
- Overall timeline and launch dates

**VITACOIN TODO**: `vitacoin/vitacoin/TODO.md` (788 lines, 20 phases)
- Detailed blockchain development tasks
- Currently on Phase 1 → Phase 2 transition
- Tracks: Custom modules, token economics, staking, governance, IBC, security, DevOps, deployment
- Includes production infrastructure plan (Phase 20: validator setup, Kubernetes, multi-region)

**VITAPAY TODO**: `vitacoin/vitapay/TODO.md` (677 lines, 13 phases)
- Payment network development tasks
- Currently in Phase 0 (Planning)
- Tracks: Mobile wallet, payment gateway, merchant dashboard, e-commerce plugins, fiat integration

### Current Sprint Priorities (October 2025)

**High Priority - VITACOIN**:
1. Implement keeper methods in `x/vitacoin/keeper/`
2. Add message handlers in `keeper/msg_server.go`
3. Write validation logic in `types/validation.go`
4. Unit tests for all keeper methods

**High Priority - VITAPAY**:
1. Complete technical specifications
2. Finalize mobile wallet wireframes
3. Set up development environments
4. Choose payment gateway infrastructure

**Blocked Items**:
- VITAPAY mobile wallet blocked on VITACOIN testnet (Q1 2026)
- Production deployment blocked on security audit completion

## Technology Stack

- **Language**: Go 1.21+ (custom installation in `vitacoin/go/`)
- **Framework**: Cosmos SDK v0.50.3
- **Consensus**: CometBFT (formerly Tendermint)
- **State Storage**: IAVL tree (via cosmos-db)
- **Proto**: Protocol Buffers with gogoproto extensions
- **CLI**: Cobra
- **Testing**: Go stdlib testing + Cosmos test utilities
- **Linting**: golangci-lint with 20+ linters

## Quick Debug Checklist

If things break:
1. ✅ Are you in the right directory? (Build from project root)
2. ✅ Did you run `go mod tidy` after dependency changes?
3. ✅ Are import paths correct? (Include full monorepo path)
4. ✅ Did proto generation succeed? (Check for `*.pb.go` files)
5. ✅ Is the Makefile using `PROJECT_ROOT` correctly?
6. ✅ Are there spaces in paths? (Quote them!)
