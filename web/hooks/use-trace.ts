"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useTrace(traceId: string) {
  return useQuery({
    queryKey: ["trace", traceId],
    queryFn: () => api.traces.get(traceId),
    enabled: !!traceId,
  });
}

export function useTraceObservations(traceId: string) {
  return useQuery({
    queryKey: ["trace-observations", traceId],
    queryFn: () => api.traces.getObservations(traceId),
    enabled: !!traceId,
  });
}

export function useTraceScores(traceId: string) {
  return useQuery({
    queryKey: ["trace-scores", traceId],
    queryFn: () => api.scores.listByTrace(traceId),
    enabled: !!traceId,
  });
}

export function useUpdateTrace() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      traceId,
      data,
    }: {
      traceId: string;
      data: {
        name?: string;
        userId?: string;
        sessionId?: string;
        metadata?: Record<string, any>;
        tags?: string[];
        public?: boolean;
      };
    }) => api.traces.update(traceId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["trace", variables.traceId] });
      queryClient.invalidateQueries({ queryKey: ["traces"] });
    },
  });
}

export function useDeleteTrace() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (traceId: string) => api.traces.delete(traceId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["traces"] });
    },
  });
}

export function useAddTraceScore() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      traceId,
      observationId,
      data,
    }: {
      traceId: string;
      observationId?: string;
      data: {
        name: string;
        value: number | boolean | string;
        dataType?: "NUMERIC" | "BOOLEAN" | "CATEGORICAL";
        comment?: string;
      };
    }) => api.scores.create({ traceId, observationId, ...data }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["trace-scores", variables.traceId],
      });
    },
  });
}

export function useTraceMetadata(traceId: string) {
  const { data: trace } = useTrace(traceId);
  const { data: observations } = useTraceObservations(traceId);

  // Compute derived metadata
  const totalTokens =
    observations?.reduce(
      (sum, obs) => sum + (obs.usage?.totalTokens || 0),
      0
    ) || 0;

  const totalCost =
    observations?.reduce((sum, obs) => sum + (obs.cost?.total || 0), 0) || 0;

  const latency = trace?.endTime && trace?.startTime
    ? new Date(trace.endTime).getTime() - new Date(trace.startTime).getTime()
    : null;

  const modelUsage = observations?.reduce((acc, obs) => {
    if (obs.model) {
      acc[obs.model] = (acc[obs.model] || 0) + 1;
    }
    return acc;
  }, {} as Record<string, number>) || {};

  return {
    trace,
    observations,
    totalTokens,
    totalCost,
    latency,
    modelUsage,
  };
}
