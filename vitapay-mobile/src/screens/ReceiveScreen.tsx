import React, { useEffect, useState } from 'react';
import {
  View, Text, TouchableOpacity, StyleSheet, Alert, Share, ScrollView,
} from 'react-native';
import QRCode from 'react-native-qrcode-svg';
import * as Clipboard from 'expo-clipboard';
import { getAddress } from '../lib/storage';

const COLORS = {
  bg: '#0a0a0a', card: '#141414', accent: '#00ff88',
  text: '#ffffff', muted: '#888888',
};

export default function ReceiveScreen() {
  const [address, setAddress] = useState('');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    getAddress().then((addr) => {
      if (addr) setAddress(addr);
    });
  }, []);

  const copyAddress = async () => {
    if (!address) return;
    try {
      await Clipboard.setStringAsync(address);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      Alert.alert('Address', address);
    }
  };

  const shareAddress = () => {
    if (!address) return;
    Share.share({
      message: `My VITA address: ${address}`,
      title: 'My VITA Address',
    });
  };

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text style={styles.title}>Receive VITA</Text>
      <Text style={styles.subtitle}>Share your address or QR code to receive payments.</Text>

      <View style={styles.qrCard}>
        {address ? (
          <QRCode value={address} size={200} backgroundColor="#ffffff" color="#000000" />
        ) : (
          <View style={styles.qrPlaceholder}>
            <Text style={{ color: COLORS.muted }}>Loading wallet...</Text>
          </View>
        )}
      </View>

      <Text style={styles.addressLabel}>Your VITA Address</Text>
      <View style={styles.addressBox}>
        <Text style={styles.address} selectable>{address || '—'}</Text>
      </View>

      <View style={styles.btnRow}>
        <TouchableOpacity style={styles.btn} onPress={copyAddress}>
          <Text style={styles.btnText}>{copied ? '✅ Copied!' : 'Copy'}</Text>
        </TouchableOpacity>
        <TouchableOpacity style={[styles.btn, styles.btnOutline]} onPress={shareAddress}>
          <Text style={[styles.btnText, { color: COLORS.accent }]}>Share</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.infoBox}>
        <Text style={styles.infoText}>
          ⚡ Only send VITA (uvita) to this address.{'\n'}
          Sending other tokens may result in permanent loss.
        </Text>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: COLORS.bg, padding: 16,
    alignItems: 'center', flexGrow: 1,
  },
  title: { fontSize: 24, fontWeight: 'bold', color: COLORS.text, marginBottom: 6, alignSelf: 'flex-start' },
  subtitle: { fontSize: 13, color: COLORS.muted, marginBottom: 24, alignSelf: 'flex-start' },
  qrCard: {
    backgroundColor: '#ffffff', borderRadius: 16, padding: 20,
    marginBottom: 24, elevation: 4,
    shadowColor: COLORS.accent, shadowOpacity: 0.2, shadowRadius: 12,
  },
  qrPlaceholder: { width: 200, height: 200, justifyContent: 'center', alignItems: 'center' },
  addressLabel: {
    fontSize: 12, color: COLORS.muted,
    textTransform: 'uppercase', letterSpacing: 1,
    marginBottom: 8, alignSelf: 'flex-start',
  },
  addressBox: {
    backgroundColor: COLORS.card, borderRadius: 10,
    padding: 14, width: '100%', marginBottom: 20,
  },
  address: { fontSize: 13, color: COLORS.text, textAlign: 'center', lineHeight: 20 },
  btnRow: { flexDirection: 'row', gap: 12, width: '100%', marginBottom: 24 },
  btn: {
    flex: 1, backgroundColor: COLORS.accent,
    borderRadius: 12, padding: 16, alignItems: 'center',
  },
  btnOutline: {
    backgroundColor: 'transparent',
    borderWidth: 1, borderColor: COLORS.accent,
  },
  btnText: { fontSize: 16, fontWeight: '700', color: COLORS.bg },
  infoBox: {
    backgroundColor: COLORS.card, borderRadius: 10,
    padding: 14, width: '100%', borderLeftWidth: 3, borderLeftColor: COLORS.accent,
  },
  infoText: { fontSize: 12, color: COLORS.muted, lineHeight: 18 },
});
