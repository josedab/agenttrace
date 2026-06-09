'use client';

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from '@tanstack/react-query';
import { api, Checkpoint } from '@/lib/api';
import type { CreateCheckpointInput as ApiCreateCheckpointInput } from '@/lib/api';

export interface CheckpointFilters {
  traceId?: string;
  type?: Checkpoint['type'];
  limit?: number;
}

export function useCheckpoints(projectId: string, filters?: CheckpointFilters) {
  return useInfiniteQuery({
    queryKey: ['checkpoints', projectId, filters],
    queryFn: ({ pageParam }) => api.checkpoints.list(projectId, { ...filters, offset: pageParam }),
    initialPageParam: 0,
    getNextPageParam: (lastPage, pages) =>
      lastPage.hasMore
        ? pages.reduce((offset, page) => offset + page.checkpoints.length, 0)
        : undefined,
    enabled: !!projectId,
  });
}

export function useCheckpoint(projectId: string, checkpointId: string) {
  return useQuery({
    queryKey: ['checkpoint', projectId, checkpointId],
    queryFn: () => api.checkpoints.get(projectId, checkpointId),
    enabled: !!projectId && !!checkpointId,
  });
}

export function useCheckpointsByTrace(projectId: string, traceId: string) {
  return useQuery({
    queryKey: ['checkpoints', projectId, 'trace', traceId],
    queryFn: () => api.checkpoints.listByTrace(projectId, traceId),
    enabled: !!projectId && !!traceId,
  });
}

export function useCreateCheckpoint(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateCheckpointInput) => api.checkpoints.create(projectId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['checkpoints', projectId] });
      if (variables.traceId) {
        queryClient.invalidateQueries({
          queryKey: ['checkpoints', projectId, 'trace', variables.traceId],
        });
      }
    },
  });
}

export function useRestoreCheckpoint(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ checkpointId, traceId }: { checkpointId: string; traceId: string }) =>
      api.checkpoints.restore(projectId, checkpointId, traceId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['checkpoints', projectId] });
    },
  });
}

// Git Links hooks
export function useGitLinks(projectId: string, filters?: { traceId?: string; limit?: number }) {
  return useInfiniteQuery({
    queryKey: ['git-links', projectId, filters],
    queryFn: ({ pageParam }) => api.gitLinks.list(projectId, { ...filters, offset: pageParam }),
    initialPageParam: 0,
    getNextPageParam: (lastPage, pages) =>
      lastPage.hasMore
        ? pages.reduce((offset, page) => offset + page.gitLinks.length, 0)
        : undefined,
    enabled: !!projectId,
  });
}

export function useGitTimeline(projectId: string, branch?: string) {
  return useQuery({
    queryKey: ['git-timeline', projectId, branch],
    queryFn: () => api.gitLinks.timeline(projectId, branch),
    enabled: !!projectId,
  });
}

// File Operations hooks
export function useFileOperations(projectId: string, traceId: string) {
  return useQuery({
    queryKey: ['file-operations', projectId, traceId],
    queryFn: () => api.fileOperations.list(projectId, traceId),
    enabled: !!projectId && !!traceId,
  });
}

// Terminal Commands hooks
export function useTerminalCommands(projectId: string, traceId: string) {
  return useQuery({
    queryKey: ['terminal-commands', projectId, traceId],
    queryFn: () => api.terminalCommands.list(projectId, traceId),
    enabled: !!projectId && !!traceId,
  });
}

// CI Runs hooks
export function useCIRuns(projectId: string) {
  return useInfiniteQuery({
    queryKey: ['ci-runs', projectId],
    queryFn: ({ pageParam }) => api.ciRuns.list(projectId, { offset: pageParam }),
    initialPageParam: 0,
    getNextPageParam: (lastPage, pages) =>
      lastPage.hasMore ? pages.reduce((offset, page) => offset + page.ciRuns.length, 0) : undefined,
    enabled: !!projectId,
  });
}

export function useCIRun(projectId: string, runId: string) {
  return useQuery({
    queryKey: ['ci-run', projectId, runId],
    queryFn: () => api.ciRuns.get(projectId, runId),
    enabled: !!projectId && !!runId,
  });
}

// Input types
export type CreateCheckpointInput = ApiCreateCheckpointInput;
