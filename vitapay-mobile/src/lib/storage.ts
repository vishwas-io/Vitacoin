import * as SecureStore from 'expo-secure-store';

const KEYS = {
  MNEMONIC: 'vitapay_mnemonic',
  ADDRESS: 'vitapay_address',
} as const;

export async function saveMnemonic(mnemonic: string): Promise<void> {
  await SecureStore.setItemAsync(KEYS.MNEMONIC, mnemonic);
}

export async function getMnemonic(): Promise<string | null> {
  return SecureStore.getItemAsync(KEYS.MNEMONIC);
}

export async function saveAddress(address: string): Promise<void> {
  await SecureStore.setItemAsync(KEYS.ADDRESS, address);
}

export async function getAddress(): Promise<string | null> {
  return SecureStore.getItemAsync(KEYS.ADDRESS);
}

export async function clearWallet(): Promise<void> {
  await SecureStore.deleteItemAsync(KEYS.MNEMONIC);
  await SecureStore.deleteItemAsync(KEYS.ADDRESS);
}
