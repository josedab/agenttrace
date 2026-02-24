"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface AnomalyDashboard {
  activeAlerts: number;
  recentAnomalies: Anomaly[];
  stats: AnomalyStats;
  channels: AlertChannel[];
  rules: AnomalyRule[];
  healthScore: number;
}

export interface Anomaly {
  id: string;
  type: string;
  severity: string;
  detectedAt: string;
  value: number;
  expected: number;
  description: string;
  traceName?: string;
}

export interface AnomalyStats {
  totalAnomalies: number;
  activeAlerts: number;
  bySeverity: Record<string, number>;
  byType: Record<string, number>;
}

export interface AlertChannel {
  id: string;
  name: string;
  type: string;
  enabled: boolean;
}

export interface AnomalyRule {
  id: string;
  name: string;
  type: string;
  method: string;
  enabled: boolean;
  severity: string;
}

export interface RootCauseAnalysis {
  anomalyId: string;
  correlatedEvents: { type: string; description: string; correlation: number }[];
  possibleCauses: { category: string; description: string; confidence: number }[];
  recommendations: string[];
}

export function useAnomalyDashboard() {
  return useQuery({
    queryKey: ["anomaly-dashboard"],
    queryFn: () =>
      api.anomaly.getDashboard() as Promise<AnomalyDashboard>,
    refetchInterval: 30000,
  });
}

export function useRootCause(anomalyId: string | null) {
  return useQuery({
    queryKey: ["root-cause", anomalyId],
    queryFn: () =>
      api.anomaly.getRootCause(anomalyId!) as Promise<RootCauseAnalysis>,
    enabled: !!anomalyId,
  });
}

export function useCreateAlertChannel() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; type: string; config: Record<string, string> }) =>
      api.anomaly.createChannel(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["anomaly-dashboard"] }),
  });
}
