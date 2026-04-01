"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface ABTest {
  id: string;
  name: string;
  description: string;
  status: "draft" | "running" | "paused" | "completed";
  variants: ABVariant[];
  metrics: ABMetricConfig[];
  trafficSplit: "equal" | "weighted" | "bayesian_adaptive";
  minimumSampleSize: number;
  maximumDurationDays: number;
  significanceLevel: number;
  startedAt?: string;
  completedAt?: string;
  createdAt: string;
}

export interface ABVariant {
  id: string;
  name: string;
  weight: number;
  promptTemplate: string;
  config: Record<string, unknown>;
  isControl: boolean;
}

export interface ABMetricConfig {
  name: string;
  type: "binary" | "continuous" | "count";
  direction: "higher_better" | "lower_better";
  minimumDetectableEffect: number;
  weight: number;
}

export interface ABTestStatistics {
  testId: string;
  sampleSize: number;
  variants: ABVariantStatistics[];
  winner?: string;
  winnerConfidence?: number;
  significanceLevel: number;
  isSignificant: boolean;
  powerAnalysis: {
    currentPower: number;
    requiredSampleSize: number;
    estimatedCompletionDate?: string;
  };
  metrics: ABMetricResult[];
}

export interface ABVariantStatistics {
  variantId: string;
  name: string;
  sampleSize: number;
  conversions: number;
  conversionRate: number;
  confidence: number;
  confidenceInterval: { lower: number; upper: number };
  bayesianProbability: number;
  metrics: Record<string, {
    mean: number;
    stdDev: number;
    confidenceInterval: { lower: number; upper: number };
    sampleSize: number;
  }>;
}

export interface ABMetricResult {
  metricName: string;
  controlValue: number;
  treatmentValue: number;
  absoluteDifference: number;
  relativeDifference: number;
  pValue: number;
  isSignificant: boolean;
  confidenceInterval: { lower: number; upper: number };
  effectSize: number;
}

export function useABTests() {
  return useQuery({
    queryKey: ["ab-tests"],
    queryFn: () => api.abTests.list() as Promise<ABTest[]>,
  });
}

export function useABTest(id: string) {
  return useQuery({
    queryKey: ["ab-tests", id],
    queryFn: () => api.abTests.get(id) as Promise<ABTest>,
    enabled: !!id,
  });
}

export function useCreateABTest() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      name: string;
      description: string;
      variants: Omit<ABVariant, "id">[];
      metrics?: ABMetricConfig[];
      trafficSplit?: ABTest["trafficSplit"];
      minimumSampleSize?: number;
      maximumDurationDays?: number;
      significanceLevel?: number;
    }) =>
      api.abTests.create(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["ab-tests"] }),
  });
}

export function useStartABTest() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.abTests.start(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["ab-tests"] }),
  });
}

export function usePauseABTest() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.abTests.pause(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["ab-tests"] }),
  });
}

export function useStopABTest() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.abTests.stop(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["ab-tests"] }),
  });
}

export function useAssignVariant() {
  return useMutation({
    mutationFn: (data: { testId: string; userId: string }) =>
      api.abTests.assignVariant(data) as Promise<{ variantId: string; variantName: string }>,
  });
}

export function useRecordABResult() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      testId: string;
      variantId: string;
      metrics: Record<string, number>;
      result?: Record<string, unknown>;
    }) =>
      api.abTests.recordResult(data),
    onSuccess: (_data, variables) =>
      queryClient.invalidateQueries({ queryKey: ["ab-test-statistics", variables.testId] }),
  });
}

export function useABTestStatistics(testId: string) {
  return useQuery({
    queryKey: ["ab-test-statistics", testId],
    queryFn: () => api.abTests.getStatistics(testId) as Promise<ABTestStatistics>,
    enabled: !!testId,
    refetchInterval: (query) => {
      const data = query.state.data as ABTestStatistics | undefined;
      if (data?.isSignificant) return false;
      return 10000;
    },
  });
}

export function useSelectWinner() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ testId, data }: { testId: string; data: { variantId: string; reason?: string } }) =>
      api.abTests.selectWinner(testId, data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["ab-tests"] }),
  });
}

export function useStartRollout() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ testId, data }: { testId: string; data: { variantId: string; percentage: number; rampUpDays?: number } }) =>
      api.abTests.startRollout(testId, data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["ab-tests"] }),
  });
}
