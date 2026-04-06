import React, { useEffect, useState, useCallback } from 'react';
import {
  View, Text, FlatList, StyleSheet, ActivityIndicator, TouchableOpacity, RefreshControl,
} from 'react-native';
import { Transaction } from '../types/wallet';
import { getTransactionHistory } from '../lib/wallet';
import { getAddress } from '../lib/storage';

const COLORS = {
  bg: '#0a0a0a', card: '#141414', accent: '#00ff88',
  text: '#ffffff', muted: '#888888', error: '#ff4444', warning: '#ffaa00',
};

const STATUS_COLOR: Record<string, string> = {
  confirmed: COLORS.accent,
  pending: COLORS.warning,
  failed: COLORS.error,
};

const TYPE_ICON: Record<string, string> = {
  send: '↑',
  receive: '↓',
  delegate: '🔒',
  undelegate: '🔓',
  claim_rewards: '🎁',
  payment: '💳',
  MsgSend: '↑',
  MsgDelegate: '🔒',
  MsgUndelegate: '🔓',
  MsgWithdrawDelegatorReward: '🎁',
  MsgIBCTransfer: '🌉',
  Unknown: '•',
};

function friendlyType(type: string): string {
  const map: Record<string, string> = {
    MsgSend: 'SEND',
    MsgDelegate: 'DELEGATE',
    MsgUndelegate: 'UNDELEGATE',
    MsgWithdrawDelegatorReward: 'CLAIM REWARDS',
    MsgIBCTransfer: 'IBC TRANSFER',
    send: 'SEND',
    receive: 'RECEIVE',
    delegate: 'DELEGATE',
    undelegate: 'UNDELEGATE',
    claim_rewards: 'CLAIM REWARDS',
    payment: 'PAYMENT',
  };
  return map[type] ?? type.replace(/_/g, ' ').toUpperCase();
}

function timeAgo(ts: number | string): string {
  const date = typeof ts === 'string' ? new Date(ts).getTime() : ts;
  const diff = Date.now() - date;
  if (diff < 60000) return 'just now';
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
  return `${Math.floor(diff / 86400000)}d ago`;
}

interface DisplayTx {
  hash: string;
  type: string;
  amount: string;
  denom: string;
  status: string;
  timestamp: number | string;
  blockHeight: number;
  fee: string;
}

function toDisplayTx(tx: any): DisplayTx {
  return {
    hash: tx.hash ?? tx.txHash ?? '',
    type: tx.type ?? 'Unknown',
    amount: tx.amount ?? '0',
    denom: tx.denom ?? 'VITA',
    status: tx.success === false ? 'failed' : (tx.status ?? 'confirmed'),
    timestamp: tx.timestamp ?? Date.now(),
    blockHeight: tx.blockHeight ?? tx.height ?? 0,
    fee: tx.fee ?? '0',
  };
}

export default function TransactionHistoryScreen() {
  const [txs, setTxs] = useState<DisplayTx[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [address, setAddress] = useState('');

  const load = useCallback(async (isRefresh = false) => {
    if (isRefresh) setRefreshing(true); else setLoading(true);
    try {
      const addr = await getAddress();
      if (!addr) {
        setTxs([]);
        return;
      }
      setAddress(addr);
      const history = await getTransactionHistory(addr);
      setTxs(history.map(toDisplayTx));
    } catch {
      setTxs([]);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={COLORS.accent} size="large" />
        <Text style={styles.muted}>Loading transactions...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>History</Text>
      {txs.length === 0 ? (
        <View style={styles.empty}>
          <Text style={styles.emptyIcon}>📭</Text>
          <Text style={styles.emptyText}>No transactions yet</Text>
          <Text style={styles.muted}>Your transaction history will appear here.</Text>
        </View>
      ) : (
        <FlatList
          data={txs}
          keyExtractor={(t) => t.hash || String(t.timestamp)}
          refreshControl={
            <RefreshControl refreshing={refreshing} onRefresh={() => load(true)} tintColor={COLORS.accent} />
          }
          renderItem={({ item }) => {
            const icon = TYPE_ICON[item.type] ?? '•';
            const statusColor = STATUS_COLOR[item.status] ?? COLORS.muted;
            return (
              <View style={styles.txCard}>
                <View style={styles.txLeft}>
                  <Text style={styles.icon}>{icon}</Text>
                </View>
                <View style={styles.txMiddle}>
                  <Text style={styles.txType}>{friendlyType(item.type)}</Text>
                  <Text style={styles.txHash}>
                    {item.hash ? item.hash.slice(0, 14) + '...' : '—'}
                  </Text>
                  <Text style={styles.txTime}>
                    {timeAgo(item.timestamp)}
                    {item.blockHeight ? ` · Block #${item.blockHeight}` : ''}
                  </Text>
                </View>
                <View style={styles.txRight}>
                  <Text style={styles.txAmount}>{item.amount} {item.denom}</Text>
                  <View style={[styles.badge, { backgroundColor: statusColor + '33' }]}>
                    <Text style={[styles.badgeText, { color: statusColor }]}>{item.status}</Text>
                  </View>
                </View>
              </View>
            );
          }}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg, padding: 16 },
  center: { flex: 1, backgroundColor: COLORS.bg, justifyContent: 'center', alignItems: 'center', gap: 12 },
  title: { fontSize: 24, fontWeight: 'bold', color: COLORS.text, marginBottom: 16 },
  empty: { flex: 1, justifyContent: 'center', alignItems: 'center', gap: 8 },
  emptyIcon: { fontSize: 48 },
  emptyText: { fontSize: 18, color: COLORS.text, fontWeight: '600' },
  muted: { fontSize: 13, color: COLORS.muted, textAlign: 'center' },
  txCard: {
    backgroundColor: COLORS.card, borderRadius: 12,
    padding: 14, marginBottom: 8,
    flexDirection: 'row', alignItems: 'center', gap: 12,
  },
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
