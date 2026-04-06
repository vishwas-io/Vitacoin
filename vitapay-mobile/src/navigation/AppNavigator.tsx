import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createStackNavigator } from '@react-navigation/stack';
import { Text } from 'react-native';

import HomeScreen from '../screens/HomeScreen';
import SendScreen from '../screens/SendScreen';
import ReceiveScreen from '../screens/ReceiveScreen';
import StakeScreen from '../screens/StakeScreen';
import PayScreen from '../screens/PayScreen';
import TransactionHistoryScreen from '../screens/TransactionHistoryScreen';
import WalletSetupScreen from '../screens/WalletSetupScreen';

const Tab = createBottomTabNavigator();
const Stack = createStackNavigator();

const COLORS = { bg: '#0a0a0a', card: '#141414', accent: '#00ff88', text: '#ffffff', muted: '#888888' };

const TAB_ICONS: Record<string, string> = {
  Home: '⚡',
  Send: '↑',
  Receive: '↓',
  Stake: '🔒',
  History: '📋',
};

function MainTabs() {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        headerStyle: { backgroundColor: COLORS.bg },
        headerTintColor: COLORS.text,
        tabBarStyle: { backgroundColor: COLORS.card, borderTopColor: '#222' },
        tabBarActiveTintColor: COLORS.accent,
        tabBarInactiveTintColor: COLORS.muted,
        tabBarIcon: ({ focused }) => (
          <Text style={{ fontSize: 18, opacity: focused ? 1 : 0.5 }}>{TAB_ICONS[route.name]}</Text>
        ),
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
  // TODO in Job 2: check AsyncStorage for existing wallet and route accordingly
  const hasWallet = false;

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
