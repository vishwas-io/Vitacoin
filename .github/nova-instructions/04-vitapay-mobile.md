# VITAPAY Mobile Wallet — React Native Standards

## Stack
- React Native 0.81.5
- Expo SDK ~54
- TypeScript 5.9.2
- React 19
- CosmJS (`@cosmjs/stargate` ^0.32.4, `@cosmjs/proto-signing` ^0.32.4)

## Always Check Context7 Before Writing CosmJS / Expo Code
```bash
mcporter call context7.resolve-library-id --args '{"libraryName": "cosmjs"}'
mcporter call context7.resolve-library-id --args '{"libraryName": "expo"}'
```

---

## The UX Bar — Phantom / Rainbow Level

VITAPAY is the face of VitaCoin. First-time users judge the entire blockchain by this app.
Reference: **Phantom wallet (Solana)**, **Rainbow wallet (Ethereum)**, **Revolut** for payments.

Every screen must feel:
- **Instant** — no loading spinners on things that should be cached
- **Trustworthy** — security-forward design (biometric prompts feel premium not annoying)
- **Beautiful** — dark, premium, crypto-native aesthetic

---

## Wallet Architecture

```
src/
├── screens/
│   ├── WelcomeScreen.tsx      ← first launch
│   ├── CreateWalletScreen.tsx ← generate HD wallet
│   ├── ImportWalletScreen.tsx ← restore from mnemonic
│   ├── DashboardScreen.tsx    ← balance, recent txs
│   ├── SendScreen.tsx         ← send VITA
│   ├── ReceiveScreen.tsx      ← QR code
│   ├── ScanScreen.tsx         ← scan vitapay:// QR
│   ├── StakingScreen.tsx      ← delegate/undelegate
│   └── SettingsScreen.tsx     ← security, backup
├── services/
│   ├── wallet.ts              ← HD wallet creation, CosmJS signer
│   ├── chain.ts               ← StargateClient, tx broadcast
│   ├── storage.ts             ← SecureStore wrapper
│   └── vitapay.ts             ← vitapay:// URI parser
└── components/
    ├── BalanceCard.tsx
    ├── TransactionItem.tsx
    ├── QRScanner.tsx
    └── BiometricPrompt.tsx
```

---

## Wallet Service Standards

```typescript
// ✅ ALWAYS use HD wallet derivation (BIP44)
import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';

export async function createWallet(): Promise<{ mnemonic: string; address: string }> {
  const wallet = await DirectSecp256k1HdWallet.generate(24, { prefix: 'vita' });
  const [account] = await wallet.getAccounts();
  return { mnemonic: wallet.mnemonic, address: account.address };
}

// ✅ ALWAYS store mnemonic in SecureStore — never AsyncStorage
import * as SecureStore from 'expo-secure-store';
await SecureStore.setItemAsync('mnemonic', mnemonic, {
  keychainAccessibility: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
});

// ✅ ALWAYS require biometric before reading mnemonic or signing
import * as LocalAuth from 'expo-local-authentication';
const auth = await LocalAuth.authenticateAsync({ promptMessage: 'Confirm' });
if (!auth.success) throw new Error('Auth failed');
```

---

## Chain Client Standards

```typescript
// ✅ Use StargateClient for queries (read-only)
import { StargateClient } from '@cosmjs/stargate';
const client = await StargateClient.connect(RPC_ENDPOINT);
const balance = await client.getBalance(address, 'uvita');

// ✅ Use SigningStargateClient for transactions
import { SigningStargateClient } from '@cosmjs/stargate';
const signer = await loadWalletSigner(); // reads from SecureStore + biometric
const signingClient = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, signer);

// ✅ ALWAYS handle broadcast errors gracefully
try {
  const result = await signingClient.sendTokens(from, to, amount, fee, memo);
  if (result.code !== 0) throw new Error(`TX failed: ${result.rawLog}`);
} catch (err) {
  // Show user-friendly error — never raw chain error
  setError(parseChainError(err));
}

// ✅ ALWAYS set appropriate gas
const fee = { amount: [{ denom: 'uvita', amount: '5000' }], gas: '200000' };
```

---

## Payment QR Standard (`vitapay://`)

```typescript
// URI format: vitapay://<address>?amount=<uvita>&memo=<optional>
// Example: vitapay://vita1abc...xyz?amount=1000000&memo=Order123

export function parseVitaPayURI(uri: string): PaymentRequest {
  if (!uri.startsWith('vitapay://')) throw new Error('Invalid QR code');
  const url = new URL(uri.replace('vitapay://', 'https://'));
  const address = url.hostname;
  const amount = url.searchParams.get('amount');
  const memo = url.searchParams.get('memo') ?? '';

  if (!address || !isValidVitaAddress(address)) throw new Error('Invalid address');
  if (!amount || parseInt(amount) <= 0) throw new Error('Invalid amount');

  return { address, amount: parseInt(amount), memo };
}

// ✅ ALWAYS show confirmation screen before executing payment
// ✅ ALWAYS show amount in VITA (not uvita) to user
// 1 VITA = 1,000,000 uvita
export function formatVITA(uvita: number): string {
  return (uvita / 1_000_000).toFixed(6) + ' VITA';
}
```

---

## Screen Quality Checklist

Every screen must have:
- [ ] Loading state (skeleton loader matching layout)
- [ ] Error state (human-readable message + retry button)
- [ ] Empty state (icon + message + CTA if applicable)
- [ ] Biometric gate on sensitive actions
- [ ] Haptic feedback on primary actions (`expo-haptics`)
- [ ] Safe area insets respected
- [ ] Works on both iOS and Android

---

## Performance Rules

- Use `React.memo` on all list items (`TransactionItem`, etc.)
- Paginate transaction history — never load all txs
- Cache chain queries in component state — don't re-fetch on every render
- Debounce address input validation (300ms)
- Lazy load heavy screens (staking, settings)
