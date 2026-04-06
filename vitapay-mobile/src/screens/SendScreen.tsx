import React, { useState } from 'react';
import {
  View, Text, TextInput, TouchableOpacity, StyleSheet, Alert, ActivityIndicator,
} from 'react-native';
import { sendVITA, estimateFee } from '../lib/wallet';
import { getMnemonic } from '../lib/storage';

const COLORS = { bg: '#0a0a0a', card: '#141414', accent: '#00ff88', text: '#ffffff', muted: '#888888', error: '#ff4444' };

export default function SendScreen({ navigation }: any) {
  const [to, setTo] = useState('');
  const [amount, setAmount] = useState('');
  const [memo, setMemo] = useState('');
  const [fee, setFee] = useState<string | null>(null);
  const [sending, setSending] = useState(false);

  const onAmountBlur = async () => {
    if (amount) setFee(await estimateFee(amount));
  };

  const onSend = async () => {
    if (!to || !amount) {
      Alert.alert('Error', 'Address and amount required');
      return;
    }
    setSending(true);
    try {
      const mnemonic = await getMnemonic();
      if (!mnemonic) {
        Alert.alert('Error', 'No wallet found. Please set up your wallet first.');
        return;
      }
      const txHash = await sendVITA(mnemonic, to, amount, memo);
      Alert.alert(
        'Transaction Sent ✓',
        `Tx hash:\n${txHash}`,
        [{ text: 'OK', onPress: () => navigation.goBack() }],
      );
    } catch (e: any) {
      Alert.alert('Transaction Failed', e.message ?? 'Unknown error');
    } finally {
      setSending(false);
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Send VITA</Text>
      <Text style={styles.label}>Recipient Address</Text>
      <TextInput
        style={styles.input} value={to} onChangeText={setTo}
        placeholder="vita1..." placeholderTextColor={COLORS.muted} autoCapitalize="none"
      />
      <Text style={styles.label}>Amount (VITA)</Text>
      <TextInput
        style={styles.input} value={amount} onChangeText={setAmount}
        onBlur={onAmountBlur} placeholder="0.000000" placeholderTextColor={COLORS.muted}
        keyboardType="decimal-pad"
      />
      <Text style={styles.label}>Memo (optional)</Text>
      <TextInput
        style={styles.input} value={memo} onChangeText={setMemo}
        placeholder="Payment memo" placeholderTextColor={COLORS.muted}
      />
      {fee && <Text style={styles.fee}>Estimated fee: {fee} VITA</Text>}
      <TouchableOpacity style={[styles.btn, sending && styles.btnDisabled]} onPress={onSend} disabled={sending}>
        {sending ? <ActivityIndicator color={COLORS.bg} /> : <Text style={styles.btnText}>Send</Text>}
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg, padding: 16 },
  title: { fontSize: 24, fontWeight: 'bold', color: COLORS.text, marginBottom: 24 },
  label: { fontSize: 12, color: COLORS.muted, textTransform: 'uppercase', letterSpacing: 1, marginBottom: 6, marginTop: 12 },
  input: { backgroundColor: COLORS.card, color: COLORS.text, borderRadius: 10, padding: 14, fontSize: 15, borderWidth: 1, borderColor: '#222' },
  fee: { color: COLORS.muted, fontSize: 13, marginTop: 8 },
  btn: { backgroundColor: COLORS.accent, borderRadius: 12, padding: 18, alignItems: 'center', marginTop: 24 },
  btnDisabled: { opacity: 0.6 },
  btnText: { fontSize: 17, fontWeight: '700', color: COLORS.bg },
});
