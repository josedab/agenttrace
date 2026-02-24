"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface SecurityScanResult {
  id: string;
  traceId: string;
  status: "clean" | "findings_detected";
  findings: SecurityFinding[];
  scannedAt: string;
  durationMs: number;
}

export interface SecurityFinding {
  id: string;
  scanId: string;
  type: "pii_leak" | "prompt_injection" | "data_exfiltration" | "unauthorized_access" | "sensitive_data";
  severity: "low" | "medium" | "high" | "critical";
  description: string;
  location: string;
  evidence: string;
  acknowledged: boolean;
  detectedAt: string;
}

export interface SecurityPolicy {
  id: string;
  name: string;
  description: string;
  rules: { type: string; pattern: string; action: "block" | "alert" | "redact" }[];
  enabled: boolean;
  createdAt: string;
}

export interface SecurityDashboard {
  totalScans: number;
  totalFindings: number;
  unresolvedFindings: number;
  bySeverity: Record<string, number>;
  recentFindings: SecurityFinding[];
  policies: SecurityPolicy[];
}

export function useScanTrace() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { traceId: string; policyIds?: string[] }) =>
      api.securityScanner.scan(data) as Promise<SecurityScanResult>,
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["security-dashboard"] }),
  });
}

export function useCreateSecurityPolicy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Omit<SecurityPolicy, "id" | "createdAt">) =>
      api.securityScanner.createPolicy(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["security-policies"] }),
  });
}

export function useSecurityPolicies() {
  return useQuery({
    queryKey: ["security-policies"],
    queryFn: () =>
      api.securityScanner.listPolicies() as Promise<SecurityPolicy[]>,
  });
}

export function useSecurityDashboard() {
  return useQuery({
    queryKey: ["security-dashboard"],
    queryFn: () =>
      api.securityScanner.getDashboard() as Promise<SecurityDashboard>,
    refetchInterval: 30000,
  });
}

export function useAcknowledgeFinding() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api.securityScanner.acknowledgeFinding(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["security-dashboard"] }),
  });
}
