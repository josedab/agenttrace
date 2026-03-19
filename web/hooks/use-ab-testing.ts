"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface ABTest {
  id: string;
  name: string;
  description: string;
  status: "draft" | "running" | "paused" | "completed";
  variants: ABVariant[];
  startedAt?: string;
  completedAt?: string;
  createdAt: string;
}

export interface ABVariant {
  id: string;
  name: string;
  weight: number;
  config: Record<string, unknown>;
}

export interface ABTestStatistics {
  testId: string;
  sampleSize: number;
  variants: {
    variantId: string;
    name: string;
    conversions: number;
    conversionRate: number;
    confidence: number;
  }[];
  winner?: string;
  significanceLevel: number;
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
    mutationFn: (data: { name: string; description: string; variants: Omit<ABVariant, "id">[] }) =>
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
      api.abTests.assignVariant(data),
  });
}

export function useRecordABResult() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { testId: string; variantId: string; result: Record<string, unknown> }) =>
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
  });
}

export function useSelectWinner() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ testId, data }: { testId: string; data: { variantId: string } }) =>
      api.abTests.selectWinner(testId, data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["ab-tests"] }),
  });
}

export function useStartRollout() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ testId, data }: { testId: string; data: { variantId: string; percentage: number } }) =>
      api.abTests.startRollout(testId, data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["ab-tests"] }),
  });
}
