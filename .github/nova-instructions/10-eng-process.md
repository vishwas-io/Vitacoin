# Engineering Process — PRs, Chain Upgrades, Incidents

## Git Workflow

```
feature/your-feature
        ↓ PR + review
     develop
        ↓ Vishwas approves
       main → testnet deploy (auto)
        ↓ Vishwas mainnet approval
     mainnet (manual, ceremony required)
```

### Branch naming
```
feature/staking-liquid-unstake
fix/ibc-packet-timeout-handler
chore/update-cosmos-sdk-v0.50.16
test/governance-voting-e2e
docs/update-api-contracts
upgrade/v1.1.0-fee-param-change
```

### Commit format
```
type(scope): description

Types: feat | fix | chore | docs | test | refactor | upgrade | security
Scopes: blockchain | gateway | mobile | website | infra
```

Examples:
```
feat(blockchain): add liquid staking unstake queue
fix(mobile): handle CosmJS broadcast timeout with retry
security(gateway): rotate JWT secret + invalidate sessions
upgrade(blockchain): v1.1.0 fee parameter governance migration
```

---

## PR Rules

### Every PR must have:
- [ ] Clear title + description
- [ ] What changed, why, how to test
- [ ] All tests passing
- [ ] Security scan clean
- [ ] Context sync: relevant instruction file updated
- [ ] For blockchain changes: `make test` + `make test-race` passing
- [ ] For mobile changes: tested on iOS simulator + Android emulator
- [ ] Screenshots for any UI change (before + after)

### PR Size
- Under 400 lines preferred
- Chain module changes: one logical change per PR
- Never bundle: tokenomics change + UI change in same PR

---

## Chain Upgrade Protocol

This is the most dangerous operation in blockchain. Follow exactly.

### Upgrade Checklist
1. [ ] Write upgrade handler in `app/upgrades/v<version>/`
2. [ ] Write migration for any state changes
3. [ ] Test upgrade on local single-node testnet
4. [ ] Test upgrade on public testnet (with real state if possible)
5. [ ] Submit `SoftwareUpgradeProposal` on testnet — verify it passes
6. [ ] Notify all validators **48 hours** before upgrade height
7. [ ] Document upgrade in `docs/upgrades/v<version>.md`
8. [ ] Be online at upgrade height
9. [ ] Monitor first 100 blocks post-upgrade for anomalies
10. [ ] Update `STATUS.md` + instruction files after success

### Rollback Plan (always have one)
```bash
# If upgrade fails at height H:
# 1. Validators coordinate to use previous binary
# 2. Set unsafe-skip-upgrades in node config
# 3. Emergency governance vote to cancel upgrade
# Never assume an upgrade will succeed — always have rollback ready
```

---

## Database / State Migration Rules

**On-chain state is permanent. Migrations are irreversible.**

1. Every state change needs a migration handler
2. Never rename a store key (add new key, migrate data, remove old key in next upgrade)
3. Test migration on testnet with real-world state size
4. Always write a genesis export/import test around migrations
5. Document every migration in `docs/migrations/`

---

## Incident Response

### Chain Halt
1. Notify Vishwas immediately
2. Check: is it a consensus failure or app-level panic?
3. Identify: is halt on all nodes or just some?
4. If app panic: identify the goroutine, fix, and prepare patch binary
5. Coordinate validators to restart with patch
6. Post-mortem in `memory/YYYY-MM-DD-incident.md`

### Mobile App Critical Bug
1. Pull the app update (if keys at risk: force update or disable)
2. Notify Vishwas
3. Fix → TestFlight/internal testing → EAS update
4. Post-mortem

### Gateway Outage
```bash
# Check Cloud Run service
gcloud run services describe vitapay-gateway --region asia-south1

# Rollback to previous revision
gcloud run services update-traffic vitapay-gateway \
  --to-revisions=<prev>=100 --region asia-south1
```

### Severity
- **P0:** Chain halt, funds at risk, auth bypass → fix in <1h, wake Vishwas
- **P1:** Gateway down, mobile crashes on launch → fix in <4h
- **P2:** Feature broken → fix in <24h
- **P3:** UI bug, non-critical → next session

---

## Testnet Validator Onboarding (for new validators)

```bash
# 1. Init node
vitacoind init <moniker> --chain-id vitacoin-testnet-1

# 2. Download genesis
curl -o ~/.vitacoind/config/genesis.json <genesis-url>

# 3. Add seed nodes to config.toml
seeds = "<seed-node-id>@<ip>:26656"

# 4. Start
vitacoind start

# 5. Create validator (after synced)
vitacoind tx staking create-validator \
  --amount 10000000000uvita \
  --pubkey $(vitacoind tendermint show-validator) \
  --moniker "<name>" \
  --commission-rate 0.05 \
  --commission-max-rate 0.2 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1 \
  --chain-id vitacoin-testnet-1 \
  --from <key>
```
