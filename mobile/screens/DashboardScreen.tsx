import React from 'react';
import { View, Text, StyleSheet, ScrollView, RefreshControl } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';

export function DashboardScreen() {
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['mobile-dashboard'],
    queryFn: api.dashboard,
    refetchInterval: 30000,
  });

  return (
    <ScrollView
      style={styles.container}
      refreshControl={<RefreshControl refreshing={isLoading} onRefresh={refetch} />}
    >
      <Text style={styles.title}>AgentTrace</Text>
      <View style={styles.metricsGrid}>
        <MetricCard label="Active Alerts" value={data?.activeAlerts ?? 0} color="#ef4444" />
        <MetricCard label="Pending Reviews" value={data?.pendingReviews ?? 0} color="#f59e0b" />
        <MetricCard label="Today's Cost" value={`$${(data?.todayCost ?? 0).toFixed(2)}`} color="#3b82f6" />
        <MetricCard label="Today's Traces" value={data?.todayTraces ?? 0} color="#10b981" />
      </View>
    </ScrollView>
  );
}

function MetricCard({ label, value, color }: { label: string; value: any; color: string }) {
  return (
    <View style={[styles.card, { borderLeftColor: color }]}>
      <Text style={styles.cardLabel}>{label}</Text>
      <Text style={styles.cardValue}>{value}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f8fafc', padding: 16 },
  title: { fontSize: 24, fontWeight: 'bold', marginBottom: 16 },
  metricsGrid: { flexDirection: 'row', flexWrap: 'wrap', gap: 12 },
  card: { backgroundColor: '#fff', borderRadius: 8, padding: 16, width: '47%', borderLeftWidth: 4, shadowColor: '#000', shadowOpacity: 0.05, shadowRadius: 4, elevation: 2 },
  cardLabel: { fontSize: 12, color: '#6b7280', marginBottom: 4 },
  cardValue: { fontSize: 20, fontWeight: 'bold' },
});
