"use client";

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from "@tanstack/react-query";
import { api, Checkpoint } from "@/lib/api";

export interface CheckpointFilters {
  traceId?: string;
  type?: "auto" | "manual";
  limit?: number;
}

export function useCheckpoints(projectId: string, filters?: CheckpointFilters) {
  return useInfiniteQuery({
    queryKey: ["checkpoints", projectId, filters],
    queryFn: ({ pageParam }) =>
      api.checkpoints.list(projectId, { ...filters, cursor: pageParam }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    enabled: !!projectId,
  });
}

export function useCheckpoint(projectId: string, checkpointId: string) {
  return useQuery({
    queryKey: ["checkpoint", projectId, checkpointId],
    queryFn: () => api.checkpoints.get(projectId, checkpointId),
    enabled: !!projectId && !!checkpointId,
  });
}

export function useCheckpointsByTrace(projectId: string, traceId: string) {
  return useQuery({
    queryKey: ["checkpoints", projectId, "trace", traceId],
    queryFn: () => api.checkpoints.listByTrace(projectId, traceId),
    enabled: !!projectId && !!traceId,
  });
}

export function useCreateCheckpoint(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateCheckpointInput) =>
      api.checkpoints.create(projectId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["checkpoints", projectId] });
      if (variables.traceId) {
        queryClient.invalidateQueries({ queryKey: ["checkpoints", projectId, "trace", variables.traceId] });
      }
    },
  });
}

export function useRestoreCheckpoint(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (checkpointId: string) =>
      api.checkpoints.restore(projectId, checkpointId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["checkpoints", projectId] });
    },
  });
}

// Git Links hooks
export function useGitLinks(projectId: string, filters?: { traceId?: string; limit?: number }) {
  return useInfiniteQuery({
    queryKey: ["git-links", projectId, filters],
    queryFn: ({ pageParam }) =>
      api.gitLinks.list(projectId, { ...filters, cursor: pageParam }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    enabled: !!projectId,
  });
}

export function useGitTimeline(projectId: string, traceId: string) {
  return useQuery({
    queryKey: ["git-timeline", projectId, traceId],
    queryFn: () => api.gitLinks.timeline(projectId, traceId),
    enabled: !!projectId && !!traceId,
  });
}

// File Operations hooks
export function useFileOperations(projectId: string, traceId: string) {
  return useQuery({
    queryKey: ["file-operations", projectId, traceId],
    queryFn: () => api.fileOperations.list(projectId, traceId),
    enabled: !!projectId && !!traceId,
  });
}

// Terminal Commands hooks
export function useTerminalCommands(projectId: string, traceId: string) {
  return useQuery({
    queryKey: ["terminal-commands", projectId, traceId],
    queryFn: () => api.terminalCommands.list(projectId, traceId),
    enabled: !!projectId && !!traceId,
  });
}

// CI Runs hooks
export function useCIRuns(projectId: string) {
  return useInfiniteQuery({
    queryKey: ["ci-runs", projectId],
    queryFn: ({ pageParam }) =>
      api.ciRuns.list(projectId, { cursor: pageParam }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    enabled: !!projectId,
  });
}

export function useCIRun(projectId: string, runId: string) {
  return useQuery({
    queryKey: ["ci-run", projectId, runId],
    queryFn: () => api.ciRuns.get(projectId, runId),
    enabled: !!projectId && !!runId,
  });
}

// Input types
export interface CreateCheckpointInput {
  traceId: string;
  type?: "auto" | "manual";
  description?: string;
  files: {
    path: string;
    content: string;
    hash: string;
  }[];
  metadata?: Record<string, unknown>;
}
