"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useStreamingDashboard() {
  return useQuery({
    queryKey: ["streaming-dashboard"],
    queryFn: () => api.get("/api/public/streaming-dashboard"),
    refetchInterval: 3000,
  });
}

export function useDashboardConfig() {
  return useQuery({
    queryKey: ["streaming-dashboard-config"],
    queryFn: () => api.get("/api/public/streaming-dashboard/config"),
  });
}

export function useUpdateDashboardConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (config: Record<string, unknown>) =>
      api.put("/api/public/streaming-dashboard/config", config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["streaming-dashboard-config"] });
    },
  });
}
