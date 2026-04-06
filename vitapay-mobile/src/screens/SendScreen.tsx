import React, { useState } from 'react';
import {
  View, Text, TextInput, TouchableOpacity, StyleSheet, Alert, ActivityIndicator,
} from 'react-native';

const COLORS = { bg: '#0a0a0a', card: '#141414', accent: '#00ff88', text: '#ffffff', muted: '#888888', error: '#ff4444' };

// Mock — real signing in Job 2
async function estimateFee(amount: string): Promise<string> {
  return (parseFloat(amount || '0') * 0.0025 + 0.01).toFixed(6);
}
async function sendTransaction(to: string, amount: string, memo: string): Promise<string> {
  await new Promise(r => setTimeout(r, 1500));
  return '7A3F9B2C1D4E5F6A8B9C0D1E2F3A4B5C6D7E8F9A0B1C2D3E4F5A6B7C8D9E0F';
}

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
    if (!to || !amount) { Alert.alert('Error', 'Address and amount required'); return; }
    setSending(true);
    try {
      const hash = await sendTransaction(to, amount, memo);
      Alert.alert('Success', `Tx hash:\n${hash}`, [{ text: 'OK', onPress: () => navigation.goBack() }]);
    } catch (e: any) {
      Alert.alert('Error', e.message);
    } finally { setSending(false); }
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
