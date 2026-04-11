import React, { useState, useMemo } from 'react';
import {
  View, Text, TextInput, TouchableOpacity, StyleSheet, Alert, ScrollView, ActivityIndicator,
} from 'react-native';
import { generateWallet as generateWalletLib, importWallet as importWalletLib } from '../lib/wallet';
import { saveMnemonic, saveAddress } from '../lib/storage';

const COLORS = { bg: '#0a0a0a', card: '#141414', accent: '#00ff88', text: '#ffffff', muted: '#888888', warning: '#ffaa00', error: '#ff4444' };

type Mode = 'choose' | 'create' | 'backup' | 'verify' | 'import';

/** Pick 4 unique random indices from 0..23, sorted ascending */
function pickQuizIndices(): number[] {
  const indices = new Set<number>();
  while (indices.size < 4) indices.add(Math.floor(Math.random() * 24));
  return Array.from(indices).sort((a, b) => a - b);
}

export default function WalletSetupScreen({ navigation }: any) {
  const [mode, setMode] = useState<Mode>('choose');
  const [mnemonic, setMnemonic] = useState('');
  const [importMnemonic, setImportMnemonic] = useState('');
  const [address, setAddress] = useState('');
  const [loading, setLoading] = useState(false);

  // Backup verification quiz state
  const quizIndices = useMemo(() => pickQuizIndices(), [mnemonic]);
  const [quizAnswers, setQuizAnswers] = useState<string[]>(['', '', '', '']);
  const [quizFailed, setQuizFailed] = useState(false);

  const words = mnemonic.split(' ');

  const onCreate = async () => {
    setLoading(true);
    try {
      const w = await generateWalletLib();
      setMnemonic(w.mnemonic);
      setAddress(w.address);
      setQuizAnswers(['', '', '', '']);
      setQuizFailed(false);
      setMode('backup');
    } catch (e: any) {
      Alert.alert('Error', e.message);
    } finally {
      setLoading(false);
    }
  };

  const onProceedToVerify = () => {
    setMode('verify');
  };

  const onVerifyBackup = async () => {
    // Check all 4 answers match the correct words
    const allCorrect = quizIndices.every(
      (wordIdx, i) => quizAnswers[i].trim().toLowerCase() === words[wordIdx]?.toLowerCase()
    );
    if (!allCorrect) {
      setQuizFailed(true);
      Alert.alert('Incorrect', 'Some words are wrong. Please check your backup and try again.');
      setQuizAnswers(['', '', '', '']);
      return;
    }
    setQuizFailed(false);
    setLoading(true);
    try {
      await saveMnemonic(mnemonic);
      await saveAddress(address);
      navigation.replace('App');
    } catch (e: any) {
      Alert.alert('Error', `Failed to save wallet: ${e.message}`);
    } finally {
      setLoading(false);
    }
  };

  const onImport = async () => {
    setLoading(true);
    try {
      const w = await importWalletLib(importMnemonic);
      await saveMnemonic(importMnemonic.trim());
      await saveAddress(w.address);
      navigation.replace('App');
    } catch (e: any) {
      Alert.alert('Import failed', e.message);
    } finally {
      setLoading(false);
    }
  };

  if (mode === 'choose') return (
    <View style={styles.container}>
      <Text style={styles.logo}>⚡ VITAPAY</Text>
      <Text style={styles.subtitle}>Your VITA Wallet</Text>
      <View style={styles.btnGroup}>
        <TouchableOpacity style={styles.btn} onPress={() => setMode('create')}>
          <Text style={styles.btnText}>Create New Wallet</Text>
        </TouchableOpacity>
        <TouchableOpacity style={[styles.btn, styles.btnOutline]} onPress={() => setMode('import')}>
          <Text style={[styles.btnText, { color: COLORS.accent }]}>Import Existing</Text>
        </TouchableOpacity>
      </View>
    </View>
  );

  if (mode === 'create') return (
    <View style={styles.container}>
      <Text style={styles.title}>Create Wallet</Text>
      <Text style={styles.body}>We'll generate a secure 24-word seed phrase. Write it down and store it safely — it's the only way to recover your wallet.</Text>
      <TouchableOpacity style={[styles.btn, loading && styles.btnDisabled]} onPress={onCreate} disabled={loading}>
        {loading ? <ActivityIndicator color={COLORS.bg} /> : <Text style={styles.btnText}>Generate Wallet</Text>}
      </TouchableOpacity>
    </View>
  );

  if (mode === 'backup') return (
    <ScrollView style={styles.container}>
      <Text style={styles.title}>⚠️ Back Up Your Seed</Text>
      <Text style={styles.warning}>Write down these 24 words IN ORDER. Never share them. If lost, your funds cannot be recovered.</Text>
      <View style={styles.mnemonicBox}>
        {words.map((w, i) => (
          <Text key={i} style={styles.mnemonicWord}>
            <Text style={styles.mnemonicNum}>{i + 1}. </Text>{w}
          </Text>
        ))}
      </View>
      <Text style={styles.addrLabel}>Your address: <Text style={{ color: COLORS.accent }}>{address}</Text></Text>
      <TouchableOpacity style={styles.btn} onPress={onProceedToVerify}>
        <Text style={styles.btnText}>I've Written It Down →</Text>
      </TouchableOpacity>
    </ScrollView>
  );

  if (mode === 'verify') return (
    <ScrollView style={styles.container}>
      <Text style={styles.title}>✅ Verify Your Backup</Text>
      <Text style={styles.body}>Enter the following words from your seed phrase to confirm you have it backed up.</Text>
      {quizFailed && (
        <Text style={{ color: COLORS.error, marginBottom: 12 }}>❌ Incorrect words. Try again.</Text>
      )}
      {quizIndices.map((wordIdx, i) => (
        <View key={i} style={{ marginBottom: 14 }}>
          <Text style={styles.quizLabel}>Word #{wordIdx + 1}</Text>
          <TextInput
            style={styles.input}
            value={quizAnswers[i]}
            onChangeText={(v) => {
              const updated = [...quizAnswers];
              updated[i] = v;
              setQuizAnswers(updated);
            }}
            placeholder={`Enter word #${wordIdx + 1}`}
            placeholderTextColor={COLORS.muted}
            autoCapitalize="none"
            autoCorrect={false}
          />
        </View>
      ))}
      <TouchableOpacity
        style={[styles.btn, loading && styles.btnDisabled]}
        onPress={onVerifyBackup}
        disabled={loading}
      >
        {loading ? <ActivityIndicator color={COLORS.bg} /> : <Text style={styles.btnText}>Confirm & Open Wallet</Text>}
      </TouchableOpacity>
    </ScrollView>
  );

  // import mode
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Import Wallet</Text>
      <Text style={styles.body}>Enter your 12 or 24-word seed phrase separated by spaces.</Text>
      <TextInput
        style={[styles.input, { height: 120, textAlignVertical: 'top' }]}
        value={importMnemonic} onChangeText={setImportMnemonic}
        placeholder="Enter seed phrase..." placeholderTextColor={COLORS.muted}
        multiline autoCapitalize="none" autoCorrect={false}
      />
      <TouchableOpacity style={[styles.btn, loading && styles.btnDisabled]} onPress={onImport} disabled={loading}>
        {loading ? <ActivityIndicator color={COLORS.bg} /> : <Text style={styles.btnText}>Import Wallet</Text>}
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: COLORS.bg, padding: 24, paddingTop: 60 },
  logo: { fontSize: 40, fontWeight: 'bold', color: COLORS.accent, textAlign: 'center', marginTop: 60 },
  subtitle: { fontSize: 18, color: COLORS.muted, textAlign: 'center', marginBottom: 60 },
  title: { fontSize: 26, fontWeight: 'bold', color: COLORS.text, marginBottom: 16 },
  body: { fontSize: 15, color: COLORS.muted, lineHeight: 22, marginBottom: 24 },
  warning: { fontSize: 14, color: COLORS.warning, lineHeight: 20, marginBottom: 16 },
  btnGroup: { gap: 16 },
  btn: { backgroundColor: COLORS.accent, borderRadius: 14, padding: 18, alignItems: 'center', marginBottom: 16 },
  btnOutline: { backgroundColor: 'transparent', borderWidth: 1, borderColor: COLORS.accent },
  btnDisabled: { opacity: 0.5 },
  btnText: { fontSize: 17, fontWeight: '700', color: COLORS.bg },
  mnemonicBox: { backgroundColor: COLORS.card, borderRadius: 12, padding: 16, marginBottom: 16, flexDirection: 'row', flexWrap: 'wrap', gap: 4 },
  mnemonicWord: { fontSize: 14, color: COLORS.accent, width: '48%', lineHeight: 24 },
  mnemonicNum: { color: COLORS.muted },
  addrLabel: { fontSize: 12, color: COLORS.muted, marginBottom: 20 },
  quizLabel: { fontSize: 13, color: COLORS.muted, marginBottom: 6, textTransform: 'uppercase', letterSpacing: 1 },
  input: { backgroundColor: COLORS.card, color: COLORS.text, borderRadius: 10, padding: 14, fontSize: 15, borderWidth: 1, borderColor: '#222' },
});
