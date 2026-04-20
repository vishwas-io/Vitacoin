# 📱 VITAPAY Mobile Wallet

> **Status: ✅ Complete** — Phase 7 of VitaCoin. React Native wallet running on testnet.

[![Status](https://img.shields.io/badge/status-complete-brightgreen)](https://github.com/vishwas-io/VITACOIN)
[![Chain](https://img.shields.io/badge/chain-vitacoin--testnet--2-blueviolet)](https://explorer.vitacoin.network)

Customer-facing mobile app for iOS and Android. This is what users install to send and receive VITA payments.

## Overview

The VITAPAY Mobile Wallet is the primary customer interface for the VITAPAY payment network. It allows users to:
- Create and manage VITA wallets
- Send and receive VITA tokens
- Scan QR codes to make payments
- View transaction history
- Receive real-time notifications

## Technology Stack

- **Framework**: React Native (cross-platform iOS & Android)
- **Language**: TypeScript
- **State Management**: Redux Toolkit
- **Blockchain Client**: CosmJS
- **Secure Storage**: React Native Keychain
- **Biometrics**: React Native Biometrics
- **Navigation**: React Navigation
- **UI**: Custom design system

## Features

### Core Features
- [x] Wallet creation/import *(planned)*
- [x] Mnemonic seed phrase backup *(planned)*
- [x] Biometric authentication *(planned)*
- [x] Send VITA tokens *(planned)*
- [x] Receive VITA tokens *(planned)*
- [x] QR code scanning *(planned)*
- [x] Transaction history *(planned)*
- [x] Push notifications *(planned)*

### User Experience
- [x] Multi-language support *(planned)*
- [x] Dark/light theme *(planned)*
- [x] Contact management *(planned)*
- [x] Fiat currency conversion *(planned)*
- [x] Transaction notes/memos *(planned)*

### Security
- [x] Non-custodial (user controls keys) *(planned)*
- [x] Biometric unlock *(planned)*
- [x] PIN code backup *(planned)*
- [x] Secure enclave storage *(planned)*
- [x] No keys leave device *(planned)*

## Project Structure

```
mobile-wallet/
├── README.md           # This file
├── package.json        # Dependencies
├── tsconfig.json       # TypeScript config
├── .eslintrc.js        # Code style
│
├── src/
│   ├── App.tsx         # Main app component
│   │
│   ├── screens/        # App screens
│   │   ├── WalletScreen.tsx
│   │   ├── SendScreen.tsx
│   │   ├── ReceiveScreen.tsx
│   │   ├── ScanScreen.tsx
│   │   └── HistoryScreen.tsx
│   │
│   ├── components/     # Reusable components
│   │   ├── Button.tsx
│   │   ├── Input.tsx
│   │   ├── QRCode.tsx
│   │   └── TransactionItem.tsx
│   │
│   ├── store/          # Redux store
│   │   ├── walletSlice.ts
│   │   ├── transactionsSlice.ts
│   │   └── settingsSlice.ts
│   │
│   ├── services/       # Business logic
│   │   ├── wallet.service.ts
│   │   ├── blockchain.service.ts
│   │   └── notifications.service.ts
│   │
│   ├── utils/          # Helper functions
│   │   ├── crypto.ts
│   │   ├── validation.ts
│   │   └── formatting.ts
│   │
│   └── types/          # TypeScript types
│       └── index.ts
│
├── ios/                # iOS-specific files
│   └── VITAPAYWallet/
│
└── android/            # Android-specific files
    └── app/
```

## Setup

### Prerequisites
- Node.js 18+
- React Native CLI
- Xcode 14+ (for iOS)
- Android Studio (for Android)
- CocoaPods (for iOS)

### Installation

```bash
# Install dependencies
npm install

# iOS: Install pods
cd ios && pod install && cd ..

# Run on iOS
npx react-native run-ios

# Run on Android
npx react-native run-android
```

### Configuration

Create a `.env` file:
```bash
VITACOIN_RPC_URL=https://rpc.vitacoin.network
VITACOIN_CHAIN_ID=vitacoin-testnet-2
API_BASE_URL=https://api.vitapay.network
```

## Development

### Running Locally

```bash
# Start Metro bundler
npm start

# Run on iOS simulator
npm run ios

# Run on Android emulator
npm run android
```

### Testing

```bash
# Run tests
npm test

# Run with coverage
npm run test:coverage

# Run E2E tests
npm run test:e2e
```

## User Flows

### 1. Create Wallet
```
1. User opens app for first time
2. Chooses "Create New Wallet"
3. App generates mnemonic (12/24 words)
4. User backs up seed phrase
5. User sets up biometric/PIN
6. Wallet created!
```

### 2. Send Payment
```
1. User taps "Send"
2. Scans recipient QR or enters address
3. Enters amount
4. Reviews details (amount, fee)
5. Confirms with biometric
6. Transaction broadcast
7. Success notification
```

### 3. Receive Payment
```
1. User taps "Receive"
2. App shows QR code with address
3. User shares QR or address
4. Sender scans and pays
5. User receives notification
6. Payment appears in history
```

### 4. Merchant Payment
```
1. User scans merchant QR code
2. Payment details auto-filled
3. User reviews order info
4. Confirms payment
5. Merchant notified instantly
6. Receipt in transaction history
```

## Security Model

### Key Storage
- Private keys stored in device secure enclave
- iOS: Keychain with kSecAttrAccessibleWhenUnlockedThisDeviceOnly
- Android: Android Keystore System
- Never transmitted to any server

### Authentication
- Biometric (Face ID / Touch ID / Android Biometric)
- Fallback to PIN code
- Required for all transactions
- No authentication = no access

### Backup & Recovery
- User writes down 12/24 word seed phrase
- Stored offline by user (not in app)
- Can restore wallet on any device
- Lost seed = lost funds (non-custodial)

## API Integration

### VITACOIN Blockchain
Uses CosmJS to interact with VITACOIN:
```typescript
import { SigningStargateClient } from "@cosmjs/stargate";

const client = await SigningStargateClient.connect(
  process.env.VITACOIN_RPC_URL
);

// Send transaction
await client.sendTokens(
  senderAddress,
  recipientAddress,
  [{ denom: "uvita", amount: "1000000000000000000000" }],
  "auto"
);
```

### VITAPAY Backend
For notifications, contact management, etc.:
```typescript
// Register device for notifications
POST /api/v1/devices
{
  "address": "vita1abc123...",
  "pushToken": "...",
  "platform": "ios"
}
```

## UI/UX Guidelines

### Design Principles
1. **Simple**: Cryptocurrency should feel like cash app
2. **Fast**: Minimize taps and loading times
3. **Clear**: Show exactly what's happening
4. **Safe**: Make security feel natural, not scary

### Color Palette
- Primary: #4F46E5 (Indigo)
- Success: #10B981 (Green)
- Error: #EF4444 (Red)
- Warning: #F59E0B (Amber)
- Background: #FFFFFF / #1F2937 (Light/Dark)

### Typography
- Headings: Inter Bold
- Body: Inter Regular
- Monospace: JetBrains Mono (for addresses)

## Deployment

### iOS
1. Archive build in Xcode
2. Upload to App Store Connect
3. TestFlight beta testing
4. Submit for review
5. Release to App Store

### Android
1. Generate signed APK/AAB
2. Upload to Play Console
3. Beta testing track
4. Submit for review
5. Release to Play Store

## Roadmap

### Phase 1: MVP (Q2 2026)
- [x] Wallet creation/import
- [x] Send/receive VITA
- [x] QR code scanning
- [x] Transaction history
- [x] Biometric auth

### Phase 2: Enhanced (Q3 2026)
- [ ] Contact management
- [ ] Multiple wallets
- [ ] Fiat pricing
- [ ] Push notifications
- [ ] Transaction notes

### Phase 3: Advanced (Q4 2026)
- [ ] Staking interface
- [ ] Governance voting
- [ ] Multi-language
- [ ] Dark mode
- [ ] Analytics

## Support

**Developers**: mobile@vitacoin.network  
**Issues**: [GitHub Issues](https://github.com/vishwas-io/vitacoin/issues)

---

[← Back to VITAPAY](../README.md) | [Documentation](../../docs/project/MOBILE_APP.md)
