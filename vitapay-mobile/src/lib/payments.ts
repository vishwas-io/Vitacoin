import { PaymentRequest } from '../types/wallet';
import { sendVITA } from './wallet';

/**
 * parsePaymentQR
 * QR format: vitapay://pay?to={address}&amount={amount}&denom={denom}&memo={memo}&expires={timestamp}
 * Returns null if the QR data is not a valid VITAPAY payment request.
 */
export function parsePaymentQR(qrData: string): PaymentRequest | null {
  try {
    if (!qrData.startsWith('vitapay://pay')) return null;

    // Handle both vitapay://pay? and vitapay://pay/?
    const queryStart = qrData.indexOf('?');
    if (queryStart === -1) return null;
    const queryString = qrData.slice(queryStart + 1);

    const params = new URLSearchParams(queryString);

    const to = params.get('to');
    const amount = params.get('amount');
    const denom = params.get('denom') ?? 'VITA';
    const memo = params.get('memo') ?? '';
    const expiresStr = params.get('expires');

    if (!to || !amount) return null;

    // Validate address format (basic check: must start with cosmos1)
    if (!to.startsWith('cosmos1')) return null;

    // Validate amount
    const parsedAmount = parseFloat(amount);
    if (isNaN(parsedAmount) || parsedAmount <= 0) return null;

    // Parse expiry
    const expiresAt = expiresStr ? parseInt(expiresStr, 10) : Date.now() + 3600_000;
    if (isNaN(expiresAt)) return null;

    // Extract optional merchant name from memo (format: "MerchantName | Order #123")
    let merchantName = 'Merchant';
    let parsedMemo = memo;
    if (memo.includes(' | ')) {
      const [name, rest] = memo.split(' | ');
      merchantName = name.trim();
      parsedMemo = rest.trim();
    }

    return {
      merchantAddress: to,
      merchantName,
      amount: parsedAmount.toFixed(6),
      denom,
      memo: parsedMemo,
      expiresAt,
    };
  } catch {
    return null;
  }
}

/**
 * executePayment
 * Validates the payment request and executes the on-chain transaction.
 * Returns the transaction hash on success, throws on failure.
 */
export async function executePayment(
  mnemonic: string,
  request: PaymentRequest,
): Promise<string> {
  // Validate: not expired
  if (request.expiresAt < Date.now()) {
    throw new Error('Payment request has expired');
  }

  // Validate: amount > 0
  const amount = parseFloat(request.amount);
  if (isNaN(amount) || amount <= 0) {
    throw new Error('Invalid payment amount');
  }

  // Validate: valid address
  if (!request.merchantAddress || !request.merchantAddress.startsWith('cosmos1')) {
    throw new Error('Invalid merchant address');
  }

  // Build memo: include merchant name if present
  const memo = request.merchantName !== 'Merchant'
    ? `${request.merchantName} | ${request.memo}`.trim()
    : request.memo;

  // Execute on-chain send
  const txHash = await sendVITA(mnemonic, request.merchantAddress, request.amount, memo);
  return txHash;
}

/**
 * buildPaymentQR
 * Generates a vitapay:// URI for use in ReceiveScreen merchant QR (optional use).
 */
export function buildPaymentQR(params: {
  address: string;
  amount?: string;
  denom?: string;
  memo?: string;
  expiresInSeconds?: number;
}): string {
  const { address, amount, denom = 'VITA', memo = '', expiresInSeconds = 3600 } = params;
  const expires = Math.floor(Date.now() / 1000) + expiresInSeconds;
  const parts = [`vitapay://pay?to=${address}`];
  if (amount) parts.push(`amount=${amount}`);
  parts.push(`denom=${denom}`);
  if (memo) parts.push(`memo=${encodeURIComponent(memo)}`);
  parts.push(`expires=${expires * 1000}`);
  return parts.join('&');
}
