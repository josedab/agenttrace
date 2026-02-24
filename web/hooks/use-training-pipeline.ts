"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface Dataset {
  id: string;
  name: string;
  traceCount: number;
  format: string;
  status: "building" | "ready" | "exported";
  createdAt: string;
}

export interface FailurePattern {
  pattern: string;
  frequency: number;
  severity: string;
  affectedTraces: string[];
  suggestedFix: string;
}

export function useTrainingDatasets() {
  return useQuery({
    queryKey: ["training", "datasets"],
    queryFn: () => api.training.listDatasets(),
  });
}

export function useCreateDataset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; filters?: Record<string, any>; format?: string }) =>
      api.training.createDataset(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["training", "datasets"] }),
  });
}

export function useExportDataset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.training.exportDataset(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["training", "datasets"] }),
  });
}

export function useFailurePatterns() {
  return useQuery({
    queryKey: ["training", "failure-patterns"],
    queryFn: () => api.training.detectFailures(),
  });
}
