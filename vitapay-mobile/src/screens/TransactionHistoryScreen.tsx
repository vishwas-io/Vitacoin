import React, { useEffect, useState } from 'react';
import {
  View, Text, FlatList, StyleSheet, ActivityIndicator, TouchableOpacity,
} from 'react-native';
import { Transaction } from '../types/wallet';

const COLORS = { bg: '#0a0a0a', card: '#141414', accent: '#00ff88', text: '#ffffff', muted: '#888888', error: '#ff4444', warning: '#ffaa00' };

const STATUS_COLOR: Record<Transaction['status'], string> = {
  confirmed: COLORS.accent,
  pending: COLORS.warning,
  failed: COLORS.error,
};

const TYPE_ICON: Record<Transaction['type'], string> = {
  send: '↑', receive: '↓', delegate: '🔒', undelegate: '🔓', claim_rewards: '🎁', payment: '💳',
};

// Mock — real data in Job 2
async function fetchTransactions(): Promise<Transaction[]> {
  return [
    { hash: 'ABC123', from: 'vita1me...', to: 'vita1other...', amount: '10.000000', denom: 'VITA', type: 'send', status: 'confirmed', timestamp: Date.now() - 3600000, blockHeight: 15043, fee: '0.025000' },
    { hash: 'DEF456', from: 'vita1shop...', to: 'vita1me...', amount: '250.000000', denom: 'VITA', type: 'receive', status: 'confirmed', timestamp: Date.now() - 86400000, blockHeight: 14998, fee: '0.025000' },
    { hash: 'GHI789', from: 'vita1me...', to: 'vitavaloper1abc...', amount: '100.000000', denom: 'VITA', type: 'delegate', status: 'confirmed', timestamp: Date.now() - 172800000, blockHeight: 14500, fee: '0.025000' },
    { hash: 'JKL012', from: 'vita1me...', to: 'vita1cafe...', amount: '12.500000', denom: 'VITA', type: 'payment', status: 'pending', timestamp: Date.now() - 600000, blockHeight: 15050, fee: '0.025000' },
  ];
}

function timeAgo(ts: number): string {
  const diff = Date.now() - ts;
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
  return `${Math.floor(diff / 86400000)}d ago`;
}

export default function TransactionHistoryScreen() {
  const [txs, setTxs] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => { fetchTransactions().then(t => { setTxs(t); setLoading(false); }); }, []);

  if (loading) return <View style={styles.center}><ActivityIndicator color={COLORS.accent} size="large" /></View>;

  return (
    <View style={styles.container}>
      <Text style={styles.title}>History</Text>
      <FlatList
        data={txs}
        keyExtractor={t => t.hash}
        renderItem={({ item }) => (
          <View style={styles.txCard}>
            <View style={styles.txLeft}>
              <Text style={styles.icon}>{TYPE_ICON[item.type]}</Text>
            </View>
            <View style={styles.txMiddle}>
              <Text style={styles.txType}>{item.type.replace('_', ' ').toUpperCase()}</Text>
              <Text style={styles.txHash}>{item.hash.slice(0, 12)}...</Text>
              <Text style={styles.txTime}>{timeAgo(item.timestamp)} · Block #{item.blockHeight}</Text>
            </View>
            <View style={styles.txRight}>
              <Text style={styles.txAmount}>{item.amount} {item.denom}</Text>
              <View style={[styles.badge, { backgroundColor: STATUS_COLOR[item.status] + '33' }]}>
                <Text style={[styles.badgeText, { color: STATUS_COLOR[item.status] }]}>{item.status}</Text>
              </View>
            </View>
          </View>
        )}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg, padding: 16 },
  center: { flex: 1, backgroundColor: COLORS.bg, justifyContent: 'center', alignItems: 'center' },
  title: { fontSize: 24, fontWeight: 'bold', color: COLORS.text, marginBottom: 16 },
  txCard: { backgroundColor: COLORS.card, borderRadius: 12, padding: 14, marginBottom: 8, flexDirection: 'row', alignItems: 'center', gap: 12 },
  txLeft: { width: 40, alignItems: 'center' },
  icon: { fontSize: 22 },
  txMiddle: { flex: 1 },
  txRight: { alignItems: 'flex-end', gap: 6 },
  txType: { fontSize: 13, fontWeight: '700', color: COLORS.text },
  txHash: { fontSize: 11, color: COLORS.muted },
  txTime: { fontSize: 11, color: COLORS.muted, marginTop: 2 },
  txAmount: { fontSize: 14, fontWeight: '600', color: COLORS.text },
  badge: { borderRadius: 6, paddingHorizontal: 8, paddingVertical: 2 },
  badgeText: { fontSize: 11, fontWeight: '700' },
});
