import React, { useEffect, useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ScrollView,
  ActivityIndicator,
} from 'react-native';
import { WalletAccount } from '../types/wallet';

// Mock API — replaced in Job 2
async function fetchWalletAccount(): Promise<WalletAccount> {
  return {
    address: 'vita1qg5eathl0pgdpe4ghrjkz6y9pljpz47v4ym6qy',
    publicKey: 'Ao3HgNq...',
    balance: '1250.500000',
    stakedBalance: '500.000000',
    pendingRewards: '12.345678',
  };
}

const COLORS = { bg: '#0a0a0a', card: '#141414', accent: '#00ff88', text: '#ffffff', muted: '#888888' };

export default function HomeScreen({ navigation }: any) {
  const [account, setAccount] = useState<WalletAccount | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchWalletAccount().then((a) => { setAccount(a); setLoading(false); });
  }, []);

  if (loading) return <View style={styles.center}><ActivityIndicator color={COLORS.accent} size="large" /></View>;

  return (
    <ScrollView style={styles.container}>
      <View style={styles.card}>
        <Text style={styles.label}>VITA Balance</Text>
        <Text style={styles.balance}>{account?.balance} VITA</Text>
        <Text style={styles.address} numberOfLines={1}>{account?.address}</Text>
      </View>
      <View style={styles.row}>
        <View style={styles.statCard}>
          <Text style={styles.label}>Staked</Text>
          <Text style={styles.statValue}>{account?.stakedBalance} VITA</Text>
        </View>
        <View style={styles.statCard}>
          <Text style={styles.label}>Rewards</Text>
          <Text style={[styles.statValue, { color: COLORS.accent }]}>{account?.pendingRewards} VITA</Text>
        </View>
      </View>
      <View style={styles.actions}>
        {[
          { label: 'Send', screen: 'Send' },
          { label: 'Receive', screen: 'Receive' },
          { label: 'Pay', screen: 'Pay' },
          { label: 'Stake', screen: 'Stake' },
        ].map(({ label, screen }) => (
          <TouchableOpacity key={label} style={styles.actionBtn} onPress={() => navigation.navigate(screen)}>
            <Text style={styles.actionText}>{label}</Text>
          </TouchableOpacity>
        ))}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg, padding: 16 },
  center: { flex: 1, backgroundColor: COLORS.bg, justifyContent: 'center', alignItems: 'center' },
  card: { backgroundColor: COLORS.card, borderRadius: 16, padding: 20, marginBottom: 16, alignItems: 'center' },
  balance: { fontSize: 32, fontWeight: 'bold', color: COLORS.accent, marginVertical: 8 },
  label: { fontSize: 12, color: COLORS.muted, textTransform: 'uppercase', letterSpacing: 1 },
  address: { fontSize: 11, color: COLORS.muted, marginTop: 4 },
  row: { flexDirection: 'row', gap: 12, marginBottom: 16 },
  statCard: { flex: 1, backgroundColor: COLORS.card, borderRadius: 12, padding: 16 },
  statValue: { fontSize: 18, fontWeight: '600', color: COLORS.text, marginTop: 4 },
  actions: { flexDirection: 'row', flexWrap: 'wrap', gap: 12 },
  actionBtn: { flex: 1, minWidth: '40%', backgroundColor: COLORS.accent, borderRadius: 12, padding: 16, alignItems: 'center' },
  actionText: { fontSize: 16, fontWeight: '700', color: COLORS.bg },
});
