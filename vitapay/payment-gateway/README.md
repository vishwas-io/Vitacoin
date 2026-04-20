# 🌐 VITAPAY Payment Gateway

Merchant-facing REST API for accepting VITA payments. This is what e-commerce sites and apps integrate with.

## Status
🚧 **Planning Phase** - Not yet implemented

## Overview

The Payment Gateway provides APIs for merchants to:
- Create payment requests
- Generate QR codes
- Receive webhook notifications
- Verify payments on-chain
- Manage API keys
- Handle refunds (if applicable)

**Think of it as**: Stripe/Razorpay API, but for VITA cryptocurrency.

## Technology Stack

- **Language**: Go 1.21+
- **Framework**: Gin (REST API)
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Queue**: RabbitMQ
- **Auth**: JWT tokens
- **Blockchain**: CosmJS (via wrapper)

## Features

### Core Features
- [x] Payment request creation *(planned)*
- [x] QR code generation *(planned)*
- [x] Payment verification *(planned)*
- [x] Webhook notifications *(planned)*
- [x] API key management *(planned)*
- [x] Transaction history *(planned)*
- [x] Multi-merchant support *(planned)*

### Advanced Features
- [x] Fiat price conversion *(planned)*
- [x] Multiple currencies *(planned)*
- [x] Refund handling *(planned)*
- [x] Payment expiry *(planned)*
- [x] Rate limiting *(planned)*
- [x] Analytics & reporting *(planned)*

## Project Structure

```
payment-gateway/
├── README.md           # This file
├── go.mod              # Go dependencies
├── Makefile            # Build commands
│
├── cmd/
│   └── server/
│       └── main.go     # Entry point
│
├── internal/
│   ├── api/            # HTTP handlers
│   │   ├── handlers/
│   │   │   ├── payments.go
│   │   │   ├── webhooks.go
│   │   │   └── merchants.go
│   │   └── middleware/
│   │       ├── auth.go
│   │       └── ratelimit.go
│   │
│   ├── services/       # Business logic
│   │   ├── payment.go
│   │   ├── webhook.go
│   │   └── blockchain.go
│   │
│   ├── models/         # Data models
│   │   ├── payment.go
│   │   ├── merchant.go
│   │   └── webhook.go
│   │
│   ├── repository/     # Database access
│   │   ├── postgres/
│   │   └── redis/
│   │
│   └── blockchain/     # VITACOIN client
│       ├── client.go
│       └── monitor.go
│
├── migrations/         # Database migrations
│   ├── 001_initial.sql
│   └── 002_webhooks.sql
│
├── config/             # Configuration
│   └── config.go
│
└── scripts/            # Utility scripts
    ├── migrate.sh
    └── seed.sh
```

## Setup

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- RabbitMQ (optional, for queues)
- VITACOIN node access

### Installation

```bash
# Install dependencies
go mod download

# Setup database
make migrate-up

# Seed test data
make seed

# Run server
make run
```

### Configuration

Create `config.yaml`:
```yaml
server:
  port: 8080
  environment: development

database:
  host: localhost
  port: 5432
  user: vitapay
  password: password
  database: vitapay_gateway

redis:
  host: localhost
  port: 6379

blockchain:
  rpc_url: https://rpc.vitacoin.network
  chain_id: vitacoin-1
  gas_price: 0.025uvita

webhooks:
  timeout: 30s
  retry_attempts: 3
  retry_delay: 5s
```

## API Endpoints

### Authentication
All requests require API key in header:
```
Authorization: Bearer <api-key>
```

### Create Payment Request
```http
POST /api/v1/payments
Content-Type: application/json

{
  "amount": "100.00",
  "currency": "VITA",
  "order_id": "ORDER-123",
  "description": "2x T-Shirts",
  "return_url": "https://mystore.com/orders/123",
  "webhook_url": "https://mystore.com/webhooks/vitapay"
}

Response:
{
  "id": "pay_abc123",
  "status": "pending",
  "amount": "100.00",
  "currency": "VITA",
  "recipient_address": "vita1merchant123...",
  "qr_code": "data:image/png;base64,...",
  "payment_url": "vitapay://pay?id=pay_abc123",
  "expires_at": "2025-10-16T12:30:00Z",
  "created_at": "2025-10-16T12:00:00Z"
}
```

### Get Payment Status
```http
GET /api/v1/payments/:id

Response:
{
  "id": "pay_abc123",
  "status": "completed",
  "amount": "100.00",
  "currency": "VITA",
  "tx_hash": "0xABC123...",
  "paid_at": "2025-10-16T12:05:23Z",
  "order_id": "ORDER-123"
}
```

### List Payments
```http
GET /api/v1/payments?status=completed&limit=50&offset=0

Response:
{
  "payments": [...],
  "total": 150,
  "limit": 50,
  "offset": 0
}
```

### Register Webhook
```http
POST /api/v1/webhooks

{
  "url": "https://mystore.com/webhooks/vitapay",
  "events": ["payment.completed", "payment.failed"],
  "secret": "webhook_secret_123"
}

Response:
{
  "id": "wh_abc123",
  "url": "https://mystore.com/webhooks/vitapay",
  "events": ["payment.completed", "payment.failed"],
  "created_at": "2025-10-16T12:00:00Z"
}
```

