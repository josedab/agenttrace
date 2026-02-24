"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface Permission {
  resource: string;
  actions: string[];
}

export interface RoleAssignment {
  userId: string;
  role: string;
  assignedAt: string;
}

export interface SSOConfig {
  provider: string;
  enabled: boolean;
  domain: string;
  clientId: string;
}

export function usePermissions(role?: string) {
  return useQuery({
    queryKey: ["rbac", "permissions", role],
    queryFn: () => api.rbac.getPermissions(role),
  });
}

export function useAssignRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { userId: string; role: string }) =>
      api.rbac.assignRole(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["rbac"] }),
  });
}

export function useCheckPermission() {
  return useMutation({
    mutationFn: (data: { userId: string; resource: string; action: string }) =>
      api.rbac.checkPermission(data),
  });
}

export function useSSOConfig() {
  return useQuery({
    queryKey: ["rbac", "sso"],
    queryFn: () => api.rbac.getSSOConfig(),
  });
}

export function useConfigureSSO() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { provider: string; domain: string; clientId: string; clientSecret: string }) =>
      api.rbac.configureSSO(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["rbac", "sso"] }),
  });
}
