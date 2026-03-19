"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface SelfHealingPolicy {
  id: string;
  name: string;
  condition: string;
  action: string;
  enabled: boolean;
  createdAt: string;
}

export interface GuardrailPipelineEvalResult {
  passed: boolean;
  violations: { rule: string; severity: string; message: string }[];
  score: number;
}

export interface GuardrailDashboardStats {
  totalPolicies: number;
  activePolicies: number;
  totalViolations: number;
  recentViolations: { id: string; policy: string; timestamp: string }[];
}

export interface PolicyAuditEntry {
  id: string;
  policyId: string;
  action: string;
  actor: string;
  timestamp: string;
  details: Record<string, unknown>;
}

export function useSelfHealingPolicies() {
  return useQuery({
    queryKey: ["self-healing-policies"],
    queryFn: () => api.guardrails.listPolicies() as Promise<{ policies: SelfHealingPolicy[] }>,
  });
}

export function useCreateSelfHealingPolicy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; condition: string; action: string }) =>
      api.guardrails.createPolicy(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["self-healing-policies"] }),
  });
}

export function useGuardrailPipelineEval() {
  return useMutation({
    mutationFn: (data: { traceId: string; policies?: string[] }) =>
      api.guardrails.evaluatePipeline(data) as Promise<GuardrailPipelineEvalResult>,
  });
}

export function useGuardrailDashboardStats() {
  return useQuery({
    queryKey: ["guardrail-dashboard"],
    queryFn: () => api.guardrails.getDashboardStats() as Promise<GuardrailDashboardStats>,
  });
}

export function usePolicyAuditTrail(policyId: string) {
  return useQuery({
    queryKey: ["policy-audit-trail", policyId],
    queryFn: () =>
      api.guardrails.getAuditTrail(policyId) as Promise<{ auditTrail: PolicyAuditEntry[] }>,
    enabled: !!policyId,
  });
}
