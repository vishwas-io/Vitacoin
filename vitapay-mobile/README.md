# VITAPAY Mobile Wallet

**VITAPAY** is the official mobile wallet for the VitaCoin blockchain — a Cosmos SDK-based payment chain built for instant, low-cost merchant payments.

---

## What is VITAPAY?

VITAPAY lets you:
- **Create or import** a Cosmos HD wallet (24-word mnemonic, BIP44)
- **Send VITA** to any `vita1...` address with memo support
- **Receive VITA** via shareable QR code (your address)
- **Pay merchants** by scanning a `vitapay://` payment QR code
- **Stake VITA** with network validators, claim staking rewards
- **View transaction history** from the VitaCoin blockchain

---

## Tech Stack

| Layer | Technology |
|---|---|
| Framework | React Native + Expo (SDK 50+) |
| Blockchain client | CosmJS (`@cosmjs/proto-signing`, `@cosmjs/stargate`) |
| Secure storage | `expo-secure-store` (mnemonic encrypted at rest) |
| Camera / QR scan | `expo-camera` (`CameraView` + `BarcodeScanningResult`) |
| QR generation | `react-native-qrcode-svg` |
| Clipboard | `expo-clipboard` |
| Navigation | `@react-navigation/native` + `@react-navigation/bottom-tabs` |
| TypeScript | Strict mode, no `any` in lib code |

---

## How to Run (Development)

### Prerequisites
- Node.js 18+
- Expo CLI: `npm install -g expo-cli`
- iOS: Xcode 15+ / Android: Android Studio

### Install dependencies
```bash
cd vitapay-mobile
npm install
```

### Start development server
```bash
npx expo start
```

Then scan the QR code with **Expo Go** on your phone, or press `i` (iOS simulator) / `a` (Android emulator).

---

## How to Build (Production)

### Using EAS Build (recommended)
```bash
npm install -g eas-cli
eas login
eas build --platform ios      # iOS .ipa
eas build --platform android  # Android .apk / .aab
eas build --platform all      # Both
```

### Preview build (APK, no store)
```bash
eas build --profile preview --platform android
```

Configure builds in `eas.json`.

---

## Screens Overview

| Screen | Description |
|---|---|
| `WalletSetupScreen` | Create new wallet (generates 24-word mnemonic) or import existing one |
| `HomeScreen` | Balance dashboard — VITA balance, quick action buttons |
| `SendScreen` | Send VITA to any address with amount + memo |
| `ReceiveScreen` | Show your address as QR code + copy/share buttons |
| `PayScreen` | Scan merchant QR (`vitapay://pay?...`), review payment, confirm |
| `TransactionHistoryScreen` | Live transaction history pulled from chain REST API |
| `StakeScreen` | Stake VITA with validators, view delegations, claim rewards |

---

## QR Payment Format

Merchant QR codes use the `vitapay://` URI scheme:

```
vitapay://pay?to={address}&amount={amount}&denom=VITA&memo={memo}&expires={unixMs}
```

**Example:**
```
vitapay://pay?to=vita1merchant9x8y...&amount=12.5&denom=VITA&memo=VitaCafe%20%7C%20Order%20%231042&expires=1712500000000
```

Parse with `parsePaymentQR(qrData)` in `src/lib/payments.ts`.

---

## Project Structure

```
vitapay-mobile/
├── App.tsx                    ← Entry point, navigation setup
├── src/
│   ├── constants/
│   │   └── config.ts          ← Chain ID, RPC/REST endpoints, denom
│   ├── lib/
│   │   ├── wallet.ts          ← generateWallet, importWallet, sendVITA, getBalance, getTransactionHistory
│   │   ├── staking.ts         ← getValidators, getDelegations, delegateVITA, undelegateVITA, claimRewards
│   │   ├── payments.ts        ← parsePaymentQR, executePayment, buildPaymentQR
│   │   └── storage.ts         ← saveMnemonic, getMnemonic, saveAddress, getAddress (SecureStore)
│   ├── screens/
│   │   ├── WalletSetupScreen.tsx
│   │   ├── HomeScreen.tsx
│   │   ├── SendScreen.tsx
│   │   ├── ReceiveScreen.tsx
│   │   ├── PayScreen.tsx
│   │   ├── TransactionHistoryScreen.tsx
│   │   └── StakeScreen.tsx
│   ├── navigation/            ← Bottom tab navigator
│   └── types/
│       └── wallet.ts          ← WalletAccount, Transaction, PaymentRequest types
├── package.json
├── tsconfig.json
└── README.md
```

---

## Chain Configuration

Edit `src/constants/config.ts`:

| Key | Value |
|---|---|
| `CHAIN_ID` | `vitacoin-testnet-2` |
| `RPC_ENDPOINT` | `https://rpc.vitacoin.network` |
| `REST_ENDPOINT` | `https://api.vitacoin.network` |
| `DENOM` | `uvita` (micro-VITA, 6 decimals) |
| `BECH32_PREFIX` | `vita` |

---

## Security

- Mnemonic is stored encrypted via `expo-secure-store` (device Keychain / Keystore)
- Never transmitted over network
- Payment QR codes include expiry (`expires` field) — expired QRs are rejected
- This repo is **PUBLIC** — never commit real mnemonics or private keys

---

## Status

Phase 7 of VitaCoin (VITAPAY Mobile Wallet) — **Complete** ✅

All 9 phases of VitaCoin are now complete. Testnet is LIVE: chain-id `vitacoin-testnet-2`.
