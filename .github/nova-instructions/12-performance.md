# Performance Targets — TPS, Finality, Mobile, Gateway

## Blockchain Performance Targets

| Metric | Target | Minimum Acceptable |
|---|---|---|
| Block time | 3–5 seconds | <8 seconds |
| Transaction finality | <6 seconds (2 blocks) | <15 seconds |
| TPS (sustained) | 100+ TPS | >50 TPS |
| TPS (burst) | 500+ TPS | >200 TPS |
| Node sync time (testnet) | <30 min from genesis | <2 hours |
| Binary size | <50 MB | <100 MB |
| Memory per validator node | <4 GB | <8 GB |

**How to benchmark:**
```bash
cd vitacoin
# Basic tx throughput test
make test-benchmark

# Load test with multiple tx senders
go test -bench=BenchmarkSendTx ./x/vitacoin/keeper/ -benchtime=30s
```

---

## Mobile Wallet Performance

| Metric | Target |
|---|---|
| App cold start | <2 seconds |
| Balance refresh | <1 second (cached), <3 seconds (chain) |
| QR scan to confirm screen | <0.5 seconds |
| TX broadcast to confirmation | <8 seconds (2 blocks) |
| Bundle size (JS) | <5 MB |
| Memory usage | <200 MB active |

### Rules
- Cache balance in state — refresh every 30s, not on every render
- Show cached balance immediately, update in background
- QR scanner must be instantaneous — no delay between scan and parse
- Never block the UI thread for chain queries — all async
- Use `FlatList` not `ScrollView` for transaction lists
- Paginate transactions: 20 per page

---

## Gateway Performance Targets

| Metric | Target |
|---|---|
| Payment relay p50 | <200ms |
| Payment relay p99 | <1 second |
| Auth endpoint | <100ms |
| Webhook delivery | <500ms after tx confirm |
| Uptime | 99.9% |

### Rules
- Use connection pooling for blockchain gRPC client
- Cache merchant config in memory (TTL: 5 min)
- Async webhook delivery — don't block payment response
- Rate limit: 100 req/min per merchant (hard), 1000 req/min total

---

## Website Performance (vitacoin.network)

| Metric | Target |
|---|---|
| Lighthouse Performance | >90 |
| LCP | <2 seconds |
| CLS | <0.1 |
| Initial page load | <1.5s on 4G |

```bash
cd workspace-vitacoin/vitacoin-web
npm run build
# Must show 0 errors, 0 warnings
npx lighthouse https://vitacoin.network --view
```

---

## Monitoring Checklist (Post-Mainnet)

- [ ] Block time alert: if >10s for 3 consecutive blocks → alert
- [ ] Validator uptime: if any validator misses >5% blocks → alert  
- [ ] Gateway error rate: if >1% 5xx → alert
- [ ] Mobile crash rate: if >0.5% sessions → alert
- [ ] Treasury balance: weekly report
- [ ] Burn rate: weekly report vs burn floor

---

## Developer Onboarding — Local Setup (60 min)

### Prerequisites
```bash
go version    # >= 1.24
node --version # >= 20
npm --version  # >= 10
make --version
```

### Step 1: Clone
```bash
git clone https://github.com/esspron/VITACOIN.git
cd VITACOIN
```

### Step 2: Build Blockchain
```bash
cd vitacoin
export PATH=$PATH:/usr/local/go/bin
make build
# Binary: build/vitacoind
```

### Step 3: Run Tests
```bash
make test
# All tests must pass before writing any code
```

### Step 4: Setup Mobile
```bash
cd vitapay-mobile
npm install
npx expo start
# Scan QR with Expo Go app
```

### Step 5: Setup Gateway
```bash
cd vitapay-gateway
cp .env.example .env  # fill in dev values
go run main.go
# Runs on :8080
```

### Step 6: Local Testnet (optional but recommended)
```bash
cd vitacoin
./scripts/local-testnet.sh  # if exists, else manual init
vitacoind start
# RPC: http://localhost:26657
# REST: http://localhost:1317
```

### Red Lines for New Devs
1. Never commit secrets, keys, or `.env` files
2. Always run `make test` before committing blockchain code
3. Never modify `go.sum` manually
4. Never push directly to `main`
5. Always update instruction files when code changes
