````instructions
````instructions
# VITACOIN Ecosystem - AI Agent Instructions

## Project Overview

VITACOIN is a **dual-project monorepo** building a complete cryptocurrency payment ecosystem:

**VITACOIN** (Cosmos SDK blockchain) - The cryptocurrency foundation
- Custom blockchain on Cosmos SDK v0.50.3 with CometBFT consensus
- VITA token with e-commerce features: merchant system, payment processing, time-locked vaults, reward pools
- Transparent fee distribution: 50% validators, 25% burn, 25% treasury
- Production-ready implementation: 5,000+ LOC keeper, 1,900+ LOC tests, 80+ functions

**VITAPAY** (Payment applications) - User-facing payment tools (planned)
- Mobile wallet (React Native) for consumers
- Payment gateway API (Go) for merchants  
- Merchant dashboard (Next.js) for analytics
- E-commerce plugins (WordPress, Shopify)

## Critical Architecture: Monorepo with Nested Go Module Structure

**⚠️ MOST IMPORTANT**: Go module root is at `/Blockchain Project/go.mod`, NOT in subdirectories

```
/Blockchain Project/              ← Go workspace root (go.mod: module vitacoin)
├── go.mod                        ← module: vitacoin (NOT github.com/...)
├── vitacoin/
│   ├── vitacoin/                 ← Blockchain code
│   │   ├── x/vitacoin/           ← Custom module (5,000+ LOC)
│   │   │   ├── keeper/           ← State management (2,500+ LOC)
│   │   │   │   ├── keeper.go, msg_server.go, fees.go, treasury.go
│   │   │   │   └── *_test.go     ← 1,900+ LOC tests
│   │   │   └── types/            ← Generated + custom types
│   │   ├── app/app.go            ← Application wiring (658 lines)
│   │   ├── cmd/vitacoind/        ← CLI binary entry point
│   │   ├── proto/                ← Protocol buffer definitions
│   │   └── build/vitacoind       ← Compiled binary
│   ├── vitapay/                  ← Payment network apps (planned)
│   └── docs/                     ← Comprehensive documentation
```

### Import Path Pattern (Critical!)
```go
// ✅ CORRECT - Relative to go.mod root
import "vitacoin/vitacoin/vitacoin/x/vitacoin/types"
import "vitacoin/vitacoin/vitacoin/x/vitacoin/keeper"

// ❌ WRONG - Using github path when go.mod says "module vitacoin"
import "github.com/vitacoin/vitacoin/x/vitacoin/types"
```

## Build System & Development Workflow

**⚠️ CRITICAL**: Always use Makefile from `vitacoin/vitacoin/` - it handles the monorepo structure

The Makefile navigates to project root before building:
```makefile
PROJECT_ROOT := $(shell cd ../.. && pwd)  # Goes to /Blockchain Project/
# Then: cd "$(PROJECT_ROOT)" && go build ...
```

### Essential Commands (from `vitacoin/vitacoin/`)
```bash
# Build blockchain binary
make build              # Output: build/vitacoind

# Regenerate protobuf code (AFTER editing .proto files)
make proto-gen          # Runs scripts/protocgen.sh from project root

# Run tests (1,900+ LOC test suite)
make test               # Unit tests with -v
make test-race          # Race condition detection
make test-cover         # Coverage report → coverage.html

# Code quality
make lint               # golangci-lint with 20+ linters
make format             # gofmt + goimports

# Dependencies
make go-mod-tidy        # Clean up go.mod (runs from PROJECT_ROOT)
```

### Path Handling
Project path has a space: `/Users/vishwasverma/Downloads/Blockchain Project/`
Makefile quotes paths correctly - preserve this when modifying!

## Protocol Buffers Architecture

**Proto → Go Code Generation Pipeline**:
1. Edit: `proto/vitacoin/v1/*.proto` (genesis.proto, params.proto, query.proto, tx.proto)
2. Generate: `make proto-gen` → runs `scripts/protocgen.sh` from project root
3. Output: `.pb.go` files in `x/vitacoin/types/` (16,855+ LOC generated)
4. **🚫 NEVER edit `.pb.go` files** - they're regenerated and overwritten

### Proto File Organization
```
proto/vitacoin/v1/
├── genesis.proto      # Initial chain state
├── params.proto       # Module parameters (fees, limits)
├── query.proto        # Read queries (gRPC)
└── tx.proto          # Transaction messages

Generated types → x/vitacoin/types/*.pb.go (read-only!)
Custom logic → x/vitacoin/types/{validation.go, msgs.go, events.go}
```

### Adding New Messages (Example: MsgCreatePayment)
1. Add to `proto/vitacoin/v1/tx.proto`:
   ```protobuf
   message MsgCreatePayment {
     string creator = 1;
     string amount = 2;
     string recipient = 3;
   }
   ```
