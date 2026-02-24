"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useAnnotations(traceId: string) {
  return useQuery({
    queryKey: ["annotations", traceId],
    queryFn: () => api.annotations.list(traceId),
    enabled: !!traceId,
  });
}

export function useCreateAnnotation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => api.annotations.create(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["annotations"] }),
  });
}

export function useAnnotationPresence(traceId: string) {
  return useQuery({
    queryKey: ["annotations", "presence", traceId],
    queryFn: () => api.annotations.getPresence(traceId),
    enabled: !!traceId,
    refetchInterval: 5000,
  });
}
