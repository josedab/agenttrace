"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useEvaluators() {
  return useQuery({
    queryKey: ["evaluators"],
    queryFn: () => api.evaluators.list(),
  });
}

export function useEvaluator(evaluatorId: string) {
  return useQuery({
    queryKey: ["evaluator", evaluatorId],
    queryFn: () => api.evaluators.get(evaluatorId),
    enabled: !!evaluatorId,
  });
}

export function useEvaluatorStats(evaluatorId: string) {
  return useQuery({
    queryKey: ["evaluator-stats", evaluatorId],
    queryFn: () => api.evaluators.getStats(evaluatorId),
    enabled: !!evaluatorId,
  });
}

export function useEvaluatorRuns(evaluatorId: string) {
  return useQuery({
    queryKey: ["evaluator-runs", evaluatorId],
    queryFn: () => api.evaluators.listRuns(evaluatorId),
    enabled: !!evaluatorId,
    refetchInterval: (query) => {
      // Keep polling if any run is in progress
      const hasRunning = query.state.data?.some(
        (run: any) => run.status === "RUNNING" || run.status === "PENDING"
      );
      return hasRunning ? 3000 : false;
    },
  });
}

export function useCreateEvaluator() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      name: string;
      description?: string;
      type: "LLM_AS_JUDGE" | "CODE" | "HUMAN";
      scoreName: string;
      template?: string;
      config?: Record<string, any>;
    }) => api.evaluators.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["evaluators"] });
    },
  });
}

export function useUpdateEvaluator() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      evaluatorId,
      data,
    }: {
      evaluatorId: string;
      data: {
        name?: string;
        description?: string;
        scoreName?: string;
        scoreDataType?: "NUMERIC" | "BOOLEAN" | "CATEGORICAL";
        status?: "ACTIVE" | "INACTIVE";
        config?: {
          model?: string;
          prompt?: string;
          code?: string;
        };
      };
    }) => api.evaluators.update(evaluatorId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["evaluator", variables.evaluatorId] });
      queryClient.invalidateQueries({ queryKey: ["evaluators"] });
    },
  });
}

export function useDeleteEvaluator() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (evaluatorId: string) => api.evaluators.delete(evaluatorId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["evaluators"] });
    },
  });
}

export function useRunEvaluator() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (evaluatorId: string) => api.evaluators.run(evaluatorId),
    onSuccess: (_, evaluatorId) => {
      queryClient.invalidateQueries({ queryKey: ["evaluator-runs", evaluatorId] });
      queryClient.invalidateQueries({ queryKey: ["evaluator-stats", evaluatorId] });
    },
  });
}

// Annotation queues
export function useAnnotationQueues() {
  return useQuery({
    queryKey: ["annotation-queues"],
    queryFn: () => api.annotationQueues.list(),
  });
}

export function useAnnotationQueue(queueId: string) {
  return useQuery({
    queryKey: ["annotation-queue", queueId],
    queryFn: () => api.annotationQueues.get(queueId),
    enabled: !!queueId,
  });
}

export function useAnnotationQueueItems(queueId: string) {
  return useQuery({
    queryKey: ["annotation-items", queueId],
    queryFn: () => api.annotationQueues.getItems(queueId),
    enabled: !!queueId,
  });
}

export function useSubmitAnnotation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      queueId,
      itemId,
      data,
    }: {
      queueId: string;
      itemId: string;
      data: {
        score: number | boolean | string;
        comment?: string;
      };
    }) => api.annotationQueues.submitScore(queueId, itemId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["annotation-items", variables.queueId] });
      queryClient.invalidateQueries({ queryKey: ["annotation-queue", variables.queueId] });
    },
  });
}

export function useSkipAnnotationItem() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      queueId,
      itemId,
    }: {
      queueId: string;
      itemId: string;
    }) => api.annotationQueues.skipItem(queueId, itemId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["annotation-items", variables.queueId] });
    },
  });
}
