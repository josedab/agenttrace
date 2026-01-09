"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, SSOConfiguration } from "@/lib/api";

export function useSSOConfiguration(organizationId: string) {
  return useQuery({
    queryKey: ["sso-configuration", organizationId],
    queryFn: () => api.sso.get(organizationId),
    enabled: !!organizationId,
    retry: false,
  });
}

export function useSSOConfigurations(organizationId: string) {
  return useQuery({
    queryKey: ["sso-configurations", organizationId],
    queryFn: () => api.sso.list(organizationId),
    enabled: !!organizationId,
  });
}

export function useCreateSSOConfiguration(organizationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateSSOConfigurationInput) =>
      api.sso.create(organizationId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sso-configuration", organizationId] });
      queryClient.invalidateQueries({ queryKey: ["sso-configurations", organizationId] });
    },
  });
}

export function useUpdateSSOConfiguration(organizationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateSSOConfigurationInput) =>
      api.sso.update(organizationId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sso-configuration", organizationId] });
      queryClient.invalidateQueries({ queryKey: ["sso-configurations", organizationId] });
    },
  });
}

export function useDeleteSSOConfiguration(organizationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => api.sso.delete(organizationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sso-configuration", organizationId] });
      queryClient.invalidateQueries({ queryKey: ["sso-configurations", organizationId] });
    },
  });
}

export function useTestSSOConfiguration(organizationId: string) {
  return useMutation({
    mutationFn: () => api.sso.test(organizationId),
  });
}

// Input types
export interface CreateSSOConfigurationInput {
  provider: "saml" | "oidc";
  enabled?: boolean;
  issuer?: string;
  ssoUrl?: string;
  certificate?: string;
  clientId?: string;
  clientSecret?: string;
  discoveryUrl?: string;
  allowedDomains?: string[];
  defaultRole?: string;
}

export interface UpdateSSOConfigurationInput {
  enabled?: boolean;
  issuer?: string;
  ssoUrl?: string;
  certificate?: string;
  clientId?: string;
  clientSecret?: string;
  discoveryUrl?: string;
  allowedDomains?: string[];
  defaultRole?: string;
}
