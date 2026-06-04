// identity API client.
import { fetchWithAuth, requireApiProjectId, unsupportedApiFeature } from '../transport';
import {
  APIKeyCreateResponse,
  APIKeyResponse,
  normalizeAPIKey,
} from '../normalizers';
import type {
  AuditExportJob,
  AuditLog,
  AuditLogListParams,
  AuditSummary,
  CreateAPIKeyInput,
  CreateProjectInput,
  CreateSSOInput,
  Organization,
  Project,
  ProjectSettings,
  SSOConfiguration,
  TeamMember,
  TeamRole,
  UpdateProjectInput,
  UpdateProjectSettingsInput,
  UpdateSSOInput,
  UpdateUserProfileInput,
  User,
  UserProfile,
} from '../contracts';

export const identityApi = {
  // Auth
  auth: {
    login: (data: { email: string; password: string }) =>
      fetchWithAuth<{ token: string; user: User }>('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify(data),
      }),

    register: (data: { email: string; password: string; name: string }) =>
      fetchWithAuth<{ token: string; user: User }>('/api/auth/register', {
        method: 'POST',
        body: JSON.stringify(data),
      }),

    refresh: (refreshToken: string) =>
      fetchWithAuth<{ token: string }>('/api/auth/refresh', {
        method: 'POST',
        body: JSON.stringify({ refreshToken }),
      }),

    forgotPassword: (_email: string) =>
      unsupportedApiFeature<{ message: string }>('Password recovery'),
  },
  user: {
    getProfile: () =>
      fetchWithAuth<{
        id: string;
        name: string;
        email: string;
        image?: string;
      }>('/api/v1/me').then(
        (user) =>
          ({
            id: user.id,
            name: user.name,
            email: user.email,
            bio: undefined,
            avatar: user.image,
          }) satisfies UserProfile
      ),

    updateProfile: (_data: UpdateUserProfileInput) =>
      unsupportedApiFeature<UserProfile>('Profile updates'),
  },
  // Organizations
  organizations: {
    list: () =>
      fetchWithAuth<{ data: Organization[] }>('/api/v1/organizations').then(
        (response) => response.data
      ),

    get: (id: string) => fetchWithAuth<Organization>(`/api/v1/organizations/${id}`),

    create: (data: { name: string }) =>
      fetchWithAuth<Organization>('/api/v1/organizations', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Projects
  projects: {
    list: (orgId?: string) => {
      const params = orgId ? `?organizationId=${orgId}` : '';
      return fetchWithAuth<{ data: Project[] }>(`/api/v1/projects${params}`).then(
        (response) => response.data
      );
    },

    get: (id: string) => fetchWithAuth<Project>(`/api/v1/projects/${id}`),

    create: (data: CreateProjectInput) =>
      fetchWithAuth<Project>('/api/v1/projects', {
        method: 'POST',
        body: JSON.stringify(data),
      }),

    update: (id: string, data: UpdateProjectInput) =>
      fetchWithAuth<Project>(`/api/v1/projects/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
  },
  project: {
    getSettings: () => {
      const projectId = requireApiProjectId();
      return fetchWithAuth<Project>(`/api/v1/projects/${projectId}`).then(
        (project): ProjectSettings => ({
          id: project.id,
          name: project.name,
          description: project.description ?? undefined,
          defaultRetentionDays: project.retentionDays,
          publicDashboard: project.settings?.publicDashboard === true,
        })
      );
    },

    updateSettings: (data: UpdateProjectSettingsInput) => {
      const projectId = requireApiProjectId();
      return fetchWithAuth<Project>(`/api/v1/projects/${projectId}`, {
        method: 'PUT',
        body: JSON.stringify({
          name: data.name,
          description: data.description,
          retentionDays: data.defaultRetentionDays,
          settings: { publicDashboard: data.publicDashboard },
        }),
      }).then(
        (project): ProjectSettings => ({
          id: project.id,
          name: project.name,
          description: project.description ?? undefined,
          defaultRetentionDays: project.retentionDays,
          publicDashboard: project.settings?.publicDashboard === true,
        })
      );
    },
  },
  // API Keys
  apiKeys: {
    list: async () => {
      const projectId = requireApiProjectId();
      const response = await fetchWithAuth<{ data: APIKeyResponse[] }>(
        `/api/v1/projects/${projectId}/api-keys`
      );
      return response.data.map(normalizeAPIKey);
    },

    create: async (data: CreateAPIKeyInput) => {
      const projectId = requireApiProjectId();
      const { expiresIn, ...request } = data;
      if (!request.expiresAt && expiresIn) {
        const match = /^(\d+)d$/.exec(expiresIn);
        if (!match) {
          throw new Error(`Invalid API key expiration: ${expiresIn}`);
        }
        request.expiresAt = new Date(
          Date.now() + Number(match[1]) * 24 * 60 * 60 * 1000
        ).toISOString();
      }
      const response = await fetchWithAuth<APIKeyCreateResponse>(
        `/api/v1/projects/${projectId}/api-keys`,
        {
          method: 'POST',
          body: JSON.stringify(request),
        }
      );
      return {
        ...normalizeAPIKey(response),
        key: response.secretKey,
      };
    },

    delete: (id: string) => {
      const projectId = requireApiProjectId();
      return fetchWithAuth<void>(`/api/v1/projects/${projectId}/api-keys/${id}`, {
        method: 'DELETE',
      });
    },
  },
  // SSO
  sso: {
    get: (_organizationId: string) =>
      unsupportedApiFeature<SSOConfiguration>('Single sign-on configuration'),

    list: (_organizationId: string) =>
      unsupportedApiFeature<SSOConfiguration[]>('Single sign-on configuration'),

    create: (_organizationId: string, _data: CreateSSOInput) =>
      unsupportedApiFeature<SSOConfiguration>('Single sign-on configuration'),

    update: (_organizationId: string, _data: UpdateSSOInput) =>
      unsupportedApiFeature<SSOConfiguration>('Single sign-on configuration'),

    delete: (_organizationId: string) =>
      unsupportedApiFeature<void>('Single sign-on configuration'),

    test: (_organizationId: string) =>
      unsupportedApiFeature<{ success: boolean; message: string }>('Single sign-on configuration'),
  },
  // Audit Logs
  auditLogs: {
    list: (_organizationId: string, _params?: AuditLogListParams) =>
      unsupportedApiFeature<{ logs: AuditLog[]; nextCursor?: string }>('Audit logs'),

    get: (_organizationId: string, _logId: string) => unsupportedApiFeature<AuditLog>('Audit logs'),

    summary: (_organizationId: string, _params?: { startDate?: string; endDate?: string }) =>
      unsupportedApiFeature<AuditSummary>('Audit logs'),

    exportJobs: (_organizationId: string) =>
      unsupportedApiFeature<AuditExportJob[]>('Audit log exports'),

    createExport: (
      _organizationId: string,
      _params: { startDate: string; endDate: string; format?: 'json' | 'csv' }
    ) => unsupportedApiFeature<AuditExportJob>('Audit log exports'),

    downloadExport: (_organizationId: string, _jobId: string) =>
      unsupportedApiFeature<unknown>('Audit log exports'),
  },
  // Team Intelligence
  team: {
    listMembers: () => unsupportedApiFeature<TeamMember[]>('Team membership listing'),

    inviteMember: (_email: string, _role: TeamRole) =>
      unsupportedApiFeature<TeamMember>('Team invitations'),

    updateMemberRole: (_memberId: string, _role: TeamRole) =>
      unsupportedApiFeature<TeamMember>('Team role updates'),

    removeMember: (_memberId: string) => unsupportedApiFeature<void>('Team member removal'),

    getDashboard: () => fetchWithAuth<unknown>('/api/public/team/dashboard'),
    calculateROI: (hourlyRate?: number) =>
      fetchWithAuth<unknown>(`/api/public/team/roi${hourlyRate ? `?hourlyRate=${hourlyRate}` : ''}`),
  },
  // RBAC
  rbac: {
    getPermissions: (role?: string) =>
      fetchWithAuth<unknown>(`/api/public/rbac/permissions${role ? `?role=${role}` : ''}`),
    assignRole: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/rbac/roles', { method: 'POST', body: JSON.stringify(data) }),
    checkPermission: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/rbac/check', { method: 'POST', body: JSON.stringify(data) }),
    getSSOConfig: () => fetchWithAuth<unknown>('/api/public/rbac/sso'),
    configureSSO: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/rbac/sso', { method: 'POST', body: JSON.stringify(data) }),
  },
};
