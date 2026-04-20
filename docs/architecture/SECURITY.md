# VITACOIN Security Guidelines

> **Version**: 1.0.0  
> **Last Updated**: October 16, 2025  
> **Security Level**: Production-Ready Framework

---

## 📋 Table of Contents

1. [Security Overview](#security-overview)
2. [Input Validation](#input-validation)
3. [Access Control](#access-control)
4. [Store Security](#store-security)
5. [Transaction Security](#transaction-security)
6. [Keeper Security Patterns](#keeper-security-patterns)
7. [Cryptographic Security](#cryptographic-security)
8. [Common Vulnerabilities](#common-vulnerabilities)
9. [Security Checklist](#security-checklist)
10. [Audit Procedures](#audit-procedures)
11. [Incident Response](#incident-response)

---

## Security Overview

VITACOIN implements **defense-in-depth security** with multiple layers of protection:

```
┌────────────────────────────────────────────────────────┐
│ Layer 1: Network Security (CometBFT Consensus)         │
│ - Byzantine Fault Tolerance (BFT)                      │
│ - 2/3+ validator signatures required                   │
└────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────┐
│ Layer 2: Transaction Validation (AnteHandler)          │
│ - Signature verification                               │
│ - Fee deduction & gas metering                         │
│ - Nonce/sequence checking                              │
└────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────┐
│ Layer 3: Message Validation (ValidateBasic)            │
│ - Stateless checks on message structure                │
│ - Input sanitization & bounds checking                 │
└────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────┐
│ Layer 4: Business Logic (Keeper)                       │
│ - Access control (authority checks)                    │
│ - State consistency validation                         │
│ - Inter-module permission checks                       │
└────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────┐
│ Layer 5: Store Security (KVStore)                      │
│ - Key prefix isolation                                 │
│ - Merkle-proof integrity                               │
│ - IAVL+ tree cryptographic hashing                     │
└────────────────────────────────────────────────────────┘
```

---

## Input Validation

### 1. ValidateBasic() - Stateless Validation

**CRITICAL**: Every message MUST implement `ValidateBasic()` for stateless checks before hitting the keeper.

#### Example: Proper Validation

```go
// MsgRegisterMerchant validation
func (msg *MsgRegisterMerchant) ValidateBasic() error {
    // 1. Validate creator address
    if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
        return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address: %s", err)
    }
    
    // 2. Validate business name
    if len(msg.BusinessName) == 0 {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "business name cannot be empty")
    }
    if len(msg.BusinessName) > 100 {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "business name too long (max 100 characters)")
    }
    
    // 3. Validate tier enum
    if msg.Tier < MerchantTier_BRONZE || msg.Tier > MerchantTier_PLATINUM {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid merchant tier")
    }
    
    // 4. Validate contact info format
    if !isValidEmail(msg.Email) {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid email format")
    }
    
    return nil
}
```

#### Validation Best Practices

✅ **DO**:
- Validate address formats (Bech32)
- Check string lengths (prevent DoS)
- Validate enum values
- Sanitize user inputs (remove control characters)
- Check numeric bounds (prevent overflow)
- Validate amount positivity

❌ **DON'T**:
- Access the store in `ValidateBasic()` (stateless only!)
- Trust user input without validation
- Ignore edge cases (empty strings, zero values, negative numbers)

### 2. Keeper-Level Validation

**Stateful validation** happens in the keeper after `ValidateBasic()`:

```go
func (ms msgServer) RegisterMerchant(ctx context.Context, msg *types.MsgRegisterMerchant) (*types.MsgRegisterMerchantResponse, error) {
    // 1. ValidateBasic already passed (done by SDK)
    
    // 2. Check if merchant already exists (stateful check)
    if ms.k.MerchantExists(ctx, msg.Creator) {
        return nil, sdkerrors.Wrap(types.ErrMerchantExists, msg.Creator)
    }
    
    // 3. Verify registration fee payment
    registrationFee := ms.k.GetParams(ctx).MerchantRegistrationFee
    if err := ms.k.bankKeeper.SendCoinsFromAccountToModule(
        ctx,
        sdk.MustAccAddressFromBech32(msg.Creator),
        types.ModuleName,
        sdk.NewCoins(sdk.NewCoin("avita", registrationFee)),
    ); err != nil {
        return nil, sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "registration fee payment failed")
    }
    
    // 4. Validate business logic constraints
    if msg.Tier == types.MerchantTier_PLATINUM {
        // Platinum merchants require KYC verification
        if !ms.k.IsKYCVerified(ctx, msg.Creator) {
            return nil, sdkerrors.Wrap(types.ErrUnauthorized, "Platinum tier requires KYC verification")
        }
    }
    
    // ... proceed with registration
}
```

---

## Access Control

### 1. Module Authority (Governance)

**CRITICAL**: Only x/gov should modify module parameters.

```go
type Keeper struct {
    authority string  // x/gov module address
}

func NewKeeper(..., authority string) Keeper {
    return Keeper{
        authority: authority,  // authtypes.NewModuleAddress(govtypes.ModuleName).String()
    }
}

// Enforce authority check
func (k Keeper) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) error {
    if msg.Authority != k.authority {
        return sdkerrors.Wrapf(types.ErrUnauthorized, 
            "expected %s, got %s", k.authority, msg.Authority)
    }
    
    // Validate new params
    if err := msg.Params.Validate(); err != nil {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
    }
    
    return k.SetParams(ctx, msg.Params)
}
```

### 2. Ownership Verification

Always verify that the signer owns the resource they're modifying:

```go
func (ms msgServer) CompletePayment(ctx context.Context, msg *types.MsgCompletePayment) error {
    // 1. Get payment from store
    payment, found := ms.k.GetPayment(ctx, msg.PaymentId)
    if !found {
        return sdkerrors.Wrap(sdkerrors.ErrNotFound, "payment not found")
    }
    
    // 2. Verify signer is the merchant (owner)
    if payment.Merchant != msg.Signer {
        return sdkerrors.Wrap(types.ErrUnauthorized, "only merchant can complete payment")
    }
    
    // 3. Verify payment state (prevent double completion)
    if payment.Status != types.PaymentStatus_PENDING {
        return sdkerrors.Wrap(types.ErrInvalidState, "payment not in pending state")
    }
    
    // ... proceed with completion
}
```

### 3. Role-Based Access Control (RBAC)

```go
// Define roles
type MerchantTier int32

const (
    MerchantTier_BRONZE   MerchantTier = 0
    MerchantTier_SILVER   MerchantTier = 1
    MerchantTier_GOLD     MerchantTier = 2
    MerchantTier_PLATINUM MerchantTier = 3
)

// Check permissions based on role
func (k Keeper) CanEnableInstantSettlement(ctx context.Context, merchantAddr string) bool {
    merchant, found := k.GetMerchant(ctx, merchantAddr)
    if !found {
        return false
    }
    
    // Only Gold and Platinum merchants can use instant settlement
    return merchant.Tier >= MerchantTier_GOLD
}
```

---

## Store Security

### 1. Key Prefix Isolation

**CRITICAL**: Use unique prefixes to prevent key collisions.

```go
// types/keys.go
var (
    ParamsKey          = []byte{0x01}  // Singleton key for params
    MerchantPrefix     = []byte{0x10}  // Prefix for merchant storage
    PaymentPrefix      = []byte{0x20}  // Prefix for payment storage
    VaultPrefix        = []byte{0x30}  // Prefix for loyalty vault storage
    RewardPoolPrefix   = []byte{0x40}  // Prefix for reward pool storage
    
    // Index keys (for lookups)
    MerchantByOwnerPrefix = []byte{0x11}
    PaymentByMerchantPrefix = []byte{0x21}
)

// Key construction with collision prevention
func MerchantKey(address string) []byte {
    return append(MerchantPrefix, []byte(address)...)
}

func PaymentKey(id string) []byte {
    return append(PaymentPrefix, []byte(id)...)
}

// Secondary index key
func MerchantByOwnerKey(ownerAddr, merchantAddr string) []byte {
    key := append(MerchantByOwnerPrefix, []byte(ownerAddr)...)
    return append(key, []byte(merchantAddr)...)
}
```

### 2. Store Access Patterns

```go
// ✅ CORRECT: Use store service pattern (v0.50.x)
func (k Keeper) SetMerchant(ctx context.Context, merchant types.Merchant) error {
    storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
    
    // Validate before storing
    if err := merchant.Validate(); err != nil {
        return err
    }
    
    key := types.MerchantKey(merchant.Address)
    value := k.cdc.MustMarshal(&merchant)
    storeAdapter.Set(key, value)
    
    return nil
}

// ❌ WRONG: Direct store access without validation
func (k Keeper) SetMerchantUnsafe(ctx context.Context, merchant types.Merchant) {
    store := k.storeService.OpenKVStore(ctx)
    // NO VALIDATION!
    store.Set([]byte(merchant.Address), k.cdc.MustMarshal(&merchant))
}
```

### 3. Atomic State Changes

Use transactional context for multi-step operations:

```go
func (k Keeper) ProcessPayment(ctx context.Context, paymentID string) error {
    // Get payment
    payment, found := k.GetPayment(ctx, paymentID)
    if !found {
        return sdkerrors.ErrNotFound
    }
    
    // 1. Transfer funds (can fail)
    if err := k.bankKeeper.SendCoins(ctx, payment.Customer, payment.Merchant, payment.Amount); err != nil {
        // State not changed if this fails
        return err
    }
    
    // 2. Update payment status (only if transfer succeeds)
    payment.Status = types.PaymentStatus_COMPLETED
    payment.CompletedAt = ctx.BlockTime().Unix()
    
    // 3. Store updated payment
    k.SetPayment(ctx, payment)
    
    // 4. Emit event
    ctx.EventManager().EmitEvent(...)
    
    // All steps succeed or all fail (atomic)
    return nil
}
```

---

## Transaction Security

### 1. Gas Metering

Prevent DoS attacks with proper gas consumption:

```go
// Charge gas for expensive operations
func (k Keeper) ListAllPayments(ctx context.Context) []types.Payment {
    storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
    
    iterator := sdk.KVStorePrefixIterator(storeAdapter, types.PaymentPrefix)
    defer iterator.Close()
    
    var payments []types.Payment
    for ; iterator.Valid(); iterator.Next() {
        // Charge gas for each iteration (prevent unbounded loops)
        ctx.GasMeter().ConsumeGas(sdk.Gas(10), "payment iteration")
        
        var payment types.Payment
        k.cdc.MustUnmarshal(iterator.Value(), &payment)
        payments = append(payments, payment)
    }
    
    return payments
}
```

### 2. Reentrancy Protection

Cosmos SDK is inherently protected from reentrancy, but be cautious with inter-module calls:

```go
// ✅ SAFE: State updated BEFORE external call
func (k Keeper) CompletePaymentSafe(ctx context.Context, payment types.Payment) error {
    // 1. Update state first
    payment.Status = types.PaymentStatus_COMPLETED
    k.SetPayment(ctx, payment)
    
    // 2. External call (to BankKeeper)
    return k.bankKeeper.SendCoins(ctx, payment.Customer, payment.Merchant, payment.Amount)
    
    // Even if SendCoins fails, state change is reverted atomically by SDK
}

// ❌ UNSAFE: External call BEFORE state update (theoretical vulnerability)
func (k Keeper) CompletePaymentUnsafe(ctx context.Context, payment types.Payment) error {
    // 1. External call first
    if err := k.bankKeeper.SendCoins(ctx, payment.Customer, payment.Merchant, payment.Amount); err != nil {
        return err
    }
    
    // 2. State update after
    payment.Status = types.PaymentStatus_COMPLETED
    k.SetPayment(ctx, payment)
    
    return nil
}
```

### 3. Integer Overflow Protection

Use `cosmossdk.io/math` for safe arithmetic:

```go
import "cosmossdk.io/math"

// ✅ SAFE: Using SDK math types
func (k Keeper) CalculateFee(amount math.Int, feePercent math.LegacyDec) math.Int {
    // Decimal multiplication with precision
    feeAmount := math.LegacyNewDecFromInt(amount).Mul(feePercent).TruncateInt()
    return feeAmount
}

// ❌ UNSAFE: Using native Go int64 (can overflow)
func (k Keeper) CalculateFeeUnsafe(amount int64, feePercent float64) int64 {
    return int64(float64(amount) * feePercent)  // OVERFLOW RISK!
}
```

---

## Keeper Security Patterns

### 1. Dependency Injection with Interfaces

Use interfaces to prevent circular dependencies and enable testing:

```go
// expected_keepers.go
type BankKeeper interface {
    SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error
    GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// keeper.go
type Keeper struct {
    bankKeeper BankKeeper  // Interface, not concrete type
}

// ✅ This allows mocking in tests
type MockBankKeeper struct{}
func (m MockBankKeeper) SendCoins(...) error { return nil }
```

### 2. Immutable Parameters

Once deployed, critical params should only change via governance:

```go
type Params struct {
    // Can be updated via governance
    TransactionFeePercent math.LegacyDec
    
    // Should never change (or require upgrade)
    PaymentTimeoutBlocks uint64
}

// Validate ensures params are within safe bounds
func (p Params) Validate() error {
    if p.TransactionFeePercent.GT(math.LegacyNewDec(10)) {
        return fmt.Errorf("fee percent too high (max 10%)")
    }
    return nil
}
```

### 3. Event Emission for Auditability

Emit events for all state changes:

```go
func (k Keeper) RegisterMerchant(ctx context.Context, merchant types.Merchant) error {
    k.SetMerchant(ctx, merchant)
    
    // Emit event for indexing and monitoring
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeMerchantRegistered,
            sdk.NewAttribute(types.AttributeKeyMerchant, merchant.Address),
            sdk.NewAttribute(types.AttributeKeyTier, merchant.Tier.String()),
            sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
        ),
    )
    
    return nil
}
```

---

## Cryptographic Security

### 1. Address Derivation

Use Cosmos SDK address utilities:

```go
// ✅ CORRECT: SDK address derivation
merchantAddr, err := sdk.AccAddressFromBech32("vitacoin1abc...")
if err != nil {
    return err
}

// Module accounts (deterministic)
moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
```

### 2. Signature Verification

The SDK's AnteHandler verifies signatures, but if you need manual verification:

```go
// Verify signature on message
pubKey, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, msg.PubKey)
if err != nil {
    return err
}

if !pubKey.VerifySignature(msg.GetSignBytes(), msg.Signature) {
    return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "signature verification failed")
}
```

### 3. Hashing & Merkle Proofs

IAVL+ tree provides Merkle proofs automatically:

```go
// Query with proof
proof, err := k.storeService.OpenKVStore(ctx).GetProof(key)
// Proof can be verified by light clients
```

---

## Common Vulnerabilities

### ❌ Vulnerability 1: Unbounded Iterations

**Problem**: Iterating entire store can cause OOM or block gas limit.

```go
// ❌ BAD: No pagination, unbounded loop
func (k Keeper) GetAllMerchants(ctx context.Context) []types.Merchant {
    var merchants []types.Merchant
    k.IterateMerchants(ctx, func(merchant types.Merchant) bool {
        merchants = append(merchants, merchant)
        return false  // Continue iteration
    })
    return merchants  // Could be millions of merchants!
}
```

**Solution**: Use pagination:

```go
// ✅ GOOD: Paginated query
func (k Keeper) GetMerchantsPaginated(ctx context.Context, req *types.QueryMerchantsRequest) (*types.QueryMerchantsResponse, error) {
    store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
    merchantStore := prefix.NewStore(store, types.MerchantPrefix)
    
    var merchants []types.Merchant
    pageRes, err := query.Paginate(merchantStore, req.Pagination, func(key, value []byte) error {
        var merchant types.Merchant
        k.cdc.MustUnmarshal(value, &merchant)
        merchants = append(merchants, merchant)
        return nil
    })
    
    return &types.QueryMerchantsResponse{
        Merchants: merchants,
        Pagination: pageRes,
    }, err
}
```

### ❌ Vulnerability 2: Missing Null Checks

```go
// ❌ BAD: Assumes payment exists
func (k Keeper) CompletePayment(ctx context.Context, id string) error {
    payment := k.GetPayment(ctx, id)  // Could return nil!
    payment.Status = types.PaymentStatus_COMPLETED  // PANIC if payment is nil
    return k.SetPayment(ctx, payment)
}

// ✅ GOOD: Check existence
func (k Keeper) CompletePayment(ctx context.Context, id string) error {
    payment, found := k.GetPayment(ctx, id)
    if !found {
        return sdkerrors.Wrap(sdkerrors.ErrNotFound, id)
    }
    payment.Status = types.PaymentStatus_COMPLETED
    return k.SetPayment(ctx, payment)
}
```

### ❌ Vulnerability 3: Time-Based Attacks

```go
// ❌ BAD: Using wall-clock time (manipulable by validators)
func (k Keeper) IsPaymentExpired(ctx context.Context, payment types.Payment) bool {
    return time.Now().Unix() > payment.ExpiresAt  // DON'T DO THIS!
}

// ✅ GOOD: Use block time (consensus-based)
func (k Keeper) IsPaymentExpired(ctx context.Context, payment types.Payment) bool {
    return ctx.BlockTime().Unix() > payment.ExpiresAt
}

// ✅ BETTER: Use block height (more reliable)
func (k Keeper) IsPaymentExpired(ctx context.Context, payment types.Payment) bool {
    return ctx.BlockHeight() > payment.ExpiresAtHeight
}
```

---

## Security Checklist

### Pre-Deployment Checklist

- [ ] All messages implement `ValidateBasic()`
- [ ] All keeper methods validate inputs
- [ ] Access control enforced (authority checks)
- [ ] Ownership verified before state changes
- [ ] Store keys use unique prefixes
- [ ] No unbounded iterations (pagination implemented)
- [ ] Gas metering for expensive operations
- [ ] Integer overflow prevention (`cosmossdk.io/math`)
- [ ] Atomic state updates (all-or-nothing)
- [ ] Events emitted for all state changes
- [ ] Parameter validation in `Params.Validate()`
- [ ] Genesis state validation in `GenesisState.Validate()`
- [ ] Module accounts have correct permissions
- [ ] No use of `time.Now()` (use `ctx.BlockTime()`)
- [ ] Error handling (no panics in production code)
- [ ] Test coverage >80% for critical paths

### Code Review Checklist

- [ ] Are all user inputs validated?
- [ ] Are access controls enforced?
- [ ] Are there any potential integer overflows?
- [ ] Is gas properly metered?
- [ ] Are store keys collision-proof?
- [ ] Are state changes atomic?
- [ ] Are errors handled gracefully?
- [ ] Are events logged for auditing?
- [ ] Is there test coverage?

---

## Audit Procedures

### Internal Audit Steps

1. **Static Analysis**: Run `golangci-lint` with security linters
2. **Unit Tests**: Achieve >80% coverage
3. **Integration Tests**: Test inter-module interactions
4. **Fuzz Testing**: Test with random inputs
5. **Gas Profiling**: Identify expensive operations
6. **Simulation Testing**: Use SDK's simulation framework

### External Audit

**Recommended Auditors**:
- Informal Systems
- Certik
- Trail of Bits
- Halborn

**Audit Focus Areas**:
1. Keeper logic (state management)
2. AnteHandler (transaction validation)
3. Inter-module calls (BankKeeper, StakingKeeper)
4. Parameter constraints
5. Upgrade paths

---

## Incident Response

### Severity Levels

| Level | Description | Response Time |
|-------|-------------|---------------|
| **Critical** | Funds at risk, consensus broken | Immediate (< 1 hour) |
| **High** | State corruption, major bug | < 4 hours |
| **Medium** | Minor bug, no immediate risk | < 24 hours |
| **Low** | Documentation, minor issues | < 7 days |

### Emergency Procedures

1. **Halt Chain** (if critical):
   ```bash
   # Emergency halt via governance or validator coordination
   vitacoind tx gov submit-proposal software-upgrade emergency-halt \
     --title "Emergency Chain Halt" \
     --description "Critical vulnerability discovered" \
     --upgrade-height <height>
   ```

2. **Coordinate with Validators**: Use validator communication channels

3. **Deploy Patch**: Test on private network first

4. **Restart Chain**: Coordinate restart after patch

---

## References

- **Cosmos SDK Security**: https://docs.cosmos.network/main/building-modules/security
- **CWE (Common Weakness Enumeration)**: https://cwe.mitre.org/
- **OWASP Blockchain Security**: https://owasp.org/www-project-blockchain/

---

**Security Contact**: security@vitacoin.network  
**Bug Bounty**: https://github.com/vishwas-io/VITACOIN/security/policy  
**Last Security Audit**: Pending (Q1 2026)

