import React, { useEffect, useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ScrollView,
  ActivityIndicator,
  RefreshControl,
} from 'react-native';
import { getBalance } from '../lib/wallet';
import { getAddress } from '../lib/storage';
import { getDelegations } from '../lib/staking';

const COLORS = { bg: '#0a0a0a', card: '#141414', accent: '#00ff88', text: '#ffffff', muted: '#888888' };

export default function HomeScreen({ navigation }: any) {
  const [address, setAddress] = useState<string | null>(null);
  const [balance, setBalance] = useState<string>('0.000000');
  const [stakedBalance, setStakedBalance] = useState<string>('0.000000');
  const [pendingRewards, setPendingRewards] = useState<string>('0.000000');
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const loadData = useCallback(async () => {
    try {
      const addr = await getAddress();
      if (!addr) {
        setLoading(false);
        return;
      }
      setAddress(addr);
      const [bal, delegations] = await Promise.all([
        getBalance(addr),
        getDelegations(addr),
      ]);
      setBalance(bal);

      const totalStaked = delegations.reduce(
        (sum, d) => sum + parseFloat(d.delegatedAmount),
        0,
      ).toFixed(6);
      const totalRewards = delegations.reduce(
        (sum, d) => sum + parseFloat(d.pendingRewards),
        0,
      ).toFixed(6);
      setStakedBalance(totalStaked);
      setPendingRewards(totalRewards);
    } catch (e) {
      // silently fail; show zeros
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const onRefresh = () => {
    setRefreshing(true);
    loadData();
  };

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={COLORS.accent} size="large" />
      </View>
    );
  }

  return (
    <ScrollView
      style={styles.container}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor={COLORS.accent} />}
    >
      <View style={styles.card}>
        <Text style={styles.label}>VITA Balance</Text>
        <Text style={styles.balance}>{balance} VITA</Text>
        <Text style={styles.address} numberOfLines={1}>{address ?? 'No wallet loaded'}</Text>
      </View>
      <View style={styles.row}>
        <View style={styles.statCard}>
          <Text style={styles.label}>Staked</Text>
          <Text style={styles.statValue}>{stakedBalance} VITA</Text>
        </View>
        <View style={styles.statCard}>
          <Text style={styles.label}>Rewards</Text>
          <Text style={[styles.statValue, { color: COLORS.accent }]}>{pendingRewards} VITA</Text>
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
