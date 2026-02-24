"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function usePIIConfig() {
  return useQuery({
    queryKey: ["privacy-config"],
    queryFn: () => api.privacy.getConfig(),
  });
}

export function useUpdatePIIConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { piiTypes: string[]; redactionMode?: string; residencyRegion?: string }) =>
      api.privacy.updateConfig(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["privacy-config"] }),
  });
}

export function useScanPII() {
  return useMutation({
    mutationFn: (data: { text: string; piiTypes?: string[] }) =>
      api.privacy.scan(data),
  });
}

export function useDeletionRequests() {
  return useQuery({
    queryKey: ["privacy-deletion-requests"],
    queryFn: () => api.privacy.listDeletionRequests(),
  });
}

export function useRequestDeletion() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { subjectId: string; reason?: string }) =>
      api.privacy.requestDeletion(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["privacy-deletion-requests"] }),
  });
}
