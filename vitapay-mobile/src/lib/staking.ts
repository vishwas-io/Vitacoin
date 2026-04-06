import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import { SigningStargateClient, GasPrice, coin } from '@cosmjs/stargate';
import { CONFIG } from '../constants/config';

export interface Delegation {
  validatorAddress: string;
  validatorName: string;
  delegatedAmount: string;
  pendingRewards: string;
}

export interface Validator {
  operatorAddress: string;
  moniker: string;
  status: string;
  commission: string;
  votingPower: string;
  jailed: boolean;
}

export async function getDelegations(address: string): Promise<Delegation[]> {
  try {
    const url = `${CONFIG.REST_ENDPOINT}/cosmos/staking/v1beta1/delegations/${address}`;
    const res = await fetch(url);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    const delegationResponses: any[] = data.delegation_responses ?? [];

    const results: Delegation[] = [];
    for (const d of delegationResponses) {
      const valAddr: string = d.delegation?.validator_address ?? '';
      const shares: string = d.delegation?.shares ?? '0';
      const balance = d.balance;
      const delegatedAmount = balance?.denom === CONFIG.DENOM
        ? (parseInt(balance.amount, 10) / Math.pow(10, CONFIG.DENOM_DECIMALS)).toFixed(6)
        : shares;

      // Fetch pending rewards for this validator
      let pendingRewards = '0.000000';
      try {
        const rewardUrl = `${CONFIG.REST_ENDPOINT}/cosmos/distribution/v1beta1/delegators/${address}/rewards/${valAddr}`;
        const rewardRes = await fetch(rewardUrl);
        if (rewardRes.ok) {
          const rewardData = await rewardRes.json();
          const rewardCoins: any[] = rewardData.rewards ?? [];
          const vitaReward = rewardCoins.find((c: any) => c.denom === CONFIG.DENOM);
          if (vitaReward) {
            pendingRewards = (parseFloat(vitaReward.amount) / Math.pow(10, CONFIG.DENOM_DECIMALS)).toFixed(6);
          }
        }
      } catch { /* ignore */ }

      results.push({
        validatorAddress: valAddr,
        validatorName: valAddr,
        delegatedAmount,
        pendingRewards,
      });
    }
    return results;
  } catch {
    return [];
  }
}

export async function getValidators(): Promise<Validator[]> {
  try {
    const url = `${CONFIG.REST_ENDPOINT}/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED&pagination.limit=100`;
    const res = await fetch(url);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    const validators: any[] = data.validators ?? [];

    return validators.map((v: any) => ({
      operatorAddress: v.operator_address ?? '',
      moniker: v.description?.moniker ?? 'Unknown',
      status: v.status ?? '',
      commission: (parseFloat(v.commission?.commission_rates?.rate ?? '0') * 100).toFixed(2) + '%',
      votingPower: (parseInt(v.tokens ?? '0', 10) / Math.pow(10, CONFIG.DENOM_DECIMALS)).toFixed(0),
      jailed: v.jailed ?? false,
    }));
  } catch {
    return [];
  }
}

async function getSignerClient(
  mnemonic: string,
): Promise<{ client: SigningStargateClient; address: string }> {
  const wallet = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic.trim(), {
    prefix: CONFIG.BECH32_PREFIX,
  });
  const [account] = await wallet.getAccounts();
  const client = await SigningStargateClient.connectWithSigner(CONFIG.RPC_ENDPOINT, wallet, {
    gasPrice: GasPrice.fromString(CONFIG.GAS_PRICE),
  });
  return { client, address: account.address };
}

export async function delegateVITA(
  mnemonic: string,
  validatorAddress: string,
  amount: string,
): Promise<string> {
  const { client, address } = await getSignerClient(mnemonic);
  const microAmount = Math.floor(parseFloat(amount) * Math.pow(10, CONFIG.DENOM_DECIMALS));
  const result = await client.delegateTokens(
    address,
    validatorAddress,
    coin(microAmount, CONFIG.DENOM),
    'auto',
  );
  if (result.code !== 0) throw new Error(`Delegation failed: ${result.rawLog}`);
  return result.transactionHash;
}

export async function undelegateVITA(
  mnemonic: string,
  validatorAddress: string,
  amount: string,
): Promise<string> {
  const { client, address } = await getSignerClient(mnemonic);
  const microAmount = Math.floor(parseFloat(amount) * Math.pow(10, CONFIG.DENOM_DECIMALS));
  const result = await client.undelegateTokens(
    address,
    validatorAddress,
    coin(microAmount, CONFIG.DENOM),
    'auto',
  );
  if (result.code !== 0) throw new Error(`Undelegation failed: ${result.rawLog}`);
  return result.transactionHash;
}

export async function claimRewards(mnemonic: string, validatorAddress: string): Promise<string> {
  const { client, address } = await getSignerClient(mnemonic);
  const result = await client.withdrawRewards(address, validatorAddress, 'auto');
  if (result.code !== 0) throw new Error(`Claim rewards failed: ${result.rawLog}`);
  return result.transactionHash;
}
