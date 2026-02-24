"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface BenchmarkSuite {
  id: string;
  name: string;
  description: string;
  tasks: { id: string; name: string; category: string }[];
  runCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface BenchmarkRun {
  id: string;
  suiteId: string;
  agentId: string;
  agentVersion: string;
  status: "pending" | "running" | "completed" | "failed";
  scores: Record<string, number>;
  overallScore: number;
  startedAt: string;
  completedAt?: string;
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
  runId: string;
  completedAt: string;
}

export function useBenchmarkSuites() {
  return useQuery({
    queryKey: ["benchmark-suites"],
    queryFn: () =>
      api.agentBenchmarks.listSuites() as Promise<BenchmarkSuite[]>,
  });
}

export function useCreateSuite() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; description: string; tasks: { name: string; category: string }[] }) =>
      api.agentBenchmarks.createSuite(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["benchmark-suites"] }),
  });
}

export function useRunBenchmark() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { suiteId: string; agentId: string; agentVersion: string }) =>
      api.agentBenchmarks.runBenchmark(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["benchmark-suites"] }),
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
