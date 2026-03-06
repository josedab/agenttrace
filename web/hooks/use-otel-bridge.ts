"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useOTelBridgeConfig() {
  return useQuery({
    queryKey: ["otel-bridge-config"],
    queryFn: () => api.get("/api/public/otel-bridge/config"),
  });
}

export function useUpdateOTelBridgeConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (config: Record<string, unknown>) =>
      api.put("/api/public/otel-bridge/config", config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["otel-bridge-config"] });
    },
  });
}

export function useOTelBridgeDestinations() {
  return useQuery({
    queryKey: ["otel-bridge-destinations"],
    queryFn: () => api.get("/api/public/otel-bridge/destinations"),
  });
}

export function useAddOTelDestination() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; endpoint: string; protocol?: string; headers?: Record<string, string> }) =>
      api.post("/api/public/otel-bridge/destinations", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["otel-bridge-destinations"] });
    },
  });
}

export function useRemoveOTelDestination() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (destinationId: string) =>
      api.delete(`/api/public/otel-bridge/destinations/${destinationId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["otel-bridge-destinations"] });
    },
  });
}

export function useImportOTelSpans() {
  return useMutation({
    mutationFn: (data: { spans: unknown[]; resourceAttributes?: Record<string, string> }) =>
      api.post("/api/public/otel-bridge/import", data),
  });
}

export function useOTelBridgeStats() {
  return useQuery({
    queryKey: ["otel-bridge-stats"],
    queryFn: () => api.get("/api/public/otel-bridge/stats"),
  });
}
