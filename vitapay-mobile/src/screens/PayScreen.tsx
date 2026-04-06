import React, { useState } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, Alert, ActivityIndicator } from 'react-native';
import { PaymentRequest } from '../types/wallet';

const COLORS = { bg: '#0a0a0a', card: '#141414', accent: '#00ff88', text: '#ffffff', muted: '#888888', error: '#ff4444' };

// Mock QR scan result — real camera scan in Job 2
async function mockScanQR(): Promise<PaymentRequest> {
  await new Promise(r => setTimeout(r, 1000));
  return {
    merchantAddress: 'vita1merchant9x8y7z6w5v4u3t2s1r0q...',
    merchantName: 'VitaCafe ☕',
    amount: '12.500000',
    denom: 'VITA',
    memo: 'Order #1042',
    expiresAt: Date.now() + 300_000,
  };
}

async function processPayment(req: PaymentRequest): Promise<string> {
  await new Promise(r => setTimeout(r, 2000));
  return 'PAY_TX_MOCK_' + Date.now();
}

export default function PayScreen() {
  const [scanning, setScanning] = useState(false);
  const [payment, setPayment] = useState<PaymentRequest | null>(null);
  const [paying, setPaying] = useState(false);

  const onScan = async () => {
    setScanning(true);
    try {
      const req = await mockScanQR();
      setPayment(req);
    } catch (e: any) {
      Alert.alert('Scan failed', e.message);
    } finally { setScanning(false); }
  };

  const onPay = async () => {
    if (!payment) return;
    if (payment.expiresAt < Date.now()) { Alert.alert('Expired', 'This payment request has expired'); return; }
    setPaying(true);
    try {
      const hash = await processPayment(payment);
      Alert.alert('Payment Sent! ✅', `${payment.amount} ${payment.denom} to ${payment.merchantName}\nTx: ${hash.slice(0, 20)}...`);
      setPayment(null);
    } catch (e: any) {
      Alert.alert('Payment failed', e.message);
    } finally { setPaying(false); }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>VITAPAY</Text>
      {!payment ? (
        <View style={styles.center}>
          <View style={styles.qrFrame}>
            {scanning
              ? <ActivityIndicator color={COLORS.accent} size="large" />
              : <Text style={styles.qrHint}>Point camera at merchant QR</Text>}
          </View>
          <TouchableOpacity style={[styles.btn, scanning && styles.btnDisabled]} onPress={onScan} disabled={scanning}>
            <Text style={styles.btnText}>{scanning ? 'Scanning...' : 'Scan QR Code'}</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <View style={styles.payCard}>
          <Text style={styles.merchantName}>{payment.merchantName}</Text>
          <Text style={styles.payAmount}>{payment.amount} {payment.denom}</Text>
          <Text style={styles.memo}>Memo: {payment.memo}</Text>
          <Text style={styles.toAddr} numberOfLines={1}>To: {payment.merchantAddress}</Text>
          <View style={styles.btnRow}>
            <TouchableOpacity style={styles.cancelBtn} onPress={() => setPayment(null)}>
              <Text style={styles.cancelText}>Cancel</Text>
            </TouchableOpacity>
            <TouchableOpacity style={[styles.btn, paying && styles.btnDisabled, { flex: 1 }]} onPress={onPay} disabled={paying}>
              {paying ? <ActivityIndicator color={COLORS.bg} /> : <Text style={styles.btnText}>Pay Now</Text>}
            </TouchableOpacity>
          </View>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg, padding: 16 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', gap: 24 },
  title: { fontSize: 24, fontWeight: 'bold', color: COLORS.accent, marginBottom: 24 },
  qrFrame: { width: 240, height: 240, borderWidth: 2, borderColor: COLORS.accent, borderRadius: 16, justifyContent: 'center', alignItems: 'center' },
  qrHint: { color: COLORS.muted, fontSize: 14, textAlign: 'center', paddingHorizontal: 20 },
  btn: { backgroundColor: COLORS.accent, borderRadius: 12, padding: 16, alignItems: 'center', minWidth: 160 },
  btnDisabled: { opacity: 0.6 },
  btnText: { fontSize: 16, fontWeight: '700', color: COLORS.bg },
  payCard: { backgroundColor: COLORS.card, borderRadius: 16, padding: 24, marginTop: 40, gap: 12 },
  merchantName: { fontSize: 22, fontWeight: 'bold', color: COLORS.text },
  payAmount: { fontSize: 36, fontWeight: 'bold', color: COLORS.accent },
  memo: { fontSize: 14, color: COLORS.muted },
  toAddr: { fontSize: 12, color: COLORS.muted },
  btnRow: { flexDirection: 'row', gap: 12, marginTop: 12 },
  cancelBtn: { flex: 1, borderWidth: 1, borderColor: COLORS.error, borderRadius: 12, padding: 16, alignItems: 'center' },
  cancelText: { fontSize: 16, fontWeight: '700', color: COLORS.error },
});
