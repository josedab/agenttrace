// Core HTTP verbs shared by the AgentTrace API client.
import { fetchWithAuth } from '../transport';

export const coreApi = {
  get: <T = unknown>(endpoint: string) => fetchWithAuth<T>(endpoint, { method: 'GET' }),

  post: <T = unknown>(endpoint: string, data?: unknown) =>
    fetchWithAuth<T>(endpoint, {
      method: 'POST',
      body: data === undefined ? undefined : JSON.stringify(data),
    }),

  put: <T = unknown>(endpoint: string, data?: unknown) =>
    fetchWithAuth<T>(endpoint, {
      method: 'PUT',
      body: data === undefined ? undefined : JSON.stringify(data),
    }),

  patch: <T = unknown>(endpoint: string, data?: unknown) =>
    fetchWithAuth<T>(endpoint, {
      method: 'PATCH',
      body: data === undefined ? undefined : JSON.stringify(data),
    }),

  delete: <T = void>(endpoint: string) => fetchWithAuth<T>(endpoint, { method: 'DELETE' }),
};
