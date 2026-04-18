# Web3 Standards — Tokenomics, IBC, Governance, Exchange Listings

## VitaCoin as a Payment-First Chain

VitaCoin is not a general-purpose chain. It is optimized for one thing: **fast, cheap payments for Indian merchants and e-commerce**.

Every technical decision must be evaluated through this lens:
> "Does this make VitaCoin faster, cheaper, or more trustworthy for a merchant?"

---

## Tokenomics (Immutable Until Governance Vote)

| Parameter | Value | Location |
|---|---|---|
| Denom | `uvita` (1 VITA = 1M uvita) | `types/params.go` |
| Fee rate | 0.1% per transaction | `params.FeeRate` |
| Validator share | 50% of fees | `params.ValidatorShare` |
| Burn share | 25% of fees | `params.BurnShare` |
| Treasury share | 25% of fees | `params.TreasuryShare` |
| Burn floor | 500,000,000 VITA | `params.BurnFloor` |
| Max supply | Defined in genesis | `genesis.json` |

**Any change to tokenomics requires:**
1. Governance proposal
2. Community vote
3. On-chain execution
4. Update this file + `01-architecture.md`

---

## IBC Integration Standards

```go
// ✅ IBC transfers follow escrow/mint/burn pattern
// Source chain: escrow tokens in IBC module account
// Destination chain: mint voucher tokens

// ✅ Always handle packet timeouts gracefully
// ✅ Always verify channel + port before sending IBC packet
// ✅ Never assume IBC ack is success — check ack.Success

// Supported IBC paths (update as channels open):
// vitacoin-1 ↔ cosmos-hub-4 (planned)
// vitacoin-1 ↔ osmosis-1 (planned for DEX liquidity)
```

---

## Governance Standards

```go
// ✅ All module params must be governance-adjustable
// ✅ Proposals have: title, description, changes, deposit
// ✅ Voting period: 7 days (mainnet)
// ✅ Quorum: 33.4% of staked VITA
// ✅ Veto threshold: 33.4%

// Proposal types VitaCoin supports:
// - ParameterChangeProposal (fee rate, shares, burn floor)
// - TextProposal (upgrades, partnerships)
// - SoftwareUpgradeProposal (chain upgrades)
// - CommunityPoolSpendProposal (treasury spending)
```

---

## Validator Strategy (Mainnet Prep)

### Target Validator Set
- 21 validators at genesis
- Minimum self-delegation: 10,000 VITA
- Commission range: 1%–20%
- Geographic distribution: India-first, then Southeast Asia

### Validator Recruitment Checklist
- [ ] 10+ committed validators before mainnet
- [ ] Each validator: hardware specs documented, uptime history verified
- [ ] Gentx ceremony date set and communicated
- [ ] Explorer operational (Ping.pub or custom)

### What Validators Get
- 50% of all transaction fees
- Block rewards (if inflationary model, else fee-only)
- Early validator badge + marketing partnership

---

## Exchange Listing Strategy

### Target Exchanges (Priority Order)
1. **CoinDCX / WazirX** — India-first, fits merchant target market
2. **Gate.io** — accessible listing, high volume
3. **KuCoin** — global exposure
4. **Osmosis DEX** — IBC native, DeFi liquidity

### Listing Requirements (prepare these now)
- [ ] Audit report from recognized firm (CertiK, Halborn)
- [ ] Tokenomics doc (public, professional)
- [ ] Whitepaper (investor-grade)
- [ ] GitHub activity (consistent commits, public tests)
- [ ] Live testnet + explorer
- [ ] Team KYC (Vishwas + team)
- [ ] Market maker contact

### Listing Materials Location
```
workspace-vitacoin/docs/
├── whitepaper.md          ← Nova maintains this
├── tokenomics.md          ← Nova maintains this
├── one-pager.md           ← for exchange inquiries
└── audit-report/          ← when available
```

---

## Chain Quality Standards

### Before Mainnet (mandatory)
- [ ] External security audit (minimum 1 recognized firm)
- [ ] 90-day testnet running with external validators
- [ ] All governance proposals tested on testnet
- [ ] IBC transfers tested with at least 2 counterparty chains
- [ ] Explorer operational and accurate
- [ ] Block explorer showing: txs, validators, proposals, IBC channels
- [ ] Emergency upgrade procedure tested

### Ongoing (post-mainnet)
- Never push an upgrade without 48h validator notice
- Every upgrade has a rollback plan
- Monitor chain health: block time, validator uptime, mempool size
- Emergency contact list for top 5 validators

---

## Code Quality for Web3

```go
// ✅ Deterministic execution — same input MUST produce same output
// No random in keeper functions. No timestamps in state transitions.
// Use block height for time-based logic, not time.Now()

// ✅ Bounded computation — no unbounded loops
// For iterating over state, always use pagination
// Unbounded iteration = chain halt vector

// ✅ Safe math everywhere
import "cosmossdk.io/math"
result, err := math.Int.Add(a, b) // not a + b directly

// ✅ Events for every state change
// Indexers, explorers, and dApps depend on events
// Missing event = invisible transaction
```
