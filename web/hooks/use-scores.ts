"use client";

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface ScoreFilters {
  scoreName?: string;
  source?: string;
  minScore?: number;
  maxScore?: number;
  traceId?: string;
  observationId?: string;
  evaluatorId?: string;
}

export function useScores(filters: ScoreFilters = {}) {
  return useInfiniteQuery({
    queryKey: ["scores", filters],
    queryFn: async ({ pageParam }) => {
      const response = await api.scores.list({
        cursor: pageParam,
        limit: 50,
        ...filters,
      });
      return response;
    },
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor,
  });
}

export function useScore(scoreId: string) {
  return useQuery({
    queryKey: ["score", scoreId],
    queryFn: () => api.scores.get(scoreId),
    enabled: !!scoreId,
  });
}

export function useScoresByTrace(traceId: string) {
  return useQuery({
    queryKey: ["trace-scores", traceId],
    queryFn: () => api.scores.listByTrace(traceId),
    enabled: !!traceId,
  });
}

export function useScoresByObservation(observationId: string) {
  return useQuery({
    queryKey: ["observation-scores", observationId],
    queryFn: () => api.scores.listByObservation(observationId),
    enabled: !!observationId,
  });
}

export function useScoreStats(dateRange: string = "7d") {
  return useQuery({
    queryKey: ["score-stats", dateRange],
    queryFn: () => api.scores.stats({ dateRange }),
  });
}

export function useScoreDistribution(scoreName: string, dateRange: string = "7d") {
  return useQuery({
    queryKey: ["score-distribution", scoreName, dateRange],
    queryFn: () => api.scores.distribution({ scoreName, dateRange }),
    enabled: !!scoreName,
  });
}

export function useCreateScore() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      traceId: string;
      observationId?: string;
      name: string;
      value: number | boolean | string;
      dataType?: "NUMERIC" | "BOOLEAN" | "CATEGORICAL";
      comment?: string;
    }) => api.scores.create(data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["scores"] });
      queryClient.invalidateQueries({ queryKey: ["trace-scores", variables.traceId] });
      if (variables.observationId) {
        queryClient.invalidateQueries({
          queryKey: ["observation-scores", variables.observationId],
        });
      }
    },
  });
}

export function useUpdateScore() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      scoreId,
      data,
    }: {
      scoreId: string;
      data: {
        value?: number | boolean | string;
        comment?: string;
      };
    }) => api.scores.update(scoreId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["score", variables.scoreId] });
      queryClient.invalidateQueries({ queryKey: ["scores"] });
    },
  });
}

export function useDeleteScore() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (scoreId: string) => api.scores.delete(scoreId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scores"] });
    },
  });
}

// Score names for autocomplete
export function useScoreNames() {
  return useQuery({
    queryKey: ["score-names"],
    queryFn: () => api.scores.names(),
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
  });
}
