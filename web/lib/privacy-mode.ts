import { api } from '@/lib/api';

export interface PrivacyCapabilities {
  mode: 'standard' | 'local_private';
  noEgress: boolean;
  redactionEnabled: boolean;
  capabilities: Record<
    string,
    {
      available: boolean;
      reason?: string;
    }
  >;
}

export const privacyModeApi = {
  getCapabilities: () => api.get<PrivacyCapabilities>('/api/public/privacy/capabilities'),
};
