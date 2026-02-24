"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface OTelExportDestination {
  id: string;
  name: string;
  type: "otlp_grpc" | "otlp_http" | "jaeger" | "zipkin" | "prometheus";
  endpoint: string;
  headers?: Record<string, string>;
  enabled: boolean;
  createdAt: string;
}

export interface OTelMapping {
  id: string;
  sourceField: string;
  targetField: string;
  transform?: string;
  enabled: boolean;
}

export interface OTelCollectorConfig {
  receivers: Record<string, unknown>;
  processors: Record<string, unknown>;
  exporters: Record<string, unknown>;
  pipelines: Record<string, unknown>;
  yaml: string;
}

export interface OTelCompatDashboard {
  destinations: OTelExportDestination[];
  mappings: OTelMapping[];
  exportedSpans: number;
  exportErrors: number;
  lastExportAt?: string;
}

export function useCreateOTelDestination() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Omit<OTelExportDestination, "id" | "createdAt">) =>
      api.otelCompat.createDestination(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["otel-destinations"] }),
  });
}

export function useOTelDestinations() {
  return useQuery({
    queryKey: ["otel-destinations"],
    queryFn: () =>
      api.otelCompat.listDestinations() as Promise<OTelExportDestination[]>,
  });
}

export function useDeleteOTelDestination() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api.otelCompat.deleteDestination(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["otel-destinations"] }),
  });
}

export function useOTelMappings() {
  return useQuery({
    queryKey: ["otel-mappings"],
    queryFn: () =>
      api.otelCompat.getMappings() as Promise<OTelMapping[]>,
  });
}

export function useOTelDashboard() {
  return useQuery({
    queryKey: ["otel-dashboard"],
    queryFn: () =>
      api.otelCompat.getDashboard() as Promise<OTelCompatDashboard>,
    refetchInterval: 30000,
  });
}

export function useGenerateCollectorConfig() {
  return useMutation({
    mutationFn: () =>
      api.otelCompat.generateCollectorConfig() as Promise<OTelCollectorConfig>,
  });
}
