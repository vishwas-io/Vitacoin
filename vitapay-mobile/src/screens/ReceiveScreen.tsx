import React, { useEffect, useState } from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Alert, Share } from 'react-native';
import QRCode from 'react-native-qrcode-svg';

const COLORS = { bg: '#0a0a0a', card: '#141414', accent: '#00ff88', text: '#ffffff', muted: '#888888' };

async function getMyAddress(): Promise<string> {
  return 'vita1qg5eathl0pgdpe4ghrjkz6y9pljpz47v4ym6qy';
}

export default function ReceiveScreen() {
  const [address, setAddress] = useState('');

  useEffect(() => { getMyAddress().then(setAddress); }, []);

  const copyAddress = async () => {
    try {
      // expo-clipboard usage
      const Clipboard = require('expo-clipboard');
      await Clipboard.setStringAsync(address);
      Alert.alert('Copied', 'Address copied to clipboard');
    } catch {
      Alert.alert('Address', address);
    }
  };

  const shareAddress = () => Share.share({ message: address, title: 'My VITA Address' });

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Receive VITA</Text>
      <View style={styles.qrCard}>
        {address ? (
          <QRCode value={address} size={200} backgroundColor="#ffffff" color="#000000" />
        ) : (
          <View style={styles.qrPlaceholder}><Text style={{ color: COLORS.muted }}>Loading...</Text></View>
        )}
      </View>
      <Text style={styles.addressLabel}>Your Address</Text>
      <Text style={styles.address}>{address}</Text>
      <View style={styles.btnRow}>
        <TouchableOpacity style={styles.btn} onPress={copyAddress}>
          <Text style={styles.btnText}>Copy</Text>
        </TouchableOpacity>
        <TouchableOpacity style={[styles.btn, styles.btnOutline]} onPress={shareAddress}>
          <Text style={[styles.btnText, { color: COLORS.accent }]}>Share</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg, padding: 16, alignItems: 'center' },
  title: { fontSize: 24, fontWeight: 'bold', color: COLORS.text, marginBottom: 24, alignSelf: 'flex-start' },
  qrCard: { backgroundColor: '#ffffff', borderRadius: 16, padding: 20, marginBottom: 24 },
  qrPlaceholder: { width: 200, height: 200, justifyContent: 'center', alignItems: 'center' },
  addressLabel: { fontSize: 12, color: COLORS.muted, textTransform: 'uppercase', letterSpacing: 1, marginBottom: 8 },
  address: { fontSize: 13, color: COLORS.text, textAlign: 'center', padding: 12, backgroundColor: COLORS.card, borderRadius: 10, width: '100%' },
  btnRow: { flexDirection: 'row', gap: 12, marginTop: 24, width: '100%' },
  btn: { flex: 1, backgroundColor: COLORS.accent, borderRadius: 12, padding: 16, alignItems: 'center' },
  btnOutline: { backgroundColor: 'transparent', borderWidth: 1, borderColor: COLORS.accent },
  btnText: { fontSize: 16, fontWeight: '700', color: COLORS.bg },
});
