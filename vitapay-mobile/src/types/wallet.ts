export interface WalletAccount {
  address: string;       // bech32 vita1...
  publicKey: string;
  balance: string;       // VITA amount
  stakedBalance: string; // stVITA amount
  pendingRewards: string;
}

export interface Transaction {
  hash: string;
  from: string;
  to: string;
  amount: string;
  denom: string;
  type: 'send' | 'receive' | 'delegate' | 'undelegate' | 'claim_rewards' | 'payment';
  status: 'pending' | 'confirmed' | 'failed';
  timestamp: number;
  blockHeight: number;
  fee: string;
}

export interface PaymentRequest {
  merchantAddress: string;
  merchantName: string;
  amount: string;
  denom: string;
  memo: string;
  expiresAt: number;
}