## Webhook Events

### Payment Completed
```json
{
  "event": "payment.completed",
  "timestamp": "2025-10-16T12:05:23Z",
  "data": {
    "payment_id": "pay_abc123",
    "order_id": "ORDER-123",
    "amount": "100.00",
    "currency": "VITA",
    "tx_hash": "0xABC123...",
    "from_address": "vita1customer...",
    "to_address": "vita1merchant...",
    "block_height": 123456
  }
}
```

### Webhook Signature
Verify webhook authenticity:
```go
signature := r.Header.Get("X-VITAPAY-Signature")
expectedSig := hmac.SHA256(webhookSecret, requestBody)
if signature != expectedSig {
    // Invalid webhook
}
```

## Payment Flow

```
1. Merchant creates payment request via API
   POST /api/v1/payments
   
2. Gateway generates:
   - Unique payment ID
   - Merchant's VITA address
   - QR code
   - Deep link

3. Customer scans QR code with VITAPAY wallet

4. Customer confirms payment
   - VITA sent on blockchain
   - Transaction broadcast

5. Gateway monitors blockchain
   - Detects incoming transaction
   - Verifies amount and recipient
   - Updates payment status

6. Gateway sends webhook to merchant
   - Merchant receives notification
   - Merchant fulfills order

7. Payment marked as completed
```

## Database Schema

### Merchants Table
```sql
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    api_key_hash VARCHAR(255) NOT NULL,
    vita_address VARCHAR(255) NOT NULL,
    webhook_url VARCHAR(255),
    webhook_secret VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Payments Table
```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id),
    order_id VARCHAR(255),
    amount DECIMAL(20, 8) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    recipient_address VARCHAR(255) NOT NULL,
    sender_address VARCHAR(255),
    tx_hash VARCHAR(255),
    block_height BIGINT,
    paid_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_payments_merchant ON payments(merchant_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_tx_hash ON payments(tx_hash);
```

### Webhooks Table
```sql
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID REFERENCES payments(id),
    webhook_url VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    response_code INT,
    response_body TEXT,
    attempts INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending',
    next_retry_at TIMESTAMP,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Blockchain Monitor

### Transaction Monitoring
```go
// Monitor blockchain for incoming payments
func (s *BlockchainService) MonitorPayments(ctx context.Context) {
    // Subscribe to new blocks
    blockCh := s.client.SubscribeNewBlocks(ctx)
    
    for block := range blockCh {
        // Process each transaction
        for _, tx := range block.Transactions {
            // Check if it's a payment we're expecting
            payment := s.findPendingPayment(tx.To, tx.Amount)
            if payment != nil {
                // Verify and complete payment
                s.completePayment(payment, tx)
            }
        }
    }
}
```

## Security

### API Key Authentication
- API keys hashed with bcrypt
- Rate limiting per key
- IP whitelisting (optional)
- Key rotation support

### Webhook Security
- HMAC signature verification
- Replay attack prevention
- TLS required
- Secret per merchant

### Payment Security
- On-chain verification
- Amount validation
- Address validation
- Expiry enforcement

## Testing

```bash
# Unit tests
make test

# Integration tests
make test-integration

# E2E tests
make test-e2e

# Load tests
make test-load
```

### Example Test
```go
func TestCreatePayment(t *testing.T) {
    req := &CreatePaymentRequest{
        Amount:      "100.00",
        Currency:    "VITA",
        OrderID:     "TEST-123",
        Description: "Test payment",
    }
    
    payment, err := service.CreatePayment(ctx, req)
    assert.NoError(t, err)
    assert.Equal(t, "pending", payment.Status)
    assert.NotEmpty(t, payment.QRCode)
}
```

## Deployment

### Docker
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o gateway cmd/server/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/gateway /gateway
EXPOSE 8080
CMD ["/gateway"]
```

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vitapay-gateway
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: gateway
        image: vitapay/gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: url
```

## Performance

### Targets
- Latency: <100ms p99
- Throughput: 1000 req/s
- Uptime: 99.9%

### Optimization
- Connection pooling
- Redis caching
- Database indexes
- Query optimization

## Roadmap

### Phase 1: Core API (Q3 2026)
- [ ] Payment creation
- [ ] QR code generation
- [ ] Payment verification
- [ ] Basic webhooks

### Phase 2: Enhanced (Q3 2026)
- [ ] Fiat conversion
- [ ] Refund support
- [ ] Advanced webhooks
- [ ] Analytics

### Phase 3: Ecosystem (Q4 2026)
- [ ] JavaScript SDK
- [ ] WordPress plugin
- [ ] Shopify app
- [ ] Advanced reporting

## Support

**Technical**: api@vitacoin.network  
**Issues**: [GitHub Issues](https://github.com/vishwas-io/vitacoin/issues)

---

[← Back to VITAPAY](../README.md) | [API Documentation](./API.md)
