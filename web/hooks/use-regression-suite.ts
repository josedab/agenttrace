"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface GoldenDataset {
  id: string;
  name: string;
  description: string;
  items: GoldenDatasetItem[];
  itemCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface GoldenDatasetItem {
  id: string;
  datasetId: string;
  input: Record<string, unknown>;
  expectedOutput: Record<string, unknown>;
  metadata?: Record<string, unknown>;
}

export interface RegressionRun {
  id: string;
  datasetId: string;
  status: "pending" | "running" | "completed" | "failed";
  results: RegressionResult[];
  passRate: number;
  startedAt: string;
  completedAt?: string;
}

export interface RegressionResult {
  id: string;
  itemId: string;
  passed: boolean;
  actualOutput: Record<string, unknown>;
  comparison: BaselineComparison;
  durationMs: number;
}

export interface BaselineComparison {
  matchScore: number;
  diffs: { field: string; expected: unknown; actual: unknown }[];
  regressionDetected: boolean;
}

export function useCreateGoldenDataset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; description: string; items: Omit<GoldenDatasetItem, "id" | "datasetId">[] }) =>
      api.regressionSuite.createDataset(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["golden-datasets"] }),
  });
}

export function useGoldenDataset(id: string | null) {
  return useQuery({
    queryKey: ["golden-dataset", id],
    queryFn: () =>
      api.regressionSuite.getDataset(id!) as Promise<GoldenDataset>,
    enabled: !!id,
  });
}

export function useGoldenDatasets() {
  return useQuery({
    queryKey: ["golden-datasets"],
    queryFn: () =>
      api.regressionSuite.listDatasets() as Promise<GoldenDataset[]>,
  });
}

export function useRunRegression() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { datasetId: string; config?: Record<string, unknown> }) =>
      api.regressionSuite.runRegression(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["regression-runs"] }),
  });
}

export function useRegressionRun(id: string | null) {
  return useQuery({
    queryKey: ["regression-run", id],
    queryFn: () =>
      api.regressionSuite.getRun(id!) as Promise<RegressionRun>,
    enabled: !!id,
  });
}

export function useRegressionRuns() {
  return useQuery({
    queryKey: ["regression-runs"],
    queryFn: () =>
      api.regressionSuite.listRuns() as Promise<RegressionRun[]>,
  });
}
