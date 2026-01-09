"use client";

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from "@tanstack/react-query";
import { api, AuditLog, AuditSummary, AuditExportJob } from "@/lib/api";

export interface AuditLogFilters {
  userId?: string;
  action?: string;
  resourceType?: string;
  startDate?: string;
  endDate?: string;
  limit?: number;
}

export function useAuditLogs(organizationId: string, filters?: AuditLogFilters) {
  return useInfiniteQuery({
    queryKey: ["audit-logs", organizationId, filters],
    queryFn: ({ pageParam }) =>
      api.auditLogs.list(organizationId, { ...filters, cursor: pageParam }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    enabled: !!organizationId,
  });
}

export function useAuditLog(organizationId: string, logId: string) {
  return useQuery({
    queryKey: ["audit-log", organizationId, logId],
    queryFn: () => api.auditLogs.get(organizationId, logId),
    enabled: !!organizationId && !!logId,
  });
}

export function useAuditSummary(organizationId: string, params?: { startDate?: string; endDate?: string }) {
  return useQuery({
    queryKey: ["audit-summary", organizationId, params],
    queryFn: () => api.auditLogs.summary(organizationId, params),
    enabled: !!organizationId,
  });
}

export function useAuditExportJobs(organizationId: string) {
  return useQuery({
    queryKey: ["audit-export-jobs", organizationId],
    queryFn: () => api.auditLogs.exportJobs(organizationId),
    enabled: !!organizationId,
  });
}

export function useCreateAuditExport(organizationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { startDate: string; endDate: string; format?: "json" | "csv" }) =>
      api.auditLogs.createExport(organizationId, params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["audit-export-jobs", organizationId] });
    },
  });
}

export function useDownloadAuditExport(organizationId: string) {
  return useMutation({
    mutationFn: (jobId: string) => api.auditLogs.downloadExport(organizationId, jobId),
  });
}

// Action type options for filtering
export const AUDIT_ACTIONS = [
  { value: "login", label: "Login" },
  { value: "logout", label: "Logout" },
  { value: "api_key.create", label: "API Key Created" },
  { value: "api_key.delete", label: "API Key Deleted" },
  { value: "project.create", label: "Project Created" },
  { value: "project.update", label: "Project Updated" },
  { value: "project.delete", label: "Project Deleted" },
  { value: "member.invite", label: "Member Invited" },
  { value: "member.remove", label: "Member Removed" },
  { value: "member.role_change", label: "Role Changed" },
  { value: "sso.configure", label: "SSO Configured" },
  { value: "sso.enable", label: "SSO Enabled" },
  { value: "sso.disable", label: "SSO Disabled" },
  { value: "settings.update", label: "Settings Updated" },
] as const;

export const RESOURCE_TYPES = [
  { value: "user", label: "User" },
  { value: "project", label: "Project" },
  { value: "api_key", label: "API Key" },
  { value: "organization", label: "Organization" },
  { value: "sso_config", label: "SSO Config" },
  { value: "member", label: "Member" },
] as const;
