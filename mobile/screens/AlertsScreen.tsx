import React from 'react';
import { View, Text, FlatList, StyleSheet } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';

export function AlertsScreen() {
  const { data } = useQuery({ queryKey: ['notifications'], queryFn: api.notifications });
  const notifications = data?.notifications ?? [];

  return (
    <View style={styles.container}>
      <FlatList
        data={notifications}
        keyExtractor={(item: any) => item.id}
        renderItem={({ item }) => (
          <View style={styles.item}>
            <Text style={styles.itemTitle}>{item.title}</Text>
            <Text style={styles.itemBody}>{item.body}</Text>
          </View>
        )}
        ListEmptyComponent={<Text style={styles.empty}>No notifications</Text>}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f8fafc' },
  item: { backgroundColor: '#fff', padding: 16, marginHorizontal: 16, marginTop: 8, borderRadius: 8 },
  itemTitle: { fontWeight: 'bold', fontSize: 14 },
  itemBody: { color: '#6b7280', fontSize: 12, marginTop: 4 },
  empty: { textAlign: 'center', color: '#9ca3af', marginTop: 40, fontSize: 14 },
});
