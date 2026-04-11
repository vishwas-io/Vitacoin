import * as LocalAuthentication from 'expo-local-authentication';

/**
 * Check if biometric authentication is available on the device.
 */
export async function isBiometricAvailable(): Promise<boolean> {
  try {
    const hasHardware = await LocalAuthentication.hasHardwareAsync();
    if (!hasHardware) return false;
    const isEnrolled = await LocalAuthentication.isEnrolledAsync();
    return isEnrolled;
  } catch (e) {
    console.warn('[biometric] availability check failed:', e);
    return false;
  }
}

/**
 * Prompt biometric authentication.
 * @param reason - Human-readable reason shown in the prompt.
 * @returns true if authenticated successfully, false otherwise.
 */
export async function authenticateWithBiometric(reason: string): Promise<boolean> {
  try {
    const available = await isBiometricAvailable();
    if (!available) {
      console.warn('[biometric] not available — skipping');
      return true; // graceful degradation: allow action if no biometric enrolled
    }

    const result = await LocalAuthentication.authenticateAsync({
      promptMessage: reason,
      fallbackLabel: 'Use PIN',
      disableDeviceFallback: false,
      cancelLabel: 'Cancel',
    });

    return result.success;
  } catch (e) {
    console.warn('[biometric] authentication error:', e);
    return false;
  }
}
