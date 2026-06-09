"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useEmbedConfig() {
  return useQuery({
    queryKey: ["embed-config"],
    queryFn: () => api.embed.getConfig(),
  });
}

export function useCreateEmbedConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { allowedDomains: string[]; branding?: Record<string, unknown>; dashboardIds?: string[] }) =>
      api.embed.createConfig(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["embed-config"] }),
  });
}

export function useUpdateEmbedConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { allowedDomains?: string[]; branding?: Record<string, unknown>; dashboardIds?: string[] }) =>
      api.embed.updateConfig(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["embed-config"] }),
  });
}

export function useGenerateEmbedToken() {
  return useMutation({
    mutationFn: () => api.embed.generateToken(),
  });
}
