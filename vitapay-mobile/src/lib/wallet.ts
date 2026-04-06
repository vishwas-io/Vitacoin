import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import { SigningStargateClient, StargateClient, GasPrice, coin } from '@cosmjs/stargate';
import { CONFIG } from '../constants/config';

export interface Transaction {
  txHash: string;
  height: number;
  timestamp: string;
  type: string;
  amount: string;
  fee: string;
  memo: string;
  success: boolean;
}

export async function generateWallet(): Promise<{ mnemonic: string; address: string }> {
  const wallet = await DirectSecp256k1HdWallet.generate(24, { prefix: CONFIG.BECH32_PREFIX });
  const [account] = await wallet.getAccounts();
  return { mnemonic: wallet.mnemonic, address: account.address };
}

export async function importWallet(mnemonic: string): Promise<{ address: string }> {
  const trimmed = mnemonic.trim();
  const wordCount = trimmed.split(/\s+/).length;
  if (wordCount !== 12 && wordCount !== 24) {
    throw new Error('Invalid mnemonic — must be 12 or 24 words');
  }
  const wallet = await DirectSecp256k1HdWallet.fromMnemonic(trimmed, { prefix: CONFIG.BECH32_PREFIX });
  const [account] = await wallet.getAccounts();
  return { address: account.address };
}

export async function getBalance(address: string): Promise<string> {
  try {
    const url = `${CONFIG.REST_ENDPOINT}/cosmos/bank/v1beta1/balances/${address}`;
    const res = await fetch(url);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    const balances: Array<{ denom: string; amount: string }> = data.balances ?? [];
    const vitaBalance = balances.find((b) => b.denom === CONFIG.DENOM);
    if (!vitaBalance) return '0.000000';
    const raw = parseInt(vitaBalance.amount, 10);
    return (raw / Math.pow(10, CONFIG.DENOM_DECIMALS)).toFixed(6);
  } catch {
    return '0.000000';
  }
}

export async function sendVITA(
  mnemonic: string,
  toAddress: string,
  amount: string,
  memo: string,
): Promise<string> {
  const wallet = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic.trim(), {
    prefix: CONFIG.BECH32_PREFIX,
  });
  const [sender] = await wallet.getAccounts();

  const client = await SigningStargateClient.connectWithSigner(CONFIG.RPC_ENDPOINT, wallet, {
    gasPrice: GasPrice.fromString(CONFIG.GAS_PRICE),
  });

  const microAmount = Math.floor(parseFloat(amount) * Math.pow(10, CONFIG.DENOM_DECIMALS));
  const result = await client.sendTokens(
    sender.address,
    toAddress,
    [coin(microAmount, CONFIG.DENOM)],
    'auto',
    memo,
  );

  if (result.code !== 0) {
    throw new Error(`Transaction failed: ${result.rawLog}`);
  }
  return result.transactionHash;
}

export async function getTransactionHistory(address: string): Promise<Transaction[]> {
  try {
    const url = `${CONFIG.REST_ENDPOINT}/cosmos/tx/v1beta1/txs?events=message.sender%3D%27${address}%27&limit=20`;
    const res = await fetch(url);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    const txResponses: any[] = data.tx_responses ?? [];

    return txResponses.map((tx: any) => {
      // Extract first transfer message if present
      const messages: any[] = tx.tx?.body?.messages ?? [];
      const firstMsg = messages[0];
      const typeUrl: string = firstMsg?.['@type'] ?? '';
      const msgType = typeUrl.split('.').pop() ?? 'Unknown';

      // Try to get send amount
      let amount = '0';
      if (firstMsg?.amount) {
        const amt = Array.isArray(firstMsg.amount) ? firstMsg.amount[0] : firstMsg.amount;
        if (amt?.denom === CONFIG.DENOM) {
          amount = (parseInt(amt.amount, 10) / Math.pow(10, CONFIG.DENOM_DECIMALS)).toFixed(6);
        }
      }

      // Fee
      const feeCoins: any[] = tx.tx?.auth_info?.fee?.amount ?? [];
      const feeCoin = feeCoins.find((c: any) => c.denom === CONFIG.DENOM);
      const fee = feeCoin
        ? (parseInt(feeCoin.amount, 10) / Math.pow(10, CONFIG.DENOM_DECIMALS)).toFixed(6)
        : '0';

      return {
        txHash: tx.txhash ?? '',
        height: parseInt(tx.height ?? '0', 10),
        timestamp: tx.timestamp ?? '',
        type: msgType,
        amount,
        fee,
        memo: tx.tx?.body?.memo ?? '',
        success: tx.code === 0 || tx.code === undefined,
      };
    });
  } catch {
    return [];
  }
}

export async function estimateFee(amount: string): Promise<string> {
  // Estimate based on gas price: ~80000 gas for a send tx
  const gasEstimate = 80000;
  const gasPrice = 0.025; // uvita per gas
  const feeUvita = gasEstimate * gasPrice;
  const feeVita = feeUvita / Math.pow(10, CONFIG.DENOM_DECIMALS);
  return feeVita.toFixed(6);
}
