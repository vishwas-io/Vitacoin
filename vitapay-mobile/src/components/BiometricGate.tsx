import React, { useEffect, useState } from 'react';
import { View, Text, TextInput, StyleSheet, Alert, ActivityIndicator } from 'react-native';
import { authenticateWithBiometric, isBiometricAvailable } from '../lib/biometric';

interface BiometricGateProps {
  reason: string;
  children: React.ReactNode;
  pinFallback?: boolean; // if true, show 6-digit PIN fallback
}

const HARDCODED_PIN = '000000'; // In production: use SecureStore-saved PIN

export default function BiometricGate({ reason, children, pinFallback = true }: BiometricGateProps) {
  const [unlocked, setUnlocked] = useState(false);
  const [showPin, setShowPin] = useState(false);
  const [pin, setPin] = useState('');
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    attemptBiometric();
  }, []);

  async function attemptBiometric() {
    setChecking(true);
    const available = await isBiometricAvailable();
    if (!available) {
      // No biometric — skip gate entirely
      console.warn('[BiometricGate] biometrics not enrolled — unlocking without auth');
      setUnlocked(true);
      setChecking(false);
      return;
    }
    const success = await authenticateWithBiometric(reason);
    if (success) {
      setUnlocked(true);
    } else if (pinFallback) {
      setShowPin(true);
    } else {
      Alert.alert('Authentication Failed', 'Could not verify identity.');
    }
    setChecking(false);
  }

  function handlePinSubmit(value: string) {
    if (value === HARDCODED_PIN) {
      setUnlocked(true);
      setShowPin(false);
    } else if (value.length === 6) {
      Alert.alert('Wrong PIN', 'Incorrect PIN. Try again.');
      setPin('');
    }
  }

  if (checking) {
    return (
      <View style={styles.overlay}>
        <ActivityIndicator size="large" color="#6C63FF" />
        <Text style={styles.label}>Verifying identity…</Text>
      </View>
    );
  }

  if (showPin) {
    return (
      <View style={styles.overlay}>
        <Text style={styles.title}>Enter PIN</Text>
        <Text style={styles.label}>{reason}</Text>
        <TextInput
          style={styles.pin}
          keyboardType="numeric"
          secureTextEntry
          maxLength={6}
          value={pin}
          onChangeText={(v) => {
            setPin(v);
            handlePinSubmit(v);
          }}
          autoFocus
          placeholder="6-digit PIN"
          placeholderTextColor="#888"
        />
      </View>
    );
  }

  if (!unlocked) {
    return (
      <View style={styles.overlay}>
        <Text style={styles.title}>🔒 Locked</Text>
        <Text style={styles.label}>Authentication required to continue.</Text>
      </View>
    );
  }

  return <>{children}</>;
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: '#0D0D1A',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 32,
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#FFFFFF',
    marginBottom: 12,
  },
  label: {
    fontSize: 16,
    color: '#AAAACC',
    textAlign: 'center',
    marginTop: 8,
  },
  pin: {
    marginTop: 24,
    backgroundColor: '#1A1A2E',
    color: '#FFFFFF',
    fontSize: 24,
    letterSpacing: 8,
    textAlign: 'center',
    borderRadius: 12,
    padding: 16,
    width: 200,
    borderWidth: 1,
    borderColor: '#6C63FF',
  },
});
