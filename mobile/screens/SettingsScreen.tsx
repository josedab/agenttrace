import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

export function SettingsScreen() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Settings</Text>
      <View style={styles.item}><Text>API URL</Text><Text style={styles.value}>http://localhost:8080</Text></View>
      <View style={styles.item}><Text>Push Notifications</Text><Text style={styles.value}>Enabled</Text></View>
      <View style={styles.item}><Text>Version</Text><Text style={styles.value}>0.1.0</Text></View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f8fafc', padding: 16 },
  title: { fontSize: 20, fontWeight: 'bold', marginBottom: 16 },
  item: { flexDirection: 'row', justifyContent: 'space-between', backgroundColor: '#fff', padding: 16, marginBottom: 1, borderRadius: 4 },
  value: { color: '#6b7280' },
});
