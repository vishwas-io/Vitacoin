# Blockchain Core — Go + Cosmos SDK Standards

## Stack
- Go 1.24
- Cosmos SDK v0.50.3
- CometBFT v0.38.20
- Protobuf + gRPC

## Always Check Context7 Before Writing Cosmos SDK Code
```bash
mcporter call context7.resolve-library-id --args '{"libraryName": "cosmos-sdk"}'
mcporter call context7.get-library-docs --args '{"context7CompatibleLibraryID": "<id>", "topic": "<specific topic>"}'
```

---

## Module Structure (`x/vitacoin/`)

```
x/vitacoin/
├── keeper/          ← All state logic — the core
│   ├── keeper.go    ← Keeper struct + constructor
│   ├── msg_server.go← Message handler implementations
│   ├── grpc_query.go← gRPC query implementations
│   └── *_test.go   ← Tests alongside each file
├── types/           ← Types, messages, events, errors, params
│   ├── msgs.go      ← MsgX types + ValidateBasic
│   ├── params.go    ← Module params (governance-adjustable)
│   ├── errors.go    ← Sentinel errors
│   └── events.go    ← Event type constants
├── module/          ← Module registration + lifecycle
└── genesis.go       ← InitGenesis + ExportGenesis
```

---

## Code Standards

### Keeper Pattern
```go
// ✅ Keeper always uses store key, not global state
type Keeper struct {
    cdc          codec.BinaryCodec
    storeService store.KVStoreService
    bankKeeper   types.BankKeeper
    // ... other module keepers
}

// ✅ Always use ctx.KVStore(k.storeKey) not global vars
func (k Keeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
    store := k.storeService.OpenKVStore(ctx)
    // ...
}
```

### Message Handlers
```go
// ✅ Always: validate → check permissions → execute → emit event
func (k msgServer) Send(goCtx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // 1. Validate (should already be done in ValidateBasic, but double-check critical fields)
    fromAddr, err := sdk.AccAddressFromBech32(msg.FromAddress)
    if err != nil {
        return nil, sdkerrors.ErrInvalidAddress
    }

    // 2. Execute
    if err := k.bankKeeper.SendCoins(ctx, fromAddr, toAddr, msg.Amount); err != nil {
        return nil, err
    }

    // 3. Apply fee
    fee, err := k.ApplyFee(ctx, msg.Amount)
    if err != nil {
        return nil, err
    }

    // 4. Emit event (MANDATORY — blockchain is auditable)
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeSend,
            sdk.NewAttribute(types.AttributeKeyFrom, msg.FromAddress),
            sdk.NewAttribute(types.AttributeKeyTo, msg.ToAddress),
            sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
            sdk.NewAttribute(types.AttributeKeyFee, fee.String()),
        ),
    )

    return &types.MsgSendResponse{}, nil
}
```

### Error Handling
```go
// ✅ Use sentinel errors — never raw errors.New()
var (
    ErrInvalidAmount    = sdkerrors.Register(ModuleName, 1, "invalid amount")
    ErrInsufficientFund = sdkerrors.Register(ModuleName, 2, "insufficient funds")
    ErrUnauthorized     = sdkerrors.Register(ModuleName, 3, "unauthorized")
)

// ✅ Wrap with context
return nil, ErrInvalidAmount.Wrapf("got %s, minimum is %s", amount, minAmount)
```

---

## Testing Standards

### Every keeper function needs a test
```go
func TestKeeperSend(t *testing.T) {
    ctx, k := setupKeeper(t)

    // Setup
    addr1 := sdk.AccAddress([]byte("addr1"))
    addr2 := sdk.AccAddress([]byte("addr2"))

    // Test happy path
    err := k.Send(ctx, addr1, addr2, sdk.NewInt64Coin("uvita", 1000000))
    require.NoError(t, err)

    // Test edge cases
    err = k.Send(ctx, addr1, addr2, sdk.NewInt64Coin("uvita", 0))
    require.ErrorIs(t, err, types.ErrInvalidAmount)

    // Test insufficient funds
    err = k.Send(ctx, addr1, addr2, sdk.NewInt64Coin("uvita", 999999999999))
    require.ErrorIs(t, err, types.ErrInsufficientFund)
}
```

### Run tests
```bash
cd vitacoin
# All tests
make test

# Specific package with verbose output
go test -v ./x/vitacoin/keeper/ -timeout 120s

# With race detection (run before every PR)
make test-race

# Coverage
make test-cover
# Target: maintain 52%+ coverage
```

---

## Protobuf Standards

```protobuf
// ✅ Every message needs option fields documented
message MsgSend {
  string from_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string to_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  repeated cosmos.base.v1beta1.Coin amount = 3 [
    (gogoproto.nullable) = false,
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
    (amino.dont_omitempty) = true
  ];
}

// ✅ Always regenerate after proto changes
make proto-gen
```

---

## Chain Upgrade Protocol

**NEVER make breaking state changes without a migration.**

```go
// Every module version bump needs a migration handler
func RegisterUpgradeHandlers(mm *module.Manager, configurator module.Configurator) {
    app.UpgradeKeeper.SetUpgradeHandler("v1.1.0",
        func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
            // Run your migration here
            return mm.RunMigrations(ctx, configurator, fromVM)
        },
    )
}
```

Protocol:
1. Write migration in `app/upgrades/v1_1_0.go`
2. Test migration on local testnet with real state
3. Announce to validators 48h before mainnet upgrade height
4. Monitor upgrade height — be online

---

## Genesis & Params

```go
// ✅ All chain params must be governance-adjustable
// ✅ Default params defined in types/params.go
// ✅ Genesis export/import must be lossless — test it
func TestGenesisRoundTrip(t *testing.T) {
    ctx, k := setupKeeper(t)
    // Setup state
    genesis := k.ExportGenesis(ctx)
    // Reset
    k.InitGenesis(ctx, *genesis)
    // Verify state matches
}
```
