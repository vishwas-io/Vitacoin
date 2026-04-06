# VITAPAY Gateway

The VITAPAY Gateway is the Go/Gin HTTP service that bridges merchants and the VitaCoin blockchain. It handles payment creation, QR/deep-link generation, on-chain confirmation, webhook delivery, and merchant management.

---

## Architecture

```
Merchant App ──► VITAPAY Gateway (Go/Gin) ──► VitaCoin Chain (REST)
                        │
                        └──► Merchant Webhook
```

- **In-memory store** (default) — swap for Supabase by setting `DATABASE_URL`
- **JWT auth** on merchant/webhook management endpoints
- **Blockchain client** (`blockchain/client.go`) calls Cosmos REST API

---

## API Endpoints

### Public

#### `GET /api/v1/health`
```json
{ "status": "ok", "chain": "vitacoin-1", "version": "1.0.0" }
```

---

### Payment (public)

#### `POST /api/v1/payment/create`
**Request:**
```json
{
  "merchantAddress": "vita1qg5ega6dykkxc307y25pecuufrjkxkacv2p07",
  "amount": "1000000",
  "denom": "uvita",
  "memo": "order-123",
  "expiresIn": 900,
  "webhookUrl": "https://yoursite.com/vitapay/webhook"
}
```
**Response (201):**
```json
{
  "paymentId": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "qrData": "vitapay://pay?to=vita1...&amount=1000000&denom=uvita&memo=...&expires=1712345678",
  "deepLink": "vitapay://pay?to=vita1...&amount=1000000&denom=uvita&memo=...&expires=1712345678",
  "expiresAt": "2026-04-07T03:00:00Z",
  "status": "pending"
}
```

#### `GET /api/v1/payment/:id`
Returns full payment record. Auto-checks chain for confirmation if still pending.

#### `POST /api/v1/payment/:id/confirm`
**Request:**
```json
{ "txHash": "ABC123DEF..." }
```
Verifies tx on-chain, marks payment confirmed, fires webhook.

---

### Merchant (JWT required)

> Add `Authorization: Bearer <token>` header.

#### `POST /api/v1/merchant/register`
**Request:**
```json
{
  "address": "vita1qg5ega6dykkxc307y25pecuufrjkxkacv2p07",
  "businessName": "Cafe Vita",
  "webhookUrl": "https://cafevita.com/payments/webhook"
}
```
**Response (201):**
```json
{
  "merchantId": "uuid",
  "address": "vita1...",
  "businessName": "Cafe Vita",
  "registeredAt": "2026-04-07T00:00:00Z"
}
```

#### `GET /api/v1/merchant/:address/payments?page=1&limit=20`
Paginated payments sorted by `createdAt` desc.

#### `GET /api/v1/merchant/:address/stats`
```json
{
  "address": "vita1...",
  "totalPayments": 42,
  "confirmedPayments": 38,
  "pendingPayments": 2,
  "expiredPayments": 2,
  "totalVolumeUVITA": 38000000,
  "successRate": 90.47,
  "last30DaysBreakdown": { "2026-04-06": 5, "2026-04-05": 3 }
}
```

---

### Webhook

#### `POST /api/v1/webhook/register` (JWT required)
```json
{ "merchantAddress": "vita1...", "webhookUrl": "https://yoursite.com/hook" }
```

Webhook POSTs (on confirmation):
```json
{
  "paymentId": "...",
  "merchantAddress": "vita1...",
  "amount": "1000000",
  "denom": "uvita",
  "status": "confirmed",
  "txHash": "ABC...",
  "confirmedAt": "2026-04-07T00:01:00Z"
}
```

---

## Running Locally

```bash
export PORT=8080
export VITAPAY_JWT_SECRET=your-secret-here
export VITACOIN_REST=https://api.vitacoin.network
export CHAIN_ID=vitacoin-1

go run main.go
```

**Test:**
```bash
export PATH=$PATH:/usr/local/go/bin
go test ./... -v
```

---

## Deploying to Cloud Run

```bash
# Build & push image
docker build -t gcr.io/YOUR_PROJECT/vitapay-gateway:latest .
docker push gcr.io/YOUR_PROJECT/vitapay-gateway:latest

# Deploy
gcloud run deploy vitapay-gateway \
  --image gcr.io/YOUR_PROJECT/vitapay-gateway:latest \
  --platform managed \
  --region asia-south1 \
  --allow-unauthenticated \
  --set-env-vars VITAPAY_JWT_SECRET=xxx,VITACOIN_REST=https://api.vitacoin.network,CHAIN_ID=vitacoin-1 \
  --port 8080
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `VITAPAY_JWT_SECRET` | *(required)* | JWT signing secret |
| `VITACOIN_REST` | `https://api.vitacoin.network` | Chain REST endpoint |
| `VITACOIN_RPC` | `https://rpc.vitacoin.network` | Chain RPC endpoint |
| `CHAIN_ID` | `vitacoin-1` | Chain ID |
| `MODULE_ADDRESS` | *(optional)* | Gateway module account |
| `DATABASE_URL` | *(optional)* | Supabase postgres (enables persistence) |

---

## Security

- JWT tokens required for all merchant/webhook management
- Never commit `.env` or secrets to git (repo is public)
- Run behind HTTPS (Cloud Run provides TLS automatically)
