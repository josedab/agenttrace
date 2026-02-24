"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface CompliancePolicy {
  id: string;
  name: string;
  framework: string;
  rules: ComplianceRule[];
  enabled: boolean;
  createdAt: string;
}

export interface ComplianceRule {
  id: string;
  name: string;
  condition: string;
  severity: string;
  action: string;
}

export interface ComplianceScore {
  framework: string;
  score: number;
  maxScore: number;
  violations: number;
  lastEvaluated: string;
}

export function useCompliancePolicies() {
  return useQuery({
    queryKey: ["compliance-policies"],
    queryFn: () => api.complianceMonitor.listPolicies() as Promise<CompliancePolicy[]>,
    refetchInterval: 30000,
  });
}

export function useEvaluateCompliance() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { traceId: string; policyIds?: string[] }) =>
      api.complianceMonitor.evaluate(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["compliance-score"] }),
  });
}

export function useComplianceScore(framework: string | null) {
  return useQuery({
    queryKey: ["compliance-score", framework],
    queryFn: () => api.complianceMonitor.getScore(framework!) as Promise<ComplianceScore>,
    enabled: !!framework,
    refetchInterval: 30000,
  });
}
