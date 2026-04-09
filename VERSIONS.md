# VERSIONS.md — VitaCoin Dependency Version Lock

## ⚠️ READ THIS BEFORE TOUCHING go.mod

All dependency versions MUST match this file exactly.
Do NOT upgrade any dependency without updating this file first.
Do NOT use beta, rc, or pre-release versions.

## Locked Versions (cosmos-sdk v0.50.15 ecosystem)

| Package | Version | Notes |
|---------|---------|-------|
| `github.com/cosmos/cosmos-sdk` | `v0.50.15` | Latest stable v0.50 |
| `cosmossdk.io/store` | `v1.1.1` | Stable store, matches sdk v0.50.15 |
| `cosmossdk.io/core` | `v0.11.0` | Must match sdk v0.50.15 |
| `cosmossdk.io/api` | `v0.7.6` | Must match sdk v0.50.15 |
| `cosmossdk.io/collections` | `v0.4.0` | Must match sdk v0.50.15 |
| `cosmossdk.io/x/tx` | `v0.13.7` | Must match sdk v0.50.15 |
| `cosmossdk.io/x/upgrade` | `v0.1.4` | Must match sdk v0.50.15 |
| `cosmossdk.io/errors` | `v1.0.1` | Stable |
| `cosmossdk.io/log` | `v1.4.1` | Stable |
| `cosmossdk.io/math` | `v1.4.0` | Stable |
| `cosmossdk.io/depinject` | `v1.1.0` | Stable |
| `github.com/cometbft/cometbft` | `v0.38.20` | CometBFT for sdk v0.50 |
| `github.com/cosmos/iavl` | `v1.2.2` | iavl for store v1.1.1 |
| `github.com/cosmos/cosmos-db` | `v1.1.1` | Stable |
| `github.com/cosmos/cosmos-proto` | `v1.0.0-beta.5` | Standard |
| `github.com/cosmos/gogoproto` | `v1.7.0` | Standard |

## Rules

1. **No beta/rc versions** — only stable releases
2. **No `replace` directives** — if something needs patching, fork properly
3. **All versions come from cosmos-sdk's own go.mod** — don't independently upgrade transitive deps
4. **Before any go.mod change:** check this file and cosmos-sdk's go.mod for the target version
5. **After any go.mod change:** run `go mod tidy && go build ./vitacoin/cmd/vitacoind && make test`

## Version Upgrade Process

1. Choose new cosmos-sdk version (e.g., v0.50.16)
2. Read its go.mod: `go list -m -json github.com/cosmos/cosmos-sdk@v0.50.16`
3. Update ALL versions in this file to match
4. Update go.mod to match
5. Fix any API changes in app.go
6. Build + test
7. Commit VERSIONS.md + go.mod + go.sum + app.go together

## Why v0.50.15?

- v0.50.x is the latest stable production line (used by Osmosis, Cosmos Hub, etc.)
- v0.53.x is unreleased/pre-release — caused the REST API bug
- store v1.3.0-beta.0 + iavl v1.3.0 had unresolved iavl `GetRoot` issues
- store v1.1.1 + iavl v1.2.2 is battle-tested in production

## What Was Broken (Apr 9, 2026)

The original go.mod used:
- cosmos-sdk v0.53.0 (unreleased)
- store v1.3.0-beta.0 (beta!)
- iavl v1.3.0 (pre-release)
- cometbft v0.39.0-beta.2 (beta!)
- A local `replace cosmossdk.io/store => ./patches/cosmossdk-store` hack

This caused REST API "version does not exist" errors because the iavl/store
versions were incompatible. Root keys existed in the DB but couldn't be read.
