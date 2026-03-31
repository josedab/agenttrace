"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface OTelBridgeConfig {
  enabled: boolean;
  ingestEnabled: boolean;
  exportEnabled: boolean;
  serviceName: string;
  resourceAttributes: Record<string, string>;
  samplingRate: number;
  batchSize: number;
  exportIntervalMs: number;
  spanMapping: OTelSpanMapping;
}

export interface OTelSpanMapping {
  traceIdField: string;
  spanIdField: string;
  parentSpanIdField: string;
  operationNameField: string;
  startTimeField: string;
  endTimeField: string;
  attributePrefix: string;
  customMappings: Record<string, string>;
}

export interface OTelDestination {
  id: string;
  name: string;
  type: "otlp_grpc" | "otlp_http" | "jaeger" | "zipkin" | "datadog" | "grafana_tempo" | "custom";
  endpoint: string;
  protocol: string;
  headers: Record<string, string>;
  enabled: boolean;
  status: "connected" | "disconnected" | "error";
  lastExportAt?: string;
  spanCount: number;
  errorCount: number;
}

export interface OTelBridgeStats {
  totalSpansIngested: number;
  totalSpansExported: number;
  ingestRatePerSec: number;
  exportRatePerSec: number;
  activeDestinations: number;
  errorRate: number;
  lastIngestAt?: string;
  lastExportAt?: string;
  uptimeSeconds: number;
}

export interface OTelCollectorConfig {
  yaml: string;
  destinations: string[];
  generatedAt: string;
}

export function useOTelBridgeConfig() {
  return useQuery({
    queryKey: ["otel-bridge-config"],
    queryFn: () => api.otelCompat.getDashboard() as Promise<OTelBridgeConfig>,
  });
}

export function useUpdateOTelBridgeConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (config: Partial<OTelBridgeConfig>) =>
      api.otelCompat.createDestination(config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["otel-bridge-config"] });
    },
  });
}

export function useOTelBridgeDestinations() {
  return useQuery({
    queryKey: ["otel-bridge-destinations"],
    queryFn: () => api.otelCompat.listDestinations() as Promise<OTelDestination[]>,
  });
}

export function useAddOTelDestination() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      name: string;
      type: OTelDestination["type"];
      endpoint: string;
      protocol?: string;
      headers?: Record<string, string>;
    }) =>
      api.otelCompat.createDestination(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["otel-bridge-destinations"] });
    },
  });
}

export function useRemoveOTelDestination() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (destinationId: string) =>
      api.otelCompat.deleteDestination(destinationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["otel-bridge-destinations"] });
    },
  });
}

export function useImportOTelSpans() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      spans: unknown[];
      resourceAttributes?: Record<string, string>;
      serviceName?: string;
    }) =>
      api.otelCompat.createDestination({ type: "import", ...data }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["otel-bridge-stats"] });
    },
  });
}

export function useOTelBridgeStats() {
  return useQuery({
    queryKey: ["otel-bridge-stats"],
    queryFn: () => api.otelCompat.getDashboard() as Promise<OTelBridgeStats>,
    refetchInterval: 10000,
  });
}

export function useGenerateCollectorConfig() {
  return useMutation({
    mutationFn: () =>
      api.otelCompat.generateCollectorConfig() as Promise<OTelCollectorConfig>,
  });
}

export function useOTelSpanMappings() {
  return useQuery({
    queryKey: ["otel-span-mappings"],
    queryFn: () => api.otelCompat.getMappings() as Promise<OTelSpanMapping>,
  });
}

export function useTestOTelDestination() {
  return useMutation({
    mutationFn: (destinationId: string) =>
      api.otelCompat.createDestination({ type: "test", destinationId }),
  });
}
