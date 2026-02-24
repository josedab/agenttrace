const API_URL = process.env.AGENTTRACE_API_URL || 'http://localhost:8080';

async function fetchAPI<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });
  if (!response.ok) throw new Error(`API error: ${response.status}`);
  return response.json();
}

export const api = {
  dashboard: () => fetchAPI<any>('/api/public/mobile/dashboard'),
  notifications: () => fetchAPI<any>('/api/public/mobile/notifications'),
  registerDevice: (data: { platform: string; pushToken: string }) =>
    fetchAPI<any>('/api/public/mobile/devices', { method: 'POST', body: JSON.stringify(data) }),
  pendingReviews: () => fetchAPI<any>('/api/public/sandbox/reviews/pending'),
  decideReview: (id: string, decision: any) =>
    fetchAPI<any>(`/api/public/sandbox/reviews/${id}/decide`, { method: 'POST', body: JSON.stringify(decision) }),
  anomalyDashboard: () => fetchAPI<any>('/api/public/anomaly/dashboard'),
};