2. Run `make proto-gen` (generates types/*.pb.go)
3. Add handler in `keeper/msg_server.go`:
   ```go
   func (k msgServer) CreatePayment(ctx context.Context, msg *types.MsgCreatePayment) (*types.MsgCreatePaymentResponse, error)
   ```
4. Add validation in `types/validation.go` (custom, not generated)

## Architecture Decisions & "Why" Behind the Structure

### Why Monorepo with Nested Structure?
The unusual `/Blockchain Project/go.mod` → `vitacoin/vitacoin/` nesting exists because:
1. **Historical evolution**: Project started as exploration, grew into ecosystem
2. **Separation of concerns**: `vitacoin/` folder contains BOTH blockchain AND payment apps
3. **Shared dependencies**: Single go.mod simplifies dependency management initially
4. **Future refactoring**: Can be split into separate repos once stable

**Tradeoff**: Complex import paths NOW for easier development INITIALLY. Once VITACOIN stabilizes (Phase 6+), recommend splitting into separate repos.

### Why Custom Module Instead of Using Existing Cosmos Chains?
VITACOIN implements a custom `x/vitacoin` module instead of forking Osmosis or using vanilla Cosmos because:
1. **E-commerce focus**: Need merchant registration, payment escrow, refunds - unique to payment use case
2. **Transparent fee model**: 50/25/25 split (validators/burn/treasury) requires custom fee distribution logic
3. **User experience**: Want to hide blockchain complexity behind VITAPAY apps - needs custom wallet/payment abstractions
4. **Regulatory positioning**: Treasury system enables governance-controlled compliance budget

### Why Cosmos SDK v0.50.3 Specifically?
- **Modern architecture**: v0.50+ uses new store service pattern (not legacy KVStore)
- **Stability**: v0.50.3 is stable (not bleeding edge v0.51+)
- **IBC compatibility**: Proven cross-chain communication
- **CometBFT integration**: Latest consensus engine (formerly Tendermint)

**Tradeoff**: More complex keeper patterns (sdk.Context vs context.Context) but better long-term maintainability.

### Why Both E-Commerce Features AND Traditional DeFi?
Phase 2 focuses on **payment-specific features** (merchants, payments) while keeping **standard staking/governance** for long-term ecosystem health:
- **Short-term**: Merchant adoption drives initial usage
- **Long-term**: Staking APR + governance control attracts token holders
- **Network effects**: Payment volume generates fees → higher validator rewards → more security

## Cosmos SDK Module Architecture (v0.50.3)

**VITACOIN Custom Module** (`x/vitacoin/`) - Purpose-built for e-commerce payments:

### Directory Structure & Purpose
```
x/vitacoin/
├── keeper/                    # State management (3,190+ LOC) ⭐ CORE LOGIC HERE
│   ├── keeper.go             # Main keeper (764 LOC): CRUD for Merchants, Payments, Vaults
│   ├── msg_server.go         # Transaction handlers (705 LOC): MsgCreatePayment, etc.
│   ├── grpc_query.go         # Query handlers (193 LOC): QueryParams, QueryMerchant
│   ├── params.go             # Parameter getters/setters (243 LOC)
│   ├── invariants.go         # State validation (308 LOC): 5 invariants
│   └── msg_server_validation.go  # Advanced validation (272 LOC)
│
├── types/                     # Data structures (16,855+ LOC, mostly generated)
│   ├── *.pb.go               # 🚫 GENERATED - Don't edit!
│   ├── codec.go              # Protobuf registration
│   ├── keys.go               # Store key prefixes (VaultPrefix, PoolPrefix)
│   ├── errors.go             # Custom error types
│   ├── events.go             # Event constants (15+ event types)
│   ├── validation.go         # Custom validation logic
│   └── msgs.go               # Message interface implementations
│
├── module.go                  # AppModule interface (lifecycle hooks)
├── genesis.go                 # InitGenesis/ExportGenesis
└── client/cli/               # CLI commands (future)
```

### Keeper State Management Pattern (CRITICAL!)
Every keeper method follows this exact flow:
```go
func (k Keeper) CreateVault(ctx context.Context, msg *types.MsgCreateVault) error {
    // 1. Validation (msg.ValidateBasic() already called by SDK)
    if msg.Amount.IsZero() {
        return types.ErrInvalidAmount.Wrap("amount must be positive")
    }
    
    // 2. Load state from IAVL store
    store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
    
    // 3. Business logic
    vault := types.Vault{
        Id:      generateID(ctx),
        Owner:   msg.Creator,
        Amount:  msg.Amount,
        // ...
    }
    
    // 4. Save state (marshaled with protobuf)
    bz := k.cdc.MustMarshal(&vault)
    store.Set(types.VaultKey(vault.Id), bz)
    
    // 5. Emit events (for indexers/explorers)
    sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(
        sdk.NewEvent(types.EventTypeCreateVault,
            sdk.NewAttribute(types.AttributeKeyVaultID, vault.Id),
        ),
    )
    
    return nil
}
```

### Store Key Organization (`types/keys.go`)
```go
const (
    ModuleName = "vitacoin"
    
    // IAVL tree prefixes (single byte for efficiency)
    MerchantPrefix    = "m:"   // Merchant data by address
    PaymentPrefix     = "p:"   // Payment data by ID
    VaultPrefix       = "v:"   // Time-locked vaults
    PoolPrefix        = "pool:" // Reward pools
    ParamsPrefix      = "params:" // Module parameters
)

// Key construction helpers
func VaultKey(id string) []byte {
    return append([]byte(VaultPrefix), []byte(id)...)
}
```

### Message Handler Examples (msg_server.go)
```go
type msgServer struct {
    Keeper  // Embedded keeper for state access
}

// Transaction handler - called when user submits MsgCreatePayment
func (m msgServer) CreatePayment(ctx context.Context, msg *types.MsgCreatePayment) (*types.MsgCreatePaymentResponse, error) {
    // Validation already done by msg.ValidateBasic()
    
    // Use keeper methods to update state
    payment := types.Payment{
        Id:          generatePaymentID(ctx),
        From:        msg.Creator,
        To:          msg.Recipient,
        Amount:      msg.Amount,
        Status:      types.PaymentStatus_PENDING,
        CreatedAt:   sdk.UnwrapSDKContext(ctx).BlockTime(),
    }
    
    if err := m.SetPayment(ctx, payment); err != nil {
        return nil, err
    }
    
    return &types.MsgCreatePaymentResponse{PaymentId: payment.Id}, nil
}
```

## Critical Data Flows & Integration Points

### Payment Processing Flow (End-to-End)
Understanding this flow is key to working with VITACOIN's payment features:

```
Customer initiates payment
  ↓
1. MsgCreatePayment submitted to chain
  ↓
2. Validation Layer (msg_server_validation.go)
   • Validates addresses (bech32 format)
   • Checks merchant is active
   • Validates amount (positive, non-zero)
   • Sanitizes memo (removes control chars)
  ↓
3. Escrow Funds (fees.go) - Phase 3
   • Transfer from customer → vitacoin module account
   • BankKeeper.SendCoinsFromAccountToModule()
   • Event: EventTypePaymentCreated
  ↓
4. Store Payment (keeper.go)
   • Generate unique payment ID
   • Set status = PENDING
   • Store in IAVL tree with key: PaymentPrefix + ID
   • Return payment ID to customer

[Time passes - goods/services delivered]

Merchant completes payment
  ↓
5. MsgCompletePayment submitted
  ↓
6. Settlement (fees.go) - Phase 3
   • Verify payment exists, status = PENDING
   • Calculate protocol fee (0.1% with min/max caps)
   • Apply merchant tier discount (0%/25%/50%/75%)
   • Send net amount to merchant
   • Accumulate protocol fee for later distribution
   • Event: EventTypePaymentSettled
  ↓
7. Update Payment Status (keeper.go)
   • Set status = COMPLETED
   • Set completion height
   • Update merchant statistics (total volume, tx count)

End of Block (EndBlocker)
  ↓
8. Fee Distribution (fees.go) - Phase 3
   • Get block fee accumulator
   • Split fees: 50% → FeeCollector (validators)
   •             25% → Burn (if under cap)
   •             25% → Treasury (governance-controlled)
   • Update statistics
   • Create supply snapshot (daily)
   • Event: EventTypeFeeDistribution
```

**Key Integration Points**:
- **BankKeeper**: All token transfers (payments, stakes, fees)
- **AccountKeeper**: Module accounts (vitacoin module, FeeCollector)
- **Distribution Module**: Validator fee distribution (via FeeCollector)
- **Governance Module**: Treasury spending proposals (Phase 3)

### Module Account Architecture
VITACOIN uses Cosmos SDK module accounts for secure fund management:

```
vitacoin module account (vitacoinModuleAccount)
├── Holds escrowed payment funds (PENDING payments)
├── Holds merchant stakes (for tier calculations)
├── Holds vault deposits (time-locked tokens)
├── Permissions: Minter, Burner (for fee burning)
└── Authority: vitacoin keeper only (no external access)

Treasury account (managed by x/gov)
├── Receives 25% of all protocol fees
├── Spending requires governance proposal
├── Used for: development, grants, partnerships
└── Authority: governance module only

FeeCollector account (x/distribution)
├── Receives 50% of all protocol fees
├── Auto-distributed to validators each block
└── Authority: distribution module only
```

### State Storage Layout
Understanding the IAVL tree structure helps with queries and debugging:

```
IAVL Tree Root
│
├── vitacoin/ (ModuleName)
│   ├── 0x00 → Params (single key-value)
│   ├── 0x01 → Merchants/
│   │   ├── [address1] → Merchant{...}
│   │   ├── [address2] → Merchant{...}
│   │   └── ...
│   ├── 0x02 → Payments/
│   │   ├── [paymentID1] → Payment{...}
│   │   ├── [paymentID2] → Payment{...}
│   │   └── ...
│   ├── 0x03 → Vaults/
│   ├── 0x04 → RewardPools/
│   ├── 0x05 → NextPaymentIDCounter
│   ├── 0x06 → BlockFeeAccumulator (Phase 3)
│   ├── 0x07 → FeeStatistics (Phase 3)
│   ├── 0x08 → BurnStatistics (Phase 3)
│   ├── 0x09 → SupplySnapshots/ (Phase 3)
│   └── 0x0A → TreasurySpending/ (Phase 3)
```

**Query Optimization**: Prefixes enable efficient iteration:
```go
// Get all merchants
store.Iterator([]byte(MerchantPrefix), sdk.PrefixEndBytes([]byte(MerchantPrefix)))

// Get single merchant by address
store.Get(append([]byte(MerchantPrefix), []byte(address)...))
```

## Common Development Tasks

### Adding a New Feature (Complete Workflow)

**Example: Add "Refund Payment" feature**

1. **Define Proto Message** (`proto/vitacoin/v1/tx.proto`):
   ```protobuf
   message MsgRefundPayment {
     string creator = 1;
     string payment_id = 2;
     string reason = 3;
   }
   message MsgRefundPaymentResponse {
     bool success = 1;
   }
   ```

2. **Generate Code**: `cd vitacoin/vitacoin && make proto-gen`
   - Creates: `types/tx.pb.go` with MsgRefundPayment struct

3. **Implement Keeper Logic** (`keeper/msg_server.go`):
   ```go
   func (m msgServer) RefundPayment(ctx context.Context, msg *types.MsgRefundPayment) (*types.MsgRefundPaymentResponse, error) {
       // 1. Load payment
       payment, err := m.GetPayment(ctx, msg.PaymentId)
       if err != nil {
           return nil, types.ErrPaymentNotFound.Wrapf("payment %s", msg.PaymentId)
       }
       
       // 2. Validate refund is possible
       if payment.Status != types.PaymentStatus_COMPLETED {
           return nil, types.ErrInvalidPaymentStatus.Wrap("only completed payments can be refunded")
       }
       
       // 3. Update state
       payment.Status = types.PaymentStatus_REFUNDED
       payment.RefundReason = msg.Reason
       if err := m.SetPayment(ctx, payment); err != nil {
           return nil, err
       }
       
       // 4. Emit event
       sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(
           sdk.NewEvent(types.EventTypeRefundPayment,
               sdk.NewAttribute(types.AttributeKeyPaymentID, msg.PaymentId),
           ),
       )
       
       return &types.MsgRefundPaymentResponse{Success: true}, nil
   }
   ```

4. **Add Validation** (`types/validation.go`):
   ```go
   func (msg *MsgRefundPayment) ValidateBasic() error {
       if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
           return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address: %s", err)
       }
       if msg.PaymentId == "" {
           return ErrInvalidPaymentID.Wrap("payment_id cannot be empty")
       }
       if len(msg.Reason) > 500 {
           return ErrInvalidReason.Wrap("reason exceeds 500 characters")
       }
       return nil
   }
   ```

5. **Write Tests** (`keeper/msg_server_test.go`):
   ```go
   func (suite *KeeperTestSuite) TestRefundPayment() {
       // Setup: Create a completed payment
       payment := types.Payment{
           Id: "pay123",
           Status: types.PaymentStatus_COMPLETED,
           // ...
       }
       suite.keeper.SetPayment(suite.ctx, payment)
       
       // Execute: Refund the payment
       msg := &types.MsgRefundPayment{
           Creator: "creator_address",
           PaymentId: "pay123",
           Reason: "Customer request",
       }
       res, err := suite.msgServer.RefundPayment(suite.ctx, msg)
       
       // Assert: Payment status changed
       suite.Require().NoError(err)
       suite.Require().True(res.Success)
       
       updatedPayment, _ := suite.keeper.GetPayment(suite.ctx, "pay123")
       suite.Require().Equal(types.PaymentStatus_REFUNDED, updatedPayment.Status)
   }
   ```

6. **Run Tests**: `make test`

### App.go Wiring (⚠️ Touch with Caution!)

`app/app.go` (658 lines) is the central nervous system - module initialization order matters:

```go
// ModuleBasics registration (line ~96)
ModuleBasics = module.NewBasicManager(
    auth.AppModuleBasic{},
    bank.AppModuleBasic{},
    staking.AppModuleBasic{},
    // ...
    vitacoin.AppModuleBasic{},  // Our custom module
)

// Module account permissions (line ~117)
maccPerms = map[string][]string{
    vitacointypes.ModuleName: {authtypes.Minter, authtypes.Burner},  // Allows minting/burning
}

// Keeper initialization order (lines ~300-400) - dependencies matter!
app.AccountKeeper = authkeeper.NewAccountKeeper(...)      // First
app.BankKeeper = bankkeeper.NewBaseKeeper(...)            // Needs AccountKeeper
app.StakingKeeper = stakingkeeper.NewKeeper(...)          // Needs BankKeeper
app.VitacoinKeeper = vitacoinkeeper.NewKeeper(...)        // Needs BankKeeper

// Module execution order (lines ~500-550)
app.mm = module.NewManager(
    auth.NewAppModule(...),
    bank.NewAppModule(...),
    vitacoin.NewAppModule(app.VitacoinKeeper),  // Runs BeginBlock/EndBlock
)

// BeginBlockers: Auth → Bank → Staking → Vitacoin
// EndBlockers: Vitacoin → Staking → Bank → Auth
```

**🚫 Don't modify app.go unless**: Adding new Cosmos SDK modules or changing inter-module dependencies

## Testing Strategy (4,000+ LOC Test Suite)

### Test Organization
```
x/vitacoin/
├── keeper/
│   ├── keeper.go              # Implementation
│   ├── keeper_test.go         # Unit tests for CRUD operations
│   ├── msg_server.go          # Message handlers
│   ├── msg_server_test.go     # Transaction tests
│   └── params_test.go         # Parameter tests
│
├── types/
│   ├── validation_test.go     # Input validation (500+ LOC)
│   └── validation_bench_test.go  # Performance benchmarks
│
├── module_test.go             # Module lifecycle tests (500+ LOC)
└── integration_test.go        # End-to-end tests (874 LOC, needs fixes)
```

### Running Tests
```bash
cd vitacoin/vitacoin

# All unit tests (fast, ~5 seconds)
make test

# With race detection (slower, catches concurrency bugs)
make test-race

# Coverage report (generates coverage.html)
make test-cover

# Specific package
cd /Users/vishwasverma/Downloads/Blockchain\ Project
go test -v ./vitacoin/vitacoin/x/vitacoin/keeper/...

# Single test
go test -v ./vitacoin/vitacoin/x/vitacoin/keeper -run TestKeeperTestSuite/TestSetGetMerchant
```

### Test Suite Pattern (Cosmos SDK Standard)
```go
// keeper_test.go
type KeeperTestSuite struct {
    suite.Suite
    
    ctx       sdk.Context
    keeper    keeper.Keeper
    msgServer types.MsgServer
    // ... other dependencies
}

func (suite *KeeperTestSuite) SetupTest() {
    // Initialize test environment
    suite.ctx = testutil.DefaultContext(...)
    suite.keeper = keeper.NewKeeper(...)
    suite.msgServer = keeper.NewMsgServerImpl(suite.keeper)
}

func (suite *KeeperTestSuite) TestCreateVault() {
    // Arrange
    msg := &types.MsgCreateVault{
        Creator: "vita1...",
        Amount:  sdk.NewInt(1000),
    }
    
    // Act
    res, err := suite.msgServer.CreateVault(suite.ctx, msg)
    
    // Assert
    suite.Require().NoError(err)
    suite.Require().NotNil(res)
    
    vault, err := suite.keeper.GetVault(suite.ctx, res.VaultId)
    suite.Require().NoError(err)
    suite.Require().Equal(msg.Amount, vault.Amount)
}

func TestKeeperTestSuite(t *testing.T) {
    suite.Run(t, new(KeeperTestSuite))  // Run all tests in suite
}
```

### Current Test Status (Phase 2 at 90%)
- ✅ CRUD operations: 100% passing
- ✅ Types validation: 90% passing (27/30 tests)
- ⚠️ Message handlers: ~75% passing (business logic tuning needed)
- ❌ Integration tests: Compilation errors (874 LOC written)

## Common Pitfalls & Solutions (Learned from This Project)

### 1. Import Path Confusion
**Symptom**: `package not found` or `cannot find module providing package`
**Root Cause**: go.mod says `module vitacoin` but import paths need full nested structure
**Fix**:
```go
// ❌ WRONG - Missing nested vitacoin/vitacoin path
import "vitacoin/x/vitacoin/types"

// ✅ CORRECT - Full path from go.mod root
import "vitacoin/vitacoin/vitacoin/x/vitacoin/types"
```
**Why this works**: Go looks for `vitacoin/` (module name) then navigates nested `vitacoin/vitacoin/` structure.

### 2. Build Failures After Dependency Changes
**Symptom**: `module requires Go 1.22` or version conflicts
**Fix**:
```bash
cd "/Users/vishwasverma/Downloads/Blockchain Project"
go mod tidy          # Clean dependencies
go mod download      # Fetch missing packages
cd vitacoin/vitacoin && make build
```

### 3. Proto Generation Issues
**Symptom**: `.pb.go` files not updating or missing
**Checklist**:
- [ ] Run from correct directory: `cd vitacoin/vitacoin && make proto-gen`
- [ ] Check `buf` is installed: `buf --version`
- [ ] Verify proto syntax: `buf lint`
- [ ] Don't manually edit `.pb.go` files - they're regenerated

**Common Error**: `protoc-gen-gocosmos not found`
```bash
# Install missing protoc plugins
go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@latest
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@latest
```

### 4. Context Type Confusion (Cosmos SDK v0.50+ Pattern)
**Symptom**: Tests fail with `context deadline exceeded` or `cannot call .EventManager() on context.Context`
**Root Cause**: v0.50+ uses Go's standard `context.Context` for keeper signatures but SDK features need `sdk.Context`
**Fix**: Use `sdk.UnwrapSDKContext(ctx)` to access SDK features:
```go
// In keeper methods accepting context.Context
func (k Keeper) CreatePayment(ctx context.Context, msg *types.MsgCreatePayment) error {
    // ❌ WRONG - context.Context doesn't have BlockTime()
    blockTime := ctx.BlockTime()
    
    // ✅ CORRECT - Unwrap to get sdk.Context
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    blockTime := sdkCtx.BlockTime()
    eventManager := sdkCtx.EventManager()
}
```
**Why this pattern**: Cosmos SDK v0.50+ moved to standard `context.Context` for better Go ecosystem compatibility, but many SDK features still need the rich `sdk.Context` type.

### 5. "Invalid Memory Address" in Tests
**Symptom**: `panic: runtime error: invalid memory address`
**Root Cause**: Keeper dependencies not initialized in test setup
**Fix**: Ensure `SetupTest()` initializes all keepers:
```go
func (suite *KeeperTestSuite) SetupTest() {
    // Must initialize: storeService, cdc, authority, bankKeeper, accountKeeper
    suite.keeper = keeper.NewKeeper(
        suite.storeService,
        suite.cdc,
        suite.authority,
        suite.bankKeeper,
        suite.accountKeeper,
    )
}
```

### 6. Path Spaces Breaking Scripts
**Project Path**: `/Users/vishwasverma/Downloads/Blockchain Project/` (note space!)
**Always quote paths**:
```bash
# ❌ WRONG
cd /Users/vishwasverma/Downloads/Blockchain Project

# ✅ CORRECT
cd "/Users/vishwasverma/Downloads/Blockchain Project"
```
Makefile handles this correctly - don't modify `PROJECT_ROOT` variable.

## VITACOIN-Specific Conventions (How We Differ from Standard Cosmos Projects)

### Why These Conventions Matter
VITACOIN prioritizes **payment use cases** over general-purpose blockchain features. These conventions reflect that focus.

### Naming Standards
| Type | Pattern | Example |
|------|---------|---------|
| Module name | lowercase | `vitacoin` |
| Store prefix | lowercase + colon | `"m:"`, `"vault:"` |
| Binary name | lowercase + d suffix | `vitacoind` (daemon) |
| Message types | `Msg*` prefix | `MsgCreatePayment` |
| Query types | `Query*` prefix | `QueryParamsRequest` |
| Event types | snake_case | `"create_vault"` |
| Error codes | 1000-1999 range | `ErrInvalidVault = 1101` |

### Error Handling Pattern
**Define in `types/errors.go`**:
```go
var (
    // 1100-1199: Vault errors
    ErrInvalidVault       = errorsmod.Register(ModuleName, 1101, "invalid vault")
    ErrVaultNotFound      = errorsmod.Register(ModuleName, 1102, "vault not found")
    
    // 1200-1299: Payment errors
    ErrInvalidPayment     = errorsmod.Register(ModuleName, 1201, "invalid payment")
    ErrPaymentExpired     = errorsmod.Register(ModuleName, 1202, "payment expired")
    
    // Code ranges reserved:
    // 1000-1099: General errors
    // 1100-1199: Vault operations
    // 1200-1299: Payment operations
    // 1300-1399: Merchant operations
    // 1400-1499: Pool operations
)
```

**Use in keeper methods**:
```go
// Wrap errors with context
if !found {
    return types.ErrVaultNotFound.Wrapf("vault ID: %s", vaultID)
}

// Chain errors
if err := k.BankKeeper.SendCoins(...); err != nil {
    return errorsmod.Wrap(err, "failed to send coins for vault creation")
}
```

### Event Architecture (15+ Types)
**Define in `types/events.go`**:
```go
const (
    EventTypeCreateVault     = "create_vault"
    EventTypeWithdrawVault   = "withdraw_vault"
    EventTypeCreatePayment   = "create_payment"
    
    AttributeKeyVaultID      = "vault_id"
    AttributeKeyOwner        = "owner"
    AttributeKeyAmount       = "amount"
    AttributeKeyPaymentID    = "payment_id"
)
```

**Emit in keeper methods**:
```go
sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(
    sdk.NewEvent(
        types.EventTypeCreateVault,
        sdk.NewAttribute(types.AttributeKeyVaultID, vault.Id),
        sdk.NewAttribute(types.AttributeKeyOwner, vault.Owner),
        sdk.NewAttribute(types.AttributeKeyAmount, vault.Amount.String()),
    ),
)
```

**Why events matter**: Block explorers, indexers, and dApps subscribe to events for real-time updates.

### Fee Distribution Philosophy (Why 50/25/25?)
**Architecture** (Phase 3 - partially implemented):
```go
// On every transaction in EndBlock
totalFees := collectTransactionFees(ctx)  // 0.1% of all tx amounts

// Split transparently and on-chain
validatorShare := totalFees.MulInt64(50).QuoInt64(100)  // 50%
burnAmount := totalFees.MulInt64(25).QuoInt64(100)      // 25%
treasuryAmount := totalFees.MulInt64(25).QuoInt64(100)  // 25%

// Distribute
k.distributionKeeper.AllocateTokensToValidator(ctx, validatorShare)
k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(burnAmount))
k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, govtypes.ModuleName, sdk.NewCoins(treasuryAmount))

// Emit transparency event
emitFeeDistributionEvent(ctx, validatorShare, burnAmount, treasuryAmount)
```

**Why This Split?**
- **50% Validators**: Competitive with traditional staking APRs (5-7%)
- **25% Burn**: Creates deflationary pressure (balances 3-10% inflation from staking rewards)
- **25% Treasury**: Funds VITAPAY development, merchant onboarding, compliance (governance-controlled)

**Key Differentiator**: All fees tracked on-chain and verifiable (unlike PayPal/Stripe's hidden fee structures). Merchants can audit where their fees go in real-time using block explorers.

## VITACOIN Token Economics & Features

### Token Specifications
| Property | Value |
|----------|-------|
| Symbol | VITA |
| Total Supply | 1,000,000,000 (1 billion) |
| Decimals | 18 |
| Base Unit | uvita (1 VITA = 10^18 uvita) |
| Chain ID (mainnet) | `vitacoin-1` |
| Chain ID (testnet) | `vitacoin-local` |
| Block Time | ~6 seconds (CometBFT) |
| Finality | Instant (no reorgs) |

### Unique Features (Implementation Phases)

#### Phase 2-3: Core Features
1. **Time-Locked Vaults** - Lock VITA for higher rewards
   - 1 month: 1.0x multiplier
   - 3 months: 1.2x multiplier
   - 6 months: 1.5x multiplier
   - 12 months: 2.0x multiplier
   - Keeper methods: `CreateVault`, `ExtendVault`, `WithdrawVault`

2. **Transparent Fee Distribution** (0.1% per transaction)
   - 50% → Validators (network security)
   - 25% → Burned (deflationary pressure)
   - 25% → Treasury (ecosystem development)
   - **All tracked on-chain** - verifiable by block explorers

3. **Merchant Payment System**
   - Register merchants with `MsgRegisterMerchant`
   - Create payments with `MsgCreatePayment`
   - Settle instantly (5-second finality)
   - Lower fees than traditional processors (0.1% vs 2-3%)

#### Phase 4-6: Advanced Features
4. **Liquid Staking** - Maintain liquidity while staking
   - Stake VITA → Receive stVITA (derivative token)
   - stVITA accrues rewards automatically
   - Trade or use stVITA in DeFi while earning staking rewards

5. **Custom Reward Pools** - Flexible incentive programs
   - Create pools with custom rewards
   - Stake any token to earn rewards
   - Use for liquidity mining, loyalty programs, airdrops

### Token Distribution
```
Total: 1,000,000,000 VITA
├── 400M (40%) - Staking Rewards (vested over 10 years)
├── 300M (30%) - Genesis Allocation (team, advisors, early supporters)
├── 200M (20%) - Ecosystem Development (grants, partnerships, VITAPAY)
└── 100M (10%) - Governance Reserve (community-controlled treasury)
```

### Staking Parameters
```go
// In types/params.go (default values)
params := Params{
    TargetBondedRatio:        "0.67",    // 67% of supply should be staked
    UnbondingPeriod:          21 * 24 * time.Hour,  // 21 days
    MinInflation:             "0.03",    // 3% minimum annual inflation
    MaxInflation:             "0.10",    // 10% maximum annual inflation
    InflationAdjustmentRate:  "0.13",    // Monthly recalculation
    ExpectedAPR:              "0.07",    // ~7% expected return for stakers
}
```

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

## Current Development Status (October 2025)

### Phase 1: Foundation Setup ✅ Complete (Oct 16, 2025)
- ✅ Go 1.21 environment with Cosmos SDK v0.50.3
- ✅ Protocol buffer definitions (5 proto files)
- ✅ Proto generation pipeline → 16,855+ LOC generated
- ✅ Module structure (keeper/, types/, module.go)
- ✅ app.go wiring (658 lines, all Cosmos SDK modules)
- ✅ Binary compilation (vitacoind builds successfully)
- ✅ CLI commands (init, export-genesis, start)
- ✅ Monorepo structure with correct import paths

### Phase 2: E-Commerce Payment Module 🚧 98% Complete (Oct 17, 2025)

**✅ Core Implementation Complete**:
- **Keeper Package** (2,500+ LOC)
  - keeper.go: CRUD operations (795 LOC)
  - msg_server.go: Transaction handlers (705 LOC)
  - msg_server_validation.go: Input validation (272 LOC)
  - grpc_query.go: Query endpoints (193 LOC)
  - params.go: Parameter management (243 LOC)
  - invariants.go: State validation (308 LOC, 5 invariants)
  
- **Business Features** (80+ functions):
  - Merchant System: Registration, tiers (Bronze/Silver/Gold/Platinum), fee discounts
  - Payment Processing: Create, complete, refund with escrow
  - Time-Locked Vaults: Lock tokens with reward multipliers
  - Reward Pools: Merchant loyalty programs

- **Testing** (1,900+ LOC tests):
  - ✅ CRUD operations: 100% passing
  - ✅ Type validation: 90% passing (27/30 tests)
  - ⚠️ Message handlers: 75% passing (business logic refinement needed)
  - ❌ Integration tests: 874 LOC written but compilation errors

**🎯 Remaining Work (2% - 1 week)**:
1. Fix failing message handler tests (business logic edge cases)
2. Resolve integration test compilation issues
3. Add performance benchmarks

### Phase 3: Fee System & Treasury 🚧 60% Complete (In Progress)

**✅ Completed Tasks**:
- Fee collection & escrow (370 LOC)
- Fee distribution (50/25/25 split) (200 LOC)
- Burn mechanism with cap (290 LOC)
- Treasury system (1,450 LOC, 9 gRPC queries)
- Parameter definitions (150 LOC)
- Security safeguards

**⏳ Remaining Tasks**:
- Additional query endpoints
- Comprehensive testing suite
- Documentation & events reference
- Genesis & vesting setup

### Upcoming Phases

**Phase 4**: Staking System (4 weeks)
**Phase 5**: Governance (3 weeks)
**Phase 6**: IBC Integration (3 weeks)

**Target Mainnet**: August 2026

## VITAPAY Payment Network (Planning Phase)

### Current Status (Phase 0 - 10% Complete)
```
vitapay/
├── mobile-wallet/          # 📱 Consumer wallet app (React Native)
│   ├── README.md          ✅ Specs written
│   └── src/               ⏳ Not started
│
├── payment-gateway/        # 🌐 Merchant API (Go)
│   ├── README.md          ✅ Specs written
│   └── internal/          ⏳ Not started
│
├── merchant-dashboard/     # � Web portal (Next.js)
│   ├── README.md          ✅ Specs written
│   └── src/               ⏳ Not started
│
└── shared/
    └── vitacoin-client/   # 🔧 TypeScript SDK for blockchain
        └── src/           ⏳ Not started (depends on VITACOIN mainnet)
```

### Technology Stack
| Component | Framework | Language | Status |
|-----------|-----------|----------|--------|
| Mobile Wallet | React Native | TypeScript | Planning |
| Payment Gateway | Gin/Echo | Go | Planning |
| Merchant Dashboard | Next.js 14 | TypeScript | Planning |
| Blockchain Client | CosmJS | TypeScript | Planning |
| Database | PostgreSQL | SQL | Planning |

### Integration Architecture
```
┌─────────────────────────────────────────────────────────┐
│                 VITAPAY Applications                     │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Mobile Wallet (iOS/Android) ──┐                       │
│                                 │                        │
│  Merchant Dashboard (Web) ──────┼──► vitacoin-client   │
│                                 │    TypeScript SDK     │
│  Payment Gateway (API) ─────────┘    (CosmJS wrapper)  │
│                                           │              │
│                                           ▼              │
│                                  ┌──────────────────┐   │
│                                  │ VITACOIN Node    │   │
│                                  │ (gRPC/REST)      │   │
│                                  └──────────────────┘   │
│                                           │              │
│                                           ▼              │
│                                  ┌──────────────────┐   │
│                                  │ VITACOIN Chain   │   │
│                                  │ (Cosmos SDK)     │   │
│                                  └──────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

### Development Timeline
- **Q1 2026**: VITACOIN testnet launch (blocks VITAPAY development)
- **Q2 2026**: Mobile wallet MVP (8 weeks) + Payment gateway API (6 weeks)
- **Q3 2026**: Security audits, beta testing
- **Q4 2026**: E-commerce plugins (Shopify, WooCommerce)
- **Q1 2027**: Merchant dashboard, analytics
- **Q2 2027**: Public launch

**Note to AI Agents**: VITAPAY code doesn't exist yet - focus on VITACOIN blockchain first!

## Quick Command Reference

### Daily Development Workflow
```bash
# 1. Navigate to blockchain directory
cd "/Users/vishwasverma/Downloads/Blockchain Project/vitacoin/vitacoin"

# 2. After editing .proto files
make proto-gen                    # Regenerates types/*.pb.go

# 3. After editing .go files
make build                        # Compiles to build/vitacoind

# 4. Run tests
make test                         # All unit tests
make test-race                    # Race detection
make test-cover                   # Coverage report

# 5. Code quality
make lint                         # Run all linters
make format                       # Format Go code
```

### Testing the Binary
```bash
# Help and version
./build/vitacoind --help
./build/vitacoind version

# Initialize local testnet
./build/vitacoind init mynode --chain-id vitacoin-local --home ./testnet

# Export default genesis
./build/vitacoind export-genesis > genesis.json

# Start node (future - after genesis setup)
./build/vitacoind start --home ./testnet
```

### Debugging
```bash
# Run single test with verbose output
cd "/Users/vishwasverma/Downloads/Blockchain Project"
go test -v ./vitacoin/vitacoin/x/vitacoin/keeper -run TestSetGetMerchant

# Check module wiring
grep -r "vitacoin" vitacoin/vitacoin/app/app.go

# Verify import paths
go list -m all | grep vitacoin

# Check generated proto files
ls -lh vitacoin/vitacoin/x/vitacoin/types/*.pb.go
```

## Documentation Map

### Start Here (Onboarding)
1. **`vitacoin/README.md`** - Ecosystem overview (read this first!)
2. **`vitacoin/docs/FOLDER_STRUCTURE.md`** - Navigate the monorepo
3. **`vitacoin/vitacoin/README.md`** - Blockchain-specific guide

### Architecture & Design
- **`docs/architecture/ARCHITECTURE.md`** - System design (module architecture, network topology)
- **`docs/architecture/DEV_SETUP.md`** - Development environment setup
- **`docs/VITACOIN.md`** - Cryptocurrency deep dive (token economics, features)

### Development Guides
- **`docs/development/GETTING_STARTED.md`** - Developer onboarding
- **`docs/development/QUICK_REFERENCE.md`** - Command cheatsheet
- **`docs/development/PHASE1_BUILD_SUCCESS.md`** - Phase 1 completion report
- **`docs/development/PHASE2_VERIFICATION_REPORT.md`** - Current status

### Task Tracking
- **`vitacoin/vitacoin/TODO.md`** - Blockchain roadmap (1,617 lines, 20 phases)
- **`vitacoin/vitapay/TODO.md`** - Payment network roadmap (677 lines, 13 phases)
- **`docs/project/DEVELOPMENT_ROADMAP.md`** - High-level timeline

### VITAPAY (Future)
- **`docs/project/VITAPAY.md`** - Payment network overview
- **`docs/project/MOBILE_APP.md`** - Wallet app specifications
- **`vitapay/*/README.md`** - Component-specific docs (not yet implemented)

## Key Files to Reference

### When implementing features:
```
keeper/keeper.go          # CRUD patterns (SetMerchant, GetVault examples)
keeper/params.go          # Simple parameter getters/setters
types/validation.go       # Input validation patterns
types/errors.go           # Error definitions and codes
types/events.go          # Event type constants
module.go                # Module lifecycle (BeginBlock/EndBlock)
```

### When debugging build issues:
```
go.mod                   # Dependency versions (at project root!)
Makefile                 # Build commands and PROJECT_ROOT variable
app/app.go               # Module wiring and initialization order
scripts/protocgen.sh     # Proto generation script
```

### When writing tests:
```
keeper/keeper_test.go          # Test suite setup pattern
keeper/msg_server_test.go      # Transaction test examples
types/validation_test.go       # Validation test examples
```

## Working with AI Agents (Guidelines)

### DO:
✅ Use `make build`, `make test`, `make proto-gen` (not raw Go commands)
✅ Follow existing patterns in keeper/keeper.go for CRUD operations
✅ Check types/events.go before adding new event types
✅ Read types/errors.go before defining new errors
✅ Test incrementally - write tests alongside implementation
✅ Emit events after state changes for observability
✅ Use sentinel errors (types.ErrInvalidVault.Wrap(...))
✅ Reference TODO.md for planned features and priorities

### DON'T:
❌ Edit .pb.go files (they're generated from .proto)
❌ Run `go build` directly (use Makefile)
❌ Modify app/app.go unless adding new modules
❌ Change import paths without full monorepo structure
❌ Skip tests ("I'll add them later" never works)
❌ Hardcode values that should be parameters
❌ Ignore the Makefile's PROJECT_ROOT handling

### When Stuck:
1. Check existing similar feature (e.g., Merchant CRUD for new entity)
2. Read Cosmos SDK docs: https://docs.cosmos.network/v0.50
3. Search for patterns: `grep -r "SetMerchant" vitacoin/vitacoin/x/vitacoin/`
4. Review phase completion docs: `docs/development/PHASE*_COMPLETE.md`
5. Check test files for usage examples

## Technology Stack Summary

```yaml
Blockchain:
  Language: Go 1.21+
  Framework: Cosmos SDK v0.50.3
  Consensus: CometBFT (BFT consensus)
  State: IAVL tree (Merkle tree for state proofs)
  RPC: gRPC (port 9090) + REST (port 1317)
  
Protocol Buffers:
  Compiler: protoc
  Extensions: gogoproto (performance optimizations)
  Generation: buf (modern proto tooling)
  
Testing:
  Framework: testify/suite (table-driven tests)
  Coverage: go test -cover
  Race Detection: go test -race
  
Code Quality:
  Linter: golangci-lint (20+ linters)
  Formatter: gofmt + goimports
  
Future (VITAPAY):
  Mobile: React Native (TypeScript)
  Backend: Go + Gin/Echo
  Frontend: Next.js 14 (TypeScript)
  Database: PostgreSQL
  Cache: Redis
```

## Critical Success Factors

1. **Import Paths**: Always use full monorepo path from go.mod root
2. **Makefile Usage**: Never bypass - it handles PROJECT_ROOT correctly
3. **Proto Generation**: Run `make proto-gen` after any .proto changes
4. **Test Coverage**: Write tests alongside implementation, not after
5. **Code Patterns**: Follow existing keeper CRUD patterns exactly
6. **Event Emission**: Emit events for every state change (observability)
7. **Error Handling**: Use wrapped sentinel errors with context
8. **Module Wiring**: Understand app.go initialization order before modifying

## Quick Wins for New AI Agents (Get Productive in 5 Minutes)

### First Time Here? Start with These
1. **Understand the "big picture"**: Read `vitacoin/README.md` (ecosystem overview)
2. **Build the binary**: `cd vitacoin/vitacoin && make build`
3. **Run existing tests**: `make test` (see what works, what doesn't)
4. **Explore keeper patterns**: Open `keeper/keeper.go` - all CRUD operations follow same pattern
5. **Check current priorities**: `vitacoin/vitacoin/TODO.md` Phase 2 section

### Most Common Tasks (With Examples)
**Task**: Add a new query endpoint
**Example**: See `keeper/grpc_query.go` lines 50-70 (QueryMerchant pattern)
**Steps**: 1) Add proto message, 2) make proto-gen, 3) Implement keeper method

**Task**: Add validation to existing message
**Example**: See `types/validation.go` (ValidateBasic pattern)
**Steps**: Add to existing ValidateBasic() method, write test

**Task**: Fix failing test
**Example**: See `keeper/keeper_test.go` for setup pattern
**Steps**: 1) Run single test with -v, 2) Check error message, 3) Fix business logic

**Task**: Add event emission
**Example**: See `types/events.go` for constants, `keeper/msg_server.go` for usage
**Steps**: 1) Define constant, 2) Emit in keeper method

### When Stuck, Check These Files First
- **Import errors?** → Check `go.mod` (module name is "vitacoin", not github path)
- **Proto errors?** → Run `make proto-gen` from `vitacoin/vitacoin/`
- **Build errors?** → Check `Makefile` PROJECT_ROOT variable
- **Test errors?** → Look at similar passing test in same file
- **Business logic?** → Read `docs/architecture/ARCHITECTURE.md` for Phase 2 features

## Next Steps for This Session

1. **Build the project**: `cd vitacoin/vitacoin && make build`
2. **Run tests**: `make test` (expect 75-90% passing)
3. **Read keeper patterns**: `keeper/keeper.go` (795 LOC, SetMerchant/GetMerchant examples)
4. **Check current status**: `docs/development/PHASE2_VERIFICATION_REPORT.md`
5. **Pick a task**: See `vitacoin/vitacoin/TODO.md` Phase 2 section
6. **Ask questions**: Reference existing code patterns before asking

---

**Last Updated**: October 17, 2025  
**VITACOIN Status**: Phase 2 at 98% complete, Phase 3 at 60% complete  
**Next Milestone**: Complete Phase 3 (Fee System & Treasury)  
**Target Mainnet**: August 2026  
**Maintainer**: @vishwas-io / GitHub: vishwas-io/vitacoin
````````

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
