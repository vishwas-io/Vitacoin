import React, { useEffect, useState, useCallback } from 'react';
import {
  View, Text, FlatList, TouchableOpacity, TextInput,
  StyleSheet, Alert, ActivityIndicator, ScrollView,
} from 'react-native';
import { getValidators, getDelegations, delegateVITA, claimRewards, Validator, Delegation } from '../lib/staking';
import { getAddress, getMnemonic } from '../lib/storage';

const COLORS = {
  bg: '#0a0a0a', card: '#141414', accent: '#00ff88',
  text: '#ffffff', muted: '#888888', error: '#ff4444', warning: '#ffaa00',
};

// Estimated APR per validator (10% network default)
const ESTIMATED_APR = '10.00%';

export default function StakeScreen() {
  const [validators, setValidators] = useState<Validator[]>([]);
  const [delegations, setDelegations] = useState<Delegation[]>([]);
  const [selected, setSelected] = useState<Validator | null>(null);
  const [amount, setAmount] = useState('');
  const [loading, setLoading] = useState(true);
  const [staking, setStaking] = useState(false);
  const [claiming, setClaiming] = useState<string | null>(null);
  const [address, setAddress] = useState('');

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const addr = await getAddress();
      if (addr) {
        setAddress(addr);
        const [vals, dels] = await Promise.all([
          getValidators(),
          getDelegations(addr),
        ]);
        setValidators(vals);
        setDelegations(dels);
      }
    } catch {
      // keep empty
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const totalStaked = delegations.reduce(
    (sum, d) => sum + parseFloat(d.delegatedAmount || '0'), 0,
  ).toFixed(6);

  const totalRewards = delegations.reduce(
    (sum, d) => sum + parseFloat(d.pendingRewards || '0'), 0,
  ).toFixed(6);

  const onDelegate = async () => {
    if (!selected || !amount) {
      Alert.alert('Missing Fields', 'Select a validator and enter an amount to stake.');
      return;
    }
    const parsedAmount = parseFloat(amount);
    if (isNaN(parsedAmount) || parsedAmount <= 0) {
      Alert.alert('Invalid Amount', 'Please enter a valid amount greater than 0.');
      return;
    }
    setStaking(true);
    try {
      const mnemonic = await getMnemonic();
      if (!mnemonic) {
        Alert.alert('No Wallet', 'Please create or import a wallet first.');
        return;
      }
      const hash = await delegateVITA(mnemonic, selected.operatorAddress, amount);
      Alert.alert(
        'Delegated! 🔒',
        `Staked ${amount} VITA with ${selected.moniker}\n\nTx: ${hash.slice(0, 20)}...`,
      );
      setAmount('');
      setSelected(null);
      await loadData();
    } catch (e: any) {
      Alert.alert('Delegation Failed', e.message ?? 'Unknown error');
    } finally {
      setStaking(false);
    }
  };

  const onClaim = async (delegation: Delegation) => {
    setClaiming(delegation.validatorAddress);
    try {
      const mnemonic = await getMnemonic();
      if (!mnemonic) {
        Alert.alert('No Wallet', 'Please create or import a wallet first.');
        return;
      }
      const hash = await claimRewards(mnemonic, delegation.validatorAddress);
      Alert.alert(
        'Rewards Claimed! 🎁',
        `Claimed ${delegation.pendingRewards} VITA\n\nTx: ${hash.slice(0, 20)}...`,
      );
      await loadData();
    } catch (e: any) {
      Alert.alert('Claim Failed', e.message ?? 'Unknown error');
    } finally {
      setClaiming(null);
    }
  };

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={COLORS.accent} size="large" />
        <Text style={styles.muted}>Loading staking data...</Text>
      </View>
    );
  }

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <Text style={styles.title}>Stake VITA</Text>

      {/* Stats row */}
      <View style={styles.statsRow}>
        <View style={styles.statCard}>
          <Text style={styles.statLabel}>Staked</Text>
          <Text style={styles.statValue}>{totalStaked}</Text>
          <Text style={styles.statUnit}>VITA</Text>
        </View>
        <View style={styles.statCard}>
          <Text style={styles.statLabel}>Rewards</Text>
          <Text style={[styles.statValue, { color: COLORS.accent }]}>{totalRewards}</Text>
          <Text style={styles.statUnit}>VITA</Text>
        </View>
        <View style={styles.statCard}>
          <Text style={styles.statLabel}>APR</Text>
          <Text style={[styles.statValue, { color: COLORS.accent }]}>{ESTIMATED_APR}</Text>
          <Text style={styles.statUnit}>est.</Text>
        </View>
      </View>

      {/* Active delegations */}
      {delegations.length > 0 && (
        <>
          <Text style={styles.sectionTitle}>Your Delegations</Text>
          {delegations.map((d) => (
            <View key={d.validatorAddress} style={styles.delegationCard}>
              <View style={styles.delLeft}>
                <Text style={styles.delValidator} numberOfLines={1}>
                  {d.validatorName !== d.validatorAddress ? d.validatorName : d.validatorAddress.slice(0, 20) + '...'}
                </Text>
                <Text style={styles.delAmount}>{d.delegatedAmount} VITA staked</Text>
                <Text style={styles.delRewards}>🎁 {d.pendingRewards} VITA pending</Text>
              </View>
              <TouchableOpacity
                style={[styles.claimBtn, claiming === d.validatorAddress && styles.btnDisabled]}
                onPress={() => onClaim(d)}
                disabled={!!claiming}
              >
                {claiming === d.validatorAddress
                  ? <ActivityIndicator color={COLORS.bg} size="small" />
                  : <Text style={styles.claimText}>Claim</Text>}
              </TouchableOpacity>
            </View>
          ))}
        </>
      )}

      {/* Validator list */}
      <Text style={styles.sectionTitle}>
        {validators.length > 0 ? `Validators (${validators.length})` : 'Validators'}
      </Text>
      {validators.length === 0 ? (
        <Text style={styles.muted}>No active validators found. Network may be offline.</Text>
      ) : (
        validators.map((v) => (
          <TouchableOpacity
            key={v.operatorAddress}
            style={[styles.validatorCard, selected?.operatorAddress === v.operatorAddress && styles.selectedCard]}
            onPress={() => setSelected(v)}
          >
            <View style={styles.valLeft}>
              <Text style={styles.moniker}>{v.moniker}</Text>
              <Text style={styles.valDetails}>
                Commission: {v.commission} · Power: {parseInt(v.votingPower || '0').toLocaleString()} VITA
              </Text>
              <Text style={styles.apr}>APR: ~{ESTIMATED_APR}</Text>
            </View>
            {v.jailed && <Text style={styles.jailedBadge}>JAILED</Text>}
            {selected?.operatorAddress === v.operatorAddress && (
              <Text style={{ color: COLORS.accent, fontSize: 18 }}>✓</Text>
            )}
          </TouchableOpacity>
        ))
      )}

      {/* Delegation input */}
      {selected && (
        <View style={styles.delegateBox}>
          <Text style={styles.selectedLabel}>Staking with: {selected.moniker}</Text>
          <TextInput
            style={styles.input}
            value={amount}
            onChangeText={setAmount}
            placeholder="Amount to stake (VITA)"
            placeholderTextColor={COLORS.muted}
            keyboardType="decimal-pad"
          />
          <TouchableOpacity
            style={[styles.btn, staking && styles.btnDisabled]}
            onPress={onDelegate}
            disabled={staking}
          >
            {staking
              ? <ActivityIndicator color={COLORS.bg} />
              : <Text style={styles.btnText}>Delegate</Text>}
          </TouchableOpacity>
        </View>
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg },
  content: { padding: 16, paddingBottom: 40 },
  center: { flex: 1, backgroundColor: COLORS.bg, justifyContent: 'center', alignItems: 'center', gap: 12 },
  title: { fontSize: 24, fontWeight: 'bold', color: COLORS.text, marginBottom: 16 },
  statsRow: { flexDirection: 'row', gap: 10, marginBottom: 20 },
  statCard: { flex: 1, backgroundColor: COLORS.card, borderRadius: 12, padding: 14, alignItems: 'center' },
  statLabel: { fontSize: 11, color: COLORS.muted, textTransform: 'uppercase', letterSpacing: 1 },
  statValue: { fontSize: 18, fontWeight: 'bold', color: COLORS.text, marginTop: 4 },
  statUnit: { fontSize: 11, color: COLORS.muted, marginTop: 2 },
  sectionTitle: { fontSize: 13, color: COLORS.muted, textTransform: 'uppercase', letterSpacing: 1, marginBottom: 10, marginTop: 4 },
  delegationCard: {
    backgroundColor: COLORS.card, borderRadius: 12, padding: 14,
    marginBottom: 8, flexDirection: 'row', alignItems: 'center',
  },
  delLeft: { flex: 1 },
  delValidator: { fontSize: 14, fontWeight: '600', color: COLORS.text },
  delAmount: { fontSize: 12, color: COLORS.muted, marginTop: 2 },
  delRewards: { fontSize: 12, color: COLORS.accent, marginTop: 2 },
  claimBtn: {
    backgroundColor: COLORS.accent, borderRadius: 8,
    paddingHorizontal: 14, paddingVertical: 8,
  },
  claimText: { fontSize: 13, fontWeight: '700', color: COLORS.bg },
  validatorCard: {
    backgroundColor: COLORS.card, borderRadius: 10,
    padding: 14, marginBottom: 8,
    flexDirection: 'row', alignItems: 'center',
    borderWidth: 1, borderColor: 'transparent',
  },
  selectedCard: { borderColor: COLORS.accent },
  valLeft: { flex: 1 },
  moniker: { fontSize: 16, color: COLORS.text, fontWeight: '600' },
  valDetails: { fontSize: 12, color: COLORS.muted, marginTop: 4 },
  apr: { fontSize: 12, color: COLORS.accent, marginTop: 2 },
  jailedBadge: {
    fontSize: 11, color: COLORS.error,
    backgroundColor: COLORS.error + '22',
    paddingHorizontal: 8, paddingVertical: 3, borderRadius: 6, fontWeight: '700',
  },
  delegateBox: { gap: 12, marginTop: 16 },
  selectedLabel: { fontSize: 13, color: COLORS.muted },
  input: {
    backgroundColor: COLORS.card, color: COLORS.text,
    borderRadius: 10, padding: 14, fontSize: 15,
    borderWidth: 1, borderColor: '#222',
  },
  btn: { backgroundColor: COLORS.accent, borderRadius: 12, padding: 16, alignItems: 'center' },
  btnDisabled: { opacity: 0.6 },
  btnText: { fontSize: 16, fontWeight: '700', color: COLORS.bg },
  muted: { color: COLORS.muted, fontSize: 13, textAlign: 'center' },
});
