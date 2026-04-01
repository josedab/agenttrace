"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface BenchmarkSuite {
  id: string;
  name: string;
  description: string;
  tasks: BenchmarkTask[];
  runCount: number;
  category: string;
  difficulty: "easy" | "medium" | "hard" | "expert";
  estimatedDurationMs: number;
  createdAt: string;
  updatedAt: string;
}

export interface BenchmarkTask {
  id: string;
  name: string;
  category: string;
  description?: string;
  expectedOutput?: string;
  maxTokens?: number;
  timeoutMs?: number;
}

export interface BenchmarkRun {
  id: string;
  suiteId: string;
  agentId: string;
  agentName: string;
  agentVersion: string;
  status: "pending" | "running" | "completed" | "failed";
  progress: number;
  scores: Record<string, number>;
  taskResults: BenchmarkTaskResult[];
  overallScore: number;
  totalCost: number;
  totalTokens: number;
  totalDurationMs: number;
  startedAt: string;
  completedAt?: string;
}

export interface BenchmarkTaskResult {
  taskId: string;
  taskName: string;
  passed: boolean;
  score: number;
  latencyMs: number;
  tokensUsed: number;
  costUsd: number;
  output?: string;
  error?: string;
}

export interface BenchmarkLeaderboard {
  suiteId: string;
  suiteName: string;
  entries: LeaderboardEntry[];
  lastUpdated: string;
}

export interface LeaderboardEntry {
  rank: number;
  agentId: string;
  agentName: string;
  agentVersion: string;
  overallScore: number;
  scores: Record<string, number>;
  totalCost: number;
  totalDurationMs: number;
  runId: string;
  completedAt: string;
  delta?: number;
}

export interface BenchmarkComparison {
  suiteId: string;
  runs: BenchmarkRun[];
  metricComparison: Record<string, { runId: string; value: number }[]>;
  winner?: string;
}

export interface BenchmarkReport {
  id: string;
  suiteId: string;
  runIds: string[];
  summary: string;
  recommendations: string[];
  shareUrl?: string;
  generatedAt: string;
}

export function useBenchmarkSuites() {
  return useQuery({
    queryKey: ["benchmark-suites"],
    queryFn: () =>
      api.agentBenchmarks.listSuites() as Promise<BenchmarkSuite[]>,
  });
}

export function useBenchmarkSuite(suiteId: string | null) {
  return useQuery({
    queryKey: ["benchmark-suite", suiteId],
    queryFn: () =>
      api.agentBenchmarks.getSuite(suiteId!) as Promise<BenchmarkSuite>,
    enabled: !!suiteId,
  });
}

export function useCreateSuite() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; description: string; category?: string; difficulty?: string; tasks: { name: string; category: string; description?: string }[] }) =>
      api.agentBenchmarks.createSuite(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["benchmark-suites"] }),
  });
}

export function useRunBenchmark() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { suiteId: string; agentId: string; agentVersion: string; agentName?: string }) =>
      api.agentBenchmarks.runBenchmark(data) as Promise<BenchmarkRun>,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["benchmark-suites"] });
      queryClient.invalidateQueries({ queryKey: ["benchmark-leaderboard"] });
    },
  });
}

export function useOneClickBenchmark() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { suiteId: string; agentId: string }) =>
      api.agentBenchmarks.runBenchmark({
        ...data,
        agentVersion: "latest",
      }) as Promise<BenchmarkRun>,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["benchmark-suites"] });
      queryClient.invalidateQueries({ queryKey: ["benchmark-leaderboard"] });
    },
  });
}

export function useBenchmarkRunProgress(runId: string | null) {
  return useQuery({
    queryKey: ["benchmark-run-progress", runId],
    queryFn: () =>
      api.agentBenchmarks.getSuite(runId!) as Promise<BenchmarkRun>,
    enabled: !!runId,
    refetchInterval: (query) => {
      const data = query.state.data as BenchmarkRun | undefined;
      if (data?.status === "completed" || data?.status === "failed") return false;
      return 2000;
    },
  });
}

export function useLeaderboard(suiteId: string | null) {
  return useQuery({
    queryKey: ["benchmark-leaderboard", suiteId],
    queryFn: () =>
      api.agentBenchmarks.getLeaderboard(suiteId!) as Promise<BenchmarkLeaderboard>,
    enabled: !!suiteId,
  });
}

export function useCompareBenchmarkRuns(runIds: string[]) {
  return useQuery({
    queryKey: ["benchmark-comparison", ...runIds],
    queryFn: async () => {
      const runs = await Promise.all(
        runIds.map((id) => api.agentBenchmarks.getSuite(id) as Promise<BenchmarkRun>)
      );
      return { runs } as BenchmarkComparison;
    },
    enabled: runIds.length >= 2,
  });
}

export function useGenerateBenchmarkReport() {
  return useMutation({
    mutationFn: (data: { suiteId: string; runIds: string[] }) =>
      api.agentBenchmarks.runBenchmark(data) as Promise<BenchmarkReport>,
  });
}
