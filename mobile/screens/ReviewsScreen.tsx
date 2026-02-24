import React from 'react';
import { View, Text, FlatList, TouchableOpacity, StyleSheet, Alert } from 'react-native';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../lib/api';

export function ReviewsScreen() {
  const queryClient = useQueryClient();
  const { data } = useQuery({ queryKey: ['pending-reviews'], queryFn: api.pendingReviews });
  const reviews = data?.reviews ?? [];

  const decideMutation = useMutation({
    mutationFn: ({ id, action }: { id: string; action: string }) =>
      api.decideReview(id, { action }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['pending-reviews'] }),
  });

  const handleDecide = (id: string, action: string) => {
    Alert.alert(`${action} review?`, 'This action cannot be undone.', [
      { text: 'Cancel', style: 'cancel' },
      { text: action, onPress: () => decideMutation.mutate({ id, action }) },
    ]);
  };

  return (
    <View style={styles.container}>
      <FlatList
        data={reviews}
        keyExtractor={(item: any) => item.id}
        renderItem={({ item }) => (
          <View style={styles.card}>
            <View style={[styles.badge, { backgroundColor: item.riskLevel === 'critical' ? '#fee2e2' : '#fef3c7' }]}>
              <Text style={styles.badgeText}>{item.riskLevel} risk</Text>
            </View>
            <Text style={styles.actions}>{item.proposedActions?.length ?? 0} actions</Text>
            <View style={styles.buttons}>
              <TouchableOpacity style={[styles.btn, styles.approveBtn]} onPress={() => handleDecide(item.id, 'approve')}>
                <Text style={styles.btnText}>Approve</Text>
              </TouchableOpacity>
              <TouchableOpacity style={[styles.btn, styles.rejectBtn]} onPress={() => handleDecide(item.id, 'reject')}>
                <Text style={styles.rejectText}>Reject</Text>
              </TouchableOpacity>
            </View>
          </View>
        )}
        ListEmptyComponent={<Text style={styles.empty}>No pending reviews</Text>}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f8fafc' },
  card: { backgroundColor: '#fff', padding: 16, marginHorizontal: 16, marginTop: 8, borderRadius: 8 },
  badge: { alignSelf: 'flex-start', paddingHorizontal: 8, paddingVertical: 2, borderRadius: 12 },
  badgeText: { fontSize: 11, fontWeight: '600' },
  actions: { color: '#6b7280', fontSize: 12, marginVertical: 8 },
  buttons: { flexDirection: 'row', gap: 8 },
  btn: { flex: 1, paddingVertical: 8, borderRadius: 6, alignItems: 'center' },
  approveBtn: { backgroundColor: '#10b981' },
  rejectBtn: { backgroundColor: '#fee2e2' },
  btnText: { color: '#fff', fontWeight: '600', fontSize: 13 },
  rejectText: { color: '#ef4444', fontWeight: '600', fontSize: 13 },
  empty: { textAlign: 'center', color: '#9ca3af', marginTop: 40, fontSize: 14 },
});
