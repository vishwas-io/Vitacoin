# Security — Blockchain-Grade (Non-Negotiable)

## Why Security is Different for Blockchain

A web app security breach = user data leaked. Bad, fixable.
A blockchain security breach = user funds permanently stolen. Irreversible.

There is no "we'll patch it later" in Web3. One mistake = project destroyed.

---

## Absolute Red Lines

### NEVER commit or push:
- Private keys of any kind (`priv_validator_key.json`, `node_key.json`)
- Wallet mnemonics or seed phrases (24-word recovery phrases)
- JWT secrets
- API keys, tokens, secrets
- `.env` files with real values
- Genesis files with real validator addresses (before mainnet ceremony)
- Any file matching `*key*.json` or `*mnemonic*` or `*seed*`

### NEVER do in code:
- Store private keys in plaintext (use `expo-secure-store` on mobile, HSM/KMS on server)
- Log private keys or mnemonics anywhere
- Accept private keys over API (sign client-side always)
- Use hardcoded fallback keys (`|| 'devkey123'`)
- Trust user-provided `user_id` without verifying JWT
- Skip signature verification on any blockchain message

---

## Pre-Push Security Scan (Run Every Time)

```bash
# Blockchain: check for keys, mnemonics, secrets
git diff --cached -U0 | grep "^\+" | grep -E "(eyJ[A-Za-z0-9]{30,}|sk-[A-Za-z0-9]{20,}|[a-f0-9]{64}|mnemonic|seed_phrase|priv_validator|node_key)"

# Must return EMPTY. If not → ABORT. Fix leak. Tell Vishwas.
```

---

## Wallet Security Standards (Mobile)

```typescript
// ✅ ALWAYS use expo-secure-store for mnemonic storage
import * as SecureStore from 'expo-secure-store';
await SecureStore.setItemAsync('wallet_mnemonic', mnemonic, {
  keychainAccessibility: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY
});

// ✅ ALWAYS require biometric auth before showing mnemonic or signing tx
import * as LocalAuthentication from 'expo-local-authentication';
const result = await LocalAuthentication.authenticateAsync({
  promptMessage: 'Confirm transaction',
  fallbackLabel: 'Use passcode'
});
if (!result.success) throw new Error('Authentication failed');

// ✅ NEVER store mnemonic in AsyncStorage, Context, or state
// ✅ NEVER log wallet addresses with amounts in production
// ✅ ALWAYS validate vitapay:// URI before processing payment
```

---

## Gateway Security Standards

```go
// ✅ ALWAYS verify JWT on every request
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        claims, err := verifyJWT(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        c.Set("merchant_id", claims.MerchantID)
        c.Next()
    }
}

// ✅ ALWAYS validate merchant owns the payment address
// ✅ ALWAYS rate limit: 100 req/min per merchant
// ✅ NEVER relay a transaction without verifying the signature
// ✅ ALWAYS validate amount > 0 and <= reasonable max
```

---

## Blockchain Module Security

```go
// ✅ ALWAYS validate msg in ValidateBasic()
func (msg *MsgSend) ValidateBasic() error {
    if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
        return sdkerrors.ErrInvalidAddress.Wrap("invalid from address")
    }
    if !msg.Amount.IsValid() || msg.Amount.IsZero() {
        return sdkerrors.ErrInvalidCoins.Wrap("amount must be positive")
    }
    return nil
}

// ✅ ALWAYS check sender balance before executing transfer
// ✅ ALWAYS use sdk.Coins not raw integers for token amounts
// ✅ NEVER allow integer overflow — use sdkmath.Int
// ✅ ALWAYS emit events for every state change (auditable)
```

---

## Validator Key Management

- Validator keys (`priv_validator_key.json`) NEVER leave the validator machine
- Use remote signer (tmkms) for mainnet validators
- Keep a backup of validator keys in an air-gapped device (Vishwas owns this)
- Rotate keys if any machine is compromised

---

## Smart Contract / Module Audit Checklist

Before any module ships to mainnet:
- [ ] Integer overflow/underflow checked for all math
- [ ] All message types have `ValidateBasic()`
- [ ] All keeper functions checked for permission (who can call this?)
- [ ] Events emitted for every state change
- [ ] Genesis import/export tested (InitGenesis/ExportGenesis)
- [ ] Upgrade migration tested
- [ ] No hardcoded addresses in module code
- [ ] All params are governance-adjustable

---

## Incident Response — If Keys Are Leaked

1. **Immediately** notify Vishwas
2. If validator key leaked → rotate validator key immediately, contact all delegators
3. If treasury key leaked → emergency governance proposal to move funds
4. If gateway JWT secret leaked → rotate secret, invalidate all sessions
5. If mobile app signing key leaked → publish emergency update, notify users
6. Document everything in `memory/YYYY-MM-DD-security-incident.md`

---

## Security Incident History

- 2026-03-30: Supabase + Resend keys hardcoded in CLAUDE.md → rotated
- 2026-04-03: `.env.production` tracked by git → purged
- Never again.
