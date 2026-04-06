import React, { useEffect, useState } from 'react';
import {
  View, Text, FlatList, TouchableOpacity, TextInput, StyleSheet, Alert, ActivityIndicator,
} from 'react-native';

const COLORS = { bg: '#0a0a0a', card: '#141414', accent: '#00ff88', text: '#ffffff', muted: '#888888' };

interface Validator {
  address: string;
  moniker: string;
  commission: string;
  votingPower: string;
  status: 'active' | 'inactive';
}

// Mock — real data in Job 2
async function fetchValidators(): Promise<Validator[]> {
  return [
    { address: 'vitavaloper1abc...', moniker: 'VitaNode Alpha', commission: '5%', votingPower: '1,250,000', status: 'active' },
    { address: 'vitavaloper1def...', moniker: 'CosmosHub Validator', commission: '8%', votingPower: '980,000', status: 'active' },
    { address: 'vitavaloper1ghi...', moniker: 'StakeWithUs', commission: '3%', votingPower: '750,000', status: 'active' },
  ];
}

async function delegateTokens(validatorAddress: string, amount: string): Promise<string> {
  await new Promise(r => setTimeout(r, 1500));
  return 'DELEGATE_TX_MOCK_HASH_' + Date.now();
}

export default function StakeScreen() {
  const [validators, setValidators] = useState<Validator[]>([]);
  const [selected, setSelected] = useState<Validator | null>(null);
  const [amount, setAmount] = useState('');
  const [loading, setLoading] = useState(true);
  const [staking, setStaking] = useState(false);
  const [stVita, setStVita] = useState('500.000000');

  useEffect(() => {
    fetchValidators().then(v => { setValidators(v); setLoading(false); });
  }, []);

  const onDelegate = async () => {
    if (!selected || !amount) { Alert.alert('Error', 'Select a validator and enter amount'); return; }
    setStaking(true);
    try {
      const hash = await delegateTokens(selected.address, amount);
      Alert.alert('Delegated!', `Staked ${amount} VITA\nTx: ${hash.slice(0, 20)}...`);
      setAmount('');
      setSelected(null);
    } catch (e: any) {
      Alert.alert('Error', e.message);
    } finally { setStaking(false); }
  };

  if (loading) return <View style={styles.center}><ActivityIndicator color={COLORS.accent} size="large" /></View>;

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Stake VITA</Text>
      <View style={styles.statCard}>
        <Text style={styles.label}>stVITA Balance</Text>
        <Text style={styles.bigValue}>{stVita} stVITA</Text>
      </View>
      <Text style={styles.sectionTitle}>Select Validator</Text>
      <FlatList
        data={validators}
        keyExtractor={v => v.address}
        style={styles.list}
        renderItem={({ item }) => (
          <TouchableOpacity
            style={[styles.validatorCard, selected?.address === item.address && styles.selected]}
            onPress={() => setSelected(item)}
          >
            <Text style={styles.moniker}>{item.moniker}</Text>
            <Text style={styles.details}>Commission: {item.commission} · Power: {item.votingPower}</Text>
          </TouchableOpacity>
        )}
      />
      {selected && (
        <View style={styles.delegateBox}>
          <TextInput
            style={styles.input} value={amount} onChangeText={setAmount}
            placeholder="Amount to stake (VITA)" placeholderTextColor={COLORS.muted} keyboardType="decimal-pad"
          />
          <TouchableOpacity style={[styles.btn, staking && styles.btnDisabled]} onPress={onDelegate} disabled={staking}>
            {staking ? <ActivityIndicator color={COLORS.bg} /> : <Text style={styles.btnText}>Delegate</Text>}
          </TouchableOpacity>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg, padding: 16 },
  center: { flex: 1, backgroundColor: COLORS.bg, justifyContent: 'center', alignItems: 'center' },
  title: { fontSize: 24, fontWeight: 'bold', color: COLORS.text, marginBottom: 16 },
  statCard: { backgroundColor: COLORS.card, borderRadius: 12, padding: 16, marginBottom: 16 },
  label: { fontSize: 12, color: COLORS.muted, textTransform: 'uppercase', letterSpacing: 1 },
  bigValue: { fontSize: 22, fontWeight: 'bold', color: COLORS.accent, marginTop: 4 },
  sectionTitle: { fontSize: 14, color: COLORS.muted, marginBottom: 8, textTransform: 'uppercase', letterSpacing: 1 },
  list: { maxHeight: 220, marginBottom: 12 },
  validatorCard: { backgroundColor: COLORS.card, borderRadius: 10, padding: 14, marginBottom: 8, borderWidth: 1, borderColor: 'transparent' },
  selected: { borderColor: COLORS.accent },
  moniker: { fontSize: 16, color: COLORS.text, fontWeight: '600' },
  details: { fontSize: 12, color: COLORS.muted, marginTop: 4 },
  delegateBox: { gap: 12 },
  input: { backgroundColor: COLORS.card, color: COLORS.text, borderRadius: 10, padding: 14, fontSize: 15, borderWidth: 1, borderColor: '#222' },
  btn: { backgroundColor: COLORS.accent, borderRadius: 12, padding: 16, alignItems: 'center' },
  btnDisabled: { opacity: 0.6 },
  btnText: { fontSize: 16, fontWeight: '700', color: COLORS.bg },
});
