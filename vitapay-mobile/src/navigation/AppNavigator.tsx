import React, { useState, useEffect } from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createStackNavigator } from '@react-navigation/stack';
import { Text, ActivityIndicator, View } from 'react-native';

import HomeScreen from '../screens/HomeScreen';
import SendScreen from '../screens/SendScreen';
import ReceiveScreen from '../screens/ReceiveScreen';
import StakeScreen from '../screens/StakeScreen';
import PayScreen from '../screens/PayScreen';
import TransactionHistoryScreen from '../screens/TransactionHistoryScreen';
import WalletSetupScreen from '../screens/WalletSetupScreen';
import { getAddress } from '../lib/storage';
import { HomeIcon, SendIcon, ReceiveIcon, StakeIcon, HistoryIcon } from '../components/TabIcons';

const Tab = createBottomTabNavigator();
const Stack = createStackNavigator();

const COLORS = { bg: '#0a0a0a', card: '#141414', accent: '#00ff88', text: '#ffffff', muted: '#888888' };

const TAB_ICON_MAP: Record<string, React.FC<{ color: string; size?: number }>> = {
  Home: HomeIcon,
  Send: SendIcon,
  Receive: ReceiveIcon,
  Stake: StakeIcon,
  History: HistoryIcon,
};

function MainTabs() {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        headerStyle: { backgroundColor: COLORS.bg },
        headerTintColor: COLORS.text,
        tabBarStyle: { backgroundColor: COLORS.card, borderTopColor: '#222', paddingTop: 6, height: 88 },
        tabBarActiveTintColor: COLORS.accent,
        tabBarInactiveTintColor: COLORS.muted,
        tabBarIcon: ({ color }) => {
          const IconComponent = TAB_ICON_MAP[route.name];
          return IconComponent ? <IconComponent color={color} size={22} /> : null;
        },
      })}
    >
      <Tab.Screen name="Home" component={HomeScreen} />
      <Tab.Screen name="Send" component={SendScreen} />
      <Tab.Screen name="Receive" component={ReceiveScreen} />
      <Tab.Screen name="Stake" component={StakeScreen} />
      <Tab.Screen name="History" component={TransactionHistoryScreen} />
    </Tab.Navigator>
  );
}

function PayStack() {
  return (
    <Stack.Navigator screenOptions={{ headerStyle: { backgroundColor: COLORS.bg }, headerTintColor: COLORS.text }}>
      <Stack.Screen name="Main" component={MainTabs} options={{ headerShown: false }} />
      <Stack.Screen name="Pay" component={PayScreen} options={{ title: 'Pay with VITAPAY' }} />
      <Stack.Screen name="WalletSetup" component={WalletSetupScreen} options={{ headerShown: false }} />
    </Stack.Navigator>
  );
}

export default function AppNavigator() {
  const [hasWallet, setHasWallet] = useState<boolean | null>(null);

  useEffect(() => {
    getAddress().then(addr => setHasWallet(!!addr)).catch(() => setHasWallet(false));
  }, []);

  if (hasWallet === null) {
    return (
      <View style={{ flex: 1, backgroundColor: '#0a0a0a', justifyContent: 'center', alignItems: 'center' }}>
        <ActivityIndicator size="large" color="#00ff88" />
      </View>
    );
  }

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false }}>
        {!hasWallet ? (
          <Stack.Screen name="WalletSetup" component={WalletSetupScreen} />
        ) : null}
        <Stack.Screen name="App" component={PayStack} />
      </Stack.Navigator>
    </NavigationContainer>
  );
}
