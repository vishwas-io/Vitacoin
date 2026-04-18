# System Architecture — VitaCoin Full Stack

## Overview

```
┌─────────────────────────────────────────────────────────┐
│                    VitaCoin Ecosystem                    │
├─────────────────────────────────────────────────────────┤
│  vitacoin.network (Next.js → Vercel)                    │
│  Investor site, roadmap, testnet stats                   │
├─────────────────────────────────────────────────────────┤
│  VITAPAY Mobile Wallet (React Native + Expo)             │
│  HD wallet, QR payments, staking, CosmJS                 │
│              ↕ REST/gRPC                                 │
│  VITAPAY Gateway (Go/Gin)                                │
│  Merchant API, payment relay, JWT auth                   │
│              ↕ gRPC + Tendermint RPC                     │
│  VitaCoin Blockchain (Go + Cosmos SDK v0.50.3)           │
│  x/vitacoin module, IBC, staking, governance             │
└─────────────────────────────────────────────────────────┘
```

## Components

### 1. VitaCoin Blockchain
- **Language:** Go 1.24
- **Framework:** Cosmos SDK v0.50.3 + CometBFT v0.38.20
- **Module:** `x/vitacoin` — custom payment + fee + treasury module
- **Location:** `vitacoin/`
- **Build:** `cd vitacoin && make build` → binary `vitacoind`
- **Test:** `cd vitacoin && make test`
- **Key files:**
  - `x/vitacoin/keeper/` — all state logic
  - `x/vitacoin/types/` — message types, params
  - `x/vitacoin/module/` — module wiring
  - `proto/vitacoin/` — protobuf definitions

### 2. VITAPAY Payment Gateway
- **Language:** Go 1.21
- **Framework:** Gin v1.9.1
- **Location:** `vitapay-gateway/`
- **Responsibilities:** Merchant auth (JWT), payment relay to chain, webhook dispatch
- **Key files:**
  - `api/` — route handlers
  - `blockchain/` — chain client (gRPC)
  - `middleware/` — auth, logging, rate limiting

### 3. VITAPAY Mobile Wallet
- **Language:** TypeScript + React Native
- **Framework:** Expo SDK ~54, React 19
- **Location:** `vitapay-mobile/`
- **Responsibilities:** HD wallet, send/receive VITA, QR payments, staking UI
- **Key libs:** `@cosmjs/stargate`, `@cosmjs/proto-signing`, `expo-secure-store`
- **Key dirs:**
  - `src/screens/` — all screens
  - `src/services/` — chain client, wallet, storage
  - `src/components/` — reusable UI

### 4. vitacoin.network Website
- **Framework:** Next.js (App Router)
- **Location:** `workspace-vitacoin/vitacoin-web/`
- **Deploy:** Vercel → `vercel --prod --yes`
- **Key file:** `app/page.tsx` — all stats, roadmap, ticker

---

## Data Flow — Payment

```
User opens VITAPAY app
  → Scans merchant QR (vitapay://<address>?amount=<vita>)
  → App signs tx with user's private key (CosmJS)
  → Broadcasts to VitaCoin RPC endpoint
  → Chain validates + applies fee (0.1%)
    → 50% → validators
    → 25% → burn
    → 25% → treasury
  → Merchant's VITAPAY Gateway webhook fires
  → Merchant's POS confirms payment
```

## Data Flow — Staking

```
User delegates VITA → chain
  → Validator set updated
  → Block rewards distributed proportionally
  → User can claim rewards or restake
  → Liquid staking: user gets stVITA (tradeable)
```

## Deployment Targets

| Component | Platform | URL |
|---|---|---|
| Blockchain (testnet) | VPS / Cloud Run | RPC: configured in STATUS.md |
| Gateway | Google Cloud Run | `https://api.vitacoin.network` (target) |
| Mobile Wallet | Expo / App Store | testnet build via EAS |
| Website | Vercel | `https://vitacoin.network` |

## Key Addresses & Params (Testnet)

> Never hardcode mainnet params. All chain params in `vitacoin/x/vitacoin/types/params.go`

- **Denom:** `uvita` (micro-VITA, 1 VITA = 1,000,000 uvita)
- **Fee rate:** 0.1% of each transaction
- **Fee split:** 50% validators / 25% burn / 25% treasury
- **Burn floor:** 500,000,000 VITA
- **Max supply:** (defined in genesis)

---

## Module Dependencies

```
x/vitacoin depends on:
  - x/bank (token transfers)
  - x/staking (validator set)
  - x/gov (governance)
  - ibc-go (cross-chain)

Never import external modules without Vishwas approval.
Adding a new Cosmos SDK module = security review first.
```
