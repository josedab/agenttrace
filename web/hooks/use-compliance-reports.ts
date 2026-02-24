"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface ComplianceReport {
  id: string;
  name: string;
  type: string;
  status: "generating" | "ready" | "failed";
  generatedAt: string;
  downloadUrl?: string;
}

export interface ComplianceTemplate {
  id: string;
  name: string;
  description: string;
  type: string;
  requiredFields: string[];
}

export function useComplianceReports() {
  return useQuery({
    queryKey: ["compliance-reports"],
    queryFn: () => api.complianceReports.list(),
  });
}

export function useGenerateReport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { templateId: string; name: string; parameters?: Record<string, any> }) =>
      api.complianceReports.generate(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["compliance-reports"] }),
  });
}

export function useComplianceReport(id: string) {
  return useQuery({
    queryKey: ["compliance-reports", id],
    queryFn: () => api.complianceReports.get(id),
    enabled: !!id,
  });
}

export function useComplianceTemplates() {
  return useQuery({
    queryKey: ["compliance-reports", "templates"],
    queryFn: () => api.complianceReports.getTemplates(),
  });
}
