# 💳 VITAPAY - The Payment Network

> **Status: ✅ Complete** — All VITAPAY components are built and running on testnet.

[![Status](https://img.shields.io/badge/status-complete-brightgreen)](https://github.com/vishwas-io/VITACOIN)
[![Testnet](https://img.shields.io/badge/testnet-LIVE-brightgreen)](https://explorer.vitacoin.network)
[![Chain](https://img.shields.io/badge/chain-vitacoin--testnet--2-blueviolet)](https://vitacoin.network)
[![Discord](https://img.shields.io/badge/discord-join-7289da)](https://discord.gg/9JsRPwDzg)

## What is This?

**VITAPAY** is a complete payment ecosystem that makes cryptocurrency payments as easy as using PayPal:
- **Mobile Wallet**: iOS & Android apps for customers
- **Payment Gateway**: APIs for merchant integration
- **Merchant Dashboard**: Web portal for business management
- **QR Code System**: Simple scan-to-pay checkout

**Think of it as**: The "PayPal for crypto" - user-friendly payment tools powered by VITACOIN.

## Relationship to VITACOIN

```
VITAPAY (this folder)
    ↓ uses
VITACOIN (cryptocurrency blockchain)
```

VITAPAY provides the user experience layer on top of the VITACOIN blockchain. When users make payments through VITAPAY, the actual value transfer happens on the VITACOIN blockchain.

## What's In This Folder?

```
vitapay/
├── README.md                    # This file
│
├── mobile-wallet/               # Customer-facing mobile app
│   ├── README.md               # Mobile app docs
│   ├── src/                    # React Native source code
│   ├── ios/                    # iOS-specific files
│   ├── android/                # Android-specific files
│   └── package.json            # Dependencies
│
├── payment-gateway/             # Merchant payment APIs
│   ├── README.md               # Gateway docs
│   ├── api/                    # REST API handlers
│   ├── webhooks/               # Webhook system
│   ├── qr-generator/           # QR code generation
│   └── go.mod                  # Go dependencies
│
├── merchant-dashboard/          # Web portal for merchants
│   ├── README.md               # Dashboard docs
│   ├── src/                    # React/Next.js source
│   ├── pages/                  # Dashboard pages
│   └── package.json            # Dependencies
│
└── shared/                      # Shared utilities
    └── vitacoin-client/        # Client SDK for VITACOIN blockchain
        ├── README.md
        ├── connection.ts       # Blockchain connection
        ├── transactions.ts     # Transaction helpers
        └── wallet.ts           # Wallet operations
```

## Quick Start

### Prerequisites

**For Mobile Wallet:**
- Node.js 18+
- React Native CLI
- Xcode (for iOS) / Android Studio (for Android)

**For Payment Gateway:**
- Go 1.21+
- PostgreSQL 15+
- Redis 7+

**For Merchant Dashboard:**
- Node.js 18+
- Next.js 14+

### Setup Mobile Wallet

```bash
cd mobile-wallet

# Install dependencies
npm install

# iOS
npx pod-install
npx react-native run-ios

# Android
npx react-native run-android
```

[Full Mobile Wallet Guide →](./mobile-wallet/README.md)

### Setup Payment Gateway

```bash
cd payment-gateway

# Install dependencies
go mod download

# Setup database
make migrate-up

# Start server
make run

# API will be available at http://localhost:8080
```

[Full Payment Gateway Guide →](./payment-gateway/README.md)

### Setup Merchant Dashboard

```bash
cd merchant-dashboard

# Install dependencies
npm install

# Start development server
npm run dev

# Dashboard available at http://localhost:3000
```

[Full Dashboard Guide →](./merchant-dashboard/README.md)

## Architecture

### High-Level Flow

```
┌─────────────────────────────────────────────────┐
│              CUSTOMER JOURNEY                    │
├─────────────────────────────────────────────────┤
│                                                  │
│  1. Customer opens VITAPAY Mobile Wallet        │
│  2. Scans merchant QR code                      │
│  3. Reviews payment details                     │
│  4. Confirms with biometric/PIN                 │
│  5. VITAPAY sends VITACOIN transaction          │
│  6. Receives confirmation in ~5 seconds         │
│                                                  │
└─────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────┐
│             MERCHANT JOURNEY                     │
├─────────────────────────────────────────────────┤
│                                                  │
│  1. Merchant integrates Payment Gateway API     │
│  2. Generates payment request                   │
│  3. Displays QR code at checkout                │
│  4. Receives webhook notification              │
│  5. Fulfills order instantly                    │
│                                                  │
└─────────────────────────────────────────────────┘
```

### Technology Stack

#### Mobile Wallet
- **Framework**: React Native
- **Language**: TypeScript
- **State**: Redux Toolkit
- **Blockchain**: CosmJS
- **Security**: React Native Keychain
- **Biometrics**: React Native Biometrics

#### Payment Gateway
- **Language**: Go
- **Framework**: Gin (REST API)
- **Database**: PostgreSQL
- **Cache**: Redis
- **Queue**: RabbitMQ
- **Auth**: JWT

#### Merchant Dashboard
- **Framework**: Next.js 14
- **Language**: TypeScript
- **UI**: Tailwind CSS + shadcn/ui
- **State**: React Query
- **Charts**: Recharts

#### Shared Client SDK
- **Language**: TypeScript
- **Purpose**: Abstract VITACOIN blockchain interactions
- **Used By**: All VITAPAY components

## Components

### 1. Mobile Wallet (`mobile-wallet/`)

**Customer-facing mobile app** for iOS and Android.

**Features:**
- ✅ Create/import wallets
- ✅ Send/receive VITA tokens
- ✅ Scan QR codes for payments
- ✅ Transaction history
- ✅ Biometric authentication
- ✅ Multi-language support
- ✅ Real-time notifications

**Target Users:**
- Crypto-native early adopters
- Users seeking low-fee payments
- Global remittance users

**Development Status:** ✅ Complete

[Mobile Wallet Details →](./mobile-wallet/README.md)

### 2. Payment Gateway (`payment-gateway/`)

**Merchant-facing APIs** for accepting VITA payments.

**Features:**
- ✅ Payment request generation
- ✅ QR code creation
- ✅ Webhook notifications
- ✅ Payment verification
- ✅ Refund handling (if needed)
- ✅ Multi-currency pricing (VITA + fiat)
- ✅ API key management

**Integration Options:**
- REST API
- JavaScript SDK
- WordPress plugin (future)
- Shopify app (future)

**Development Status:** ✅ Complete

[Payment Gateway Details →](./payment-gateway/README.md)

### 3. Merchant Dashboard (`merchant-dashboard/`)

**Web portal** for merchants to manage their payments.

**Features:**
- ✅ Transaction monitoring
- ✅ Analytics & reports
- ✅ Payment history
- ✅ API key management
- ✅ Webhook configuration
- ✅ Settlement tracking
- ✅ Customer insights

**Development Status:** ✅ Complete

[Merchant Dashboard Details →](./merchant-dashboard/README.md)

### 4. Shared Client SDK (`shared/vitacoin-client/`)

**TypeScript/JavaScript SDK** for interacting with VITACOIN blockchain.

**Features:**
- ✅ Connect to VITACOIN nodes
- ✅ Sign transactions
- ✅ Query balances
- ✅ Send tokens
- ✅ Listen for events
- ✅ Handle errors gracefully

**Used By:**
- Mobile Wallet (transaction signing)
- Payment Gateway (payment verification)
- Merchant Dashboard (data display)

**Development Status:** ✅ Complete

## Development

### Project Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/vishwas-io/vitacoin
   cd vitacoin/vitapay
   ```

2. **Install dependencies for each component**
   ```bash
   # Mobile wallet
   cd mobile-wallet && npm install
   
   # Payment gateway
   cd ../payment-gateway && go mod download
   
   # Merchant dashboard
   cd ../merchant-dashboard && npm install
   ```

3. **Setup local VITACOIN node**
   ```bash
   # From the vitacoin/ directory
   cd ../vitacoin
   make install
   make localnet
   ```

4. **Configure environment**
   ```bash
   # Each component has .env.example
   cp .env.example .env
   # Edit with your settings
   ```

### Running Locally

**Terminal 1 - VITACOIN Node:**
```bash
cd vitacoin
make localnet
```

**Terminal 2 - Payment Gateway:**
```bash
cd vitapay/payment-gateway
make run
```

**Terminal 3 - Merchant Dashboard:**
```bash
cd vitapay/merchant-dashboard
npm run dev
```

**Terminal 4 - Mobile Wallet:**
```bash
cd vitapay/mobile-wallet
npx react-native start
# In another terminal:
npx react-native run-ios  # or run-android
```

### Testing

```bash
# Mobile wallet tests
cd mobile-wallet
npm test

# Payment gateway tests
cd payment-gateway
go test ./...

# Merchant dashboard tests
cd merchant-dashboard
npm test

# End-to-end tests
cd ../
npm run test:e2e
```

## API Documentation

### Payment Gateway REST API

**Base URL**: `https://api.vitapay.network/v1`

**Endpoints:**
- `POST /payments` - Create payment request
- `GET /payments/:id` - Get payment status
- `POST /webhooks` - Register webhook
- `GET /webhooks` - List webhooks
- `DELETE /webhooks/:id` - Delete webhook

[Full API Documentation →](./payment-gateway/API.md) (Coming soon)

### Mobile Wallet Deep Links

**Format**: `vitapay://pay?recipient=<address>&amount=<amount>&memo=<memo>`

**Example:**
```
vitapay://pay?recipient=vita1abc123&amount=100&memo=Order123
```

## Security

### Mobile Wallet Security
- Private keys stored in secure enclave
- Biometric authentication required
- No keys transmitted to servers
- Open source for auditability

### Payment Gateway Security
- API key authentication
- Rate limiting
- HTTPS only
- Webhook signature verification
- PCI DSS compliance (future)

### Best Practices
- Never expose API keys
- Validate all webhooks
- Use HTTPS for all requests
- Implement rate limiting
- Log all transactions

[Security Guide →](../docs/architecture/SECURITY.md)

## Deployment

### Mobile Wallet
- **iOS**: TestFlight → App Store
- **Android**: Play Store Beta → Production

### Payment Gateway
- **Platform**: AWS / Google Cloud
- **Container**: Docker + Kubernetes
- **Database**: RDS PostgreSQL
- **Cache**: ElastiCache Redis
- **Load Balancer**: ALB

### Merchant Dashboard
- **Platform**: Vercel / Netlify
- **CDN**: CloudFront
- **Analytics**: PostHog

[Deployment Guide →](../docs/deployment/README.md) (Coming soon)

## Roadmap

### Phase 1: Mobile Wallet MVP ✨ (Q2 2026)
- [ ] Wallet creation/import
- [ ] Send/receive VITA
- [ ] QR code scanning
- [ ] Transaction history
- [ ] iOS & Android apps

### Phase 2: Payment Gateway ✨ (Q3 2026)
- [ ] REST API
- [ ] Payment requests
- [ ] QR code generation
- [ ] Webhook system
- [ ] JavaScript SDK

### Phase 3: Merchant Dashboard ✨ (Q3 2026)
- [ ] Transaction dashboard
- [ ] Analytics
- [ ] API key management
- [ ] Settings & configuration

### Phase 4: Ecosystem Expansion 🚀 (Q4 2026)
- [ ] E-commerce plugins (Shopify, WooCommerce)
- [ ] Point-of-sale terminals
- [ ] Merchant mobile app
- [ ] Advanced analytics

[Full Roadmap →](../docs/project/DEVELOPMENT_ROADMAP.md)

## Contributing

We welcome contributions! Here's how:

1. Check [TODO.md](../docs/project/TODO.md) for tasks
2. Pick a component to work on
3. Follow the style guide in each folder
4. Submit a pull request

[Contributing Guide →](../CONTRIBUTING.md) (Coming soon)

## Resources

### Documentation
- [VITAPAY Guide](../docs/project/VITAPAY.md) - Payment network overview
- [Mobile App Specs](../docs/project/MOBILE_APP.md) - Detailed app specifications
- [VITACOIN Blockchain](../vitacoin/README.md) - Underlying blockchain

### External Resources
- [React Native Docs](https://reactnative.dev/)
- [Next.js Docs](https://nextjs.org/docs)
- [CosmJS Docs](https://cosmos.github.io/cosmjs/)

### Community
- **GitHub**: [github.com/vishwas-io/vitacoin](https://github.com/vishwas-io/vitacoin)
- **Discord**: Coming soon
- **Forum**: Coming soon

## Support

- **Integration Help**: payments@vitacoin.network
- **Bug Reports**: Open an issue on GitHub
- **Partnership**: partners@vitacoin.network

## License

Apache 2.0 - See [LICENSE](../LICENSE)

---

**This is the VITAPAY payment network.** For the blockchain, see [../vitacoin/](../vitacoin/)
