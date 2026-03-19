"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface Anomaly {
  id: string;
  type: string;
  severity: "low" | "medium" | "high" | "critical";
  description: string;
  status: "detected" | "acknowledged" | "resolved";
  detectedAt: string;
  metadata: Record<string, unknown>;
}

export interface AlertChannel {
  id: string;
  name: string;
  type: string;
  config: Record<string, unknown>;
  enabled: boolean;
}

export interface CorrelationRule {
  id: string;
  name: string;
  conditions: Record<string, unknown>[];
  action: string;
  enabled: boolean;
}

export interface AlertDashboardStats {
  totalAnomalies: number;
  activeAlerts: number;
  channelCount: number;
  correlationRuleCount: number;
  recentAnomalies: Anomaly[];
}

export interface Investigation {
  id: string;
  anomalyId: string;
  status: "open" | "in_progress" | "resolved";
  findings: string;
  assignee: string;
  createdAt: string;
  updatedAt: string;
}

export function useAnomalies() {
  return useQuery({
    queryKey: ["anomalies"],
    queryFn: () => api.anomaly.list() as Promise<Anomaly[]>,
  });
}

export function useAnomaly(id: string) {
  return useQuery({
    queryKey: ["anomalies", id],
    queryFn: () => api.anomaly.get(id) as Promise<Anomaly>,
    enabled: !!id,
  });
}

export function useDetectAnomaly() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { type: string; data: Record<string, unknown> }) =>
      api.anomaly.detect(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["anomalies"] }),
  });
}

export function useAcknowledgeAnomaly() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.anomaly.acknowledge(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["anomalies"] }),
  });
}

export function useAlertChannels() {
  return useQuery({
    queryKey: ["alert-channels"],
    queryFn: () => api.anomaly.listChannels() as Promise<AlertChannel[]>,
  });
}

export function useCreateAlertChannel() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; type: string; config: Record<string, unknown> }) =>
      api.anomaly.createChannel(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["alert-channels"] }),
  });
}

export function useTestAlertChannel() {
  return useMutation({
    mutationFn: (id: string) => api.anomaly.testChannel(id),
  });
}

export function useCorrelationRules() {
  return useQuery({
    queryKey: ["correlation-rules"],
    queryFn: () => api.anomaly.listCorrelationRules() as Promise<CorrelationRule[]>,
  });
}

export function useCreateCorrelationRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; conditions: Record<string, unknown>[]; action: string }) =>
      api.anomaly.createCorrelationRule(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["correlation-rules"] }),
  });
}

export function useAlertDashboardStats() {
  return useQuery({
    queryKey: ["alert-dashboard"],
    queryFn: () => api.anomaly.getDashboard() as Promise<AlertDashboardStats>,
  });
}

export function useInvestigations() {
  return useQuery({
    queryKey: ["investigations"],
    queryFn: () => api.anomaly.listInvestigations() as Promise<Investigation[]>,
  });
}

export function useCreateInvestigation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { anomalyId: string; assignee?: string }) =>
      api.anomaly.createInvestigation(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["investigations"] }),
  });
}

export function useUpdateInvestigation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Investigation> }) =>
      api.anomaly.updateInvestigation(id, data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["investigations"] }),
  });
}
