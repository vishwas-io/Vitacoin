# Context7 & Dependency Standards

## MANDATORY: Always Use Context7 Before Writing Integration Code

```bash
# Step 1: Resolve library ID
mcporter call context7.resolve-library-id --args '{"libraryName": "<library>"}'

# Step 2: Fetch current docs
mcporter call context7.get-library-docs --args '{
  "context7CompatibleLibraryID": "<id>",
  "topic": "<specific topic>",
  "tokens": 8000
}'
```

If Context7 doesn't have it → read the source in `vendor/` or `node_modules/` directly.
**Never rely on training knowledge alone for third-party APIs.**

---

## Libraries That Always Require Context7

| Library | Why |
|---|---|
| `github.com/cosmos/cosmos-sdk` | Keeper/module APIs evolve rapidly across minor versions |
| `github.com/cometbft/cometbft` | ABCI interface changes |
| `@cosmjs/stargate` | TX building, signing, broadcasting patterns |
| `@cosmjs/proto-signing` | Wallet/signer API |
| `expo` | SDK API surface changes every SDK release |
| `expo-secure-store` | Storage API + accessibility options |
| `expo-local-authentication` | Biometric auth patterns |
| `react-native` | Platform-specific APIs change frequently |
| `github.com/gin-gonic/gin` | Middleware + handler patterns |
| `github.com/golang-jwt/jwt` | Token generation + validation |
| Any new package | Always |

---

## Version Registry — Single Source of Truth

**NEVER introduce a version without updating this table.**
**NEVER have version mismatches across modules.**

### Blockchain Core (`vitacoin/go.mod`)

| Package | Version | Last Verified |
|---|---|---|
| github.com/cosmos/cosmos-sdk | v0.50.15 | — |
| github.com/cometbft/cometbft | v0.38.20 | — |
| cosmossdk.io/core | v0.11.0 | — |
| cosmossdk.io/math | v1.4.0 | — |
| cosmossdk.io/store | v1.1.1 | — |
| cosmossdk.io/x/tx | v0.13.7 | — |
| github.com/cosmos/ibc-go | (check go.mod) | — |
| go | 1.24 | — |

### VITAPAY Gateway (`vitapay-gateway/go.mod`)

| Package | Version | Last Verified |
|---|---|---|
| github.com/gin-gonic/gin | v1.9.1 | — |
| github.com/golang-jwt/jwt/v5 | v5.2.0 | — |
| github.com/google/uuid | v1.4.0 | — |
| go | 1.21 | — |

### VITAPAY Mobile (`vitapay-mobile/package.json`)

| Package | Version | Last Verified |
|---|---|---|
| expo | ~54.0.33 | — |
| react | 19.1.0 | — |
| react-native | 0.81.5 | — |
| @cosmjs/stargate | ^0.32.4 | — |
| @cosmjs/proto-signing | ^0.32.4 | — |
| expo-secure-store | ~14.0.1 | — |
| expo-local-authentication | ^55.0.13 | — |
| typescript | ~5.9.2 | — |

---

## Rules for Adding New Dependencies

### Go
1. Check Context7 for current API docs
2. Verify compatibility with Cosmos SDK v0.50.x
3. Add to the relevant `go.mod` — run `go mod tidy`
4. Add to this version registry
5. Commit: `chore: add <package>@<version> for <reason>`

### React Native
1. Check Context7 + Expo compatibility matrix
2. Verify it works with Expo SDK 54
3. Check iOS + Android support (don't add iOS-only packages)
4. Run `npx expo install <package>` — NOT `npm install` (Expo manages versions)
5. Add to this version registry

### Never Add
- Packages not maintained in last 12 months
- Packages with known security vulnerabilities
- Packages that require native module linking without Expo prebuild
- Duplicate functionality (e.g., don't add `axios` if `fetch` works)

---

## Context7 API Key

`ctx7sk-d460bbf1-8354-4219-ace7-30c3b91c62f8`

Fallback to curl if MCP unavailable:
```bash
curl -s "https://context7.com/api/v1/search?query=cosmos-sdk" \
  -H "Authorization: Bearer ctx7sk-d460bbf1-8354-4219-ace7-30c3b91c62f8"
```
