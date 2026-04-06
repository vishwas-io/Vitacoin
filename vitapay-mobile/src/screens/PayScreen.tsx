import React, { useState, useEffect, useRef } from 'react';
import {
  View, Text, StyleSheet, TouchableOpacity, Alert, ActivityIndicator,
} from 'react-native';
import { Camera, CameraView, BarcodeScanningResult } from 'expo-camera';
import { PaymentRequest } from '../types/wallet';
import { parsePaymentQR, executePayment } from '../lib/payments';
import { getMnemonic } from '../lib/storage';

const COLORS = {
  bg: '#0a0a0a', card: '#141414', accent: '#00ff88',
  text: '#ffffff', muted: '#888888', error: '#ff4444',
};

export default function PayScreen() {
  const [hasPermission, setHasPermission] = useState<boolean | null>(null);
  const [scanning, setScanning] = useState(false);
  const [payment, setPayment] = useState<PaymentRequest | null>(null);
  const [paying, setPaying] = useState(false);
  const scannedRef = useRef(false);

  useEffect(() => {
    Camera.requestCameraPermissionsAsync().then(({ status }) => {
      setHasPermission(status === 'granted');
    });
  }, []);

  const startScan = () => {
    scannedRef.current = false;
    setScanning(true);
  };

  const onBarcodeScanned = (result: BarcodeScanningResult) => {
    if (scannedRef.current) return;
    scannedRef.current = true;
    setScanning(false);

    const req = parsePaymentQR(result.data);
    if (!req) {
      Alert.alert('Invalid QR', 'This QR code is not a valid VITAPAY payment request.');
      return;
    }
    if (req.expiresAt < Date.now()) {
      Alert.alert('Expired', 'This payment request has expired. Ask the merchant for a new QR code.');
      return;
    }
    setPayment(req);
  };

  const onPay = async () => {
    if (!payment) return;
    setPaying(true);
    try {
      const mnemonic = await getMnemonic();
      if (!mnemonic) {
        Alert.alert('No Wallet', 'Please create or import a wallet first.');
        return;
      }
      const hash = await executePayment(mnemonic, payment);
      Alert.alert(
        'Payment Sent ✅',
        `${payment.amount} ${payment.denom} sent to ${payment.merchantName}\n\nTx: ${hash.slice(0, 20)}...`,
      );
      setPayment(null);
    } catch (e: any) {
      Alert.alert('Payment Failed', e.message ?? 'Unknown error');
    } finally {
      setPaying(false);
    }
  };

  const timeUntilExpiry = payment
    ? Math.max(0, Math.floor((payment.expiresAt - Date.now()) / 1000))
    : 0;

  if (hasPermission === null) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={COLORS.accent} size="large" />
        <Text style={styles.muted}>Requesting camera access...</Text>
      </View>
    );
  }

  if (hasPermission === false) {
    return (
      <View style={styles.center}>
        <Text style={styles.errorText}>Camera permission denied.</Text>
        <Text style={styles.muted}>Enable camera access in Settings to scan QR codes.</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>VITAPAY</Text>

      {scanning ? (
        <View style={styles.scannerContainer}>
          <CameraView
            style={StyleSheet.absoluteFill}
            barcodeScannerSettings={{ barcodeTypes: ['qr'] }}
            onBarcodeScanned={onBarcodeScanned}
          />
          <View style={styles.scanOverlay}>
            <View style={styles.scanFrame} />
            <Text style={styles.scanHint}>Align QR code within the frame</Text>
          </View>
          <TouchableOpacity style={styles.cancelScanBtn} onPress={() => setScanning(false)}>
            <Text style={styles.cancelScanText}>Cancel</Text>
          </TouchableOpacity>
        </View>
      ) : !payment ? (
        <View style={styles.center}>
          <View style={styles.qrFrame}>
            <Text style={styles.qrIcon}>📷</Text>
            <Text style={styles.qrHint}>Point camera at merchant QR</Text>
          </View>
          <TouchableOpacity style={styles.btn} onPress={startScan}>
            <Text style={styles.btnText}>Scan QR Code</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <View style={styles.payCard}>
          <Text style={styles.merchantLabel}>PAY TO</Text>
          <Text style={styles.merchantName}>{payment.merchantName}</Text>
          <Text style={styles.payAmount}>{payment.amount} {payment.denom}</Text>
          {payment.memo ? <Text style={styles.memo}>📝 {payment.memo}</Text> : null}
          <Text style={styles.toAddr} numberOfLines={1}>→ {payment.merchantAddress}</Text>
          {timeUntilExpiry < 300 && (
            <Text style={styles.expiryWarning}>⏳ Expires in {timeUntilExpiry}s</Text>
          )}
          <View style={styles.btnRow}>
            <TouchableOpacity style={styles.cancelBtn} onPress={() => setPayment(null)}>
              <Text style={styles.cancelText}>Cancel</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.btn, paying && styles.btnDisabled, { flex: 1 }]}
              onPress={onPay}
              disabled={paying}
            >
              {paying
                ? <ActivityIndicator color={COLORS.bg} />
                : <Text style={styles.btnText}>Pay Now</Text>}
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
  scannerContainer: { flex: 1, position: 'relative', borderRadius: 16, overflow: 'hidden' },
  scanOverlay: {
    ...StyleSheet.absoluteFillObject,
    justifyContent: 'center',
    alignItems: 'center',
    gap: 16,
  },
  scanFrame: {
    width: 240, height: 240,
    borderWidth: 3, borderColor: COLORS.accent,
    borderRadius: 16,
  },
  scanHint: { color: COLORS.text, fontSize: 14, backgroundColor: 'rgba(0,0,0,0.6)', paddingHorizontal: 12, paddingVertical: 6, borderRadius: 8 },
  cancelScanBtn: { position: 'absolute', bottom: 32, alignSelf: 'center', backgroundColor: 'rgba(0,0,0,0.7)', paddingHorizontal: 24, paddingVertical: 12, borderRadius: 12 },
  cancelScanText: { color: COLORS.text, fontSize: 16, fontWeight: '600' },
  qrFrame: { width: 240, height: 240, borderWidth: 2, borderColor: COLORS.accent, borderRadius: 16, justifyContent: 'center', alignItems: 'center', gap: 12 },
  qrIcon: { fontSize: 48 },
  qrHint: { color: COLORS.muted, fontSize: 14, textAlign: 'center', paddingHorizontal: 20 },
  btn: { backgroundColor: COLORS.accent, borderRadius: 12, padding: 16, alignItems: 'center', minWidth: 160 },
  btnDisabled: { opacity: 0.6 },
  btnText: { fontSize: 16, fontWeight: '700', color: COLORS.bg },
  payCard: { backgroundColor: COLORS.card, borderRadius: 16, padding: 24, marginTop: 40, gap: 10 },
  merchantLabel: { fontSize: 11, color: COLORS.muted, textTransform: 'uppercase', letterSpacing: 1 },
  merchantName: { fontSize: 22, fontWeight: 'bold', color: COLORS.text },
  payAmount: { fontSize: 36, fontWeight: 'bold', color: COLORS.accent },
  memo: { fontSize: 14, color: COLORS.muted },
  toAddr: { fontSize: 12, color: COLORS.muted },
  expiryWarning: { fontSize: 13, color: '#ffaa00', fontWeight: '600' },
  btnRow: { flexDirection: 'row', gap: 12, marginTop: 12 },
  cancelBtn: { flex: 1, borderWidth: 1, borderColor: COLORS.error, borderRadius: 12, padding: 16, alignItems: 'center' },
  cancelText: { fontSize: 16, fontWeight: '700', color: COLORS.error },
  muted: { color: COLORS.muted, marginTop: 8, fontSize: 13, textAlign: 'center' },
  errorText: { color: COLORS.error, fontSize: 16, fontWeight: '600', textAlign: 'center' },
});
