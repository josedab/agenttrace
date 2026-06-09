'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import {
  evalHubApi,
  type EvalHubAssetKind,
  type EvalHubVisibility,
  type PublishEvalHubPackageInput,
} from '@/lib/eval-hub';

export function useEvalHubPackages(
  params: {
    query?: string;
    kind?: EvalHubAssetKind;
    visibility?: EvalHubVisibility;
  } = {}
) {
  return useQuery({
    queryKey: ['eval-hub-packages', params],
    queryFn: () => evalHubApi.listPackages(params),
  });
}

export function usePublishEvalHubPackage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: PublishEvalHubPackageInput) => evalHubApi.publish(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['eval-hub-packages'] });
    },
  });
}

export function useForkEvalHubPackage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      packageId,
      name,
      visibility,
    }: {
      packageId: string;
      name?: string;
      visibility?: EvalHubVisibility;
    }) => evalHubApi.fork(packageId, { name, visibility }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['eval-hub-packages'] });
    },
  });
}

export function useRunEvalHubPackage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      packageId,
      name,
      idempotencyKey,
    }: {
      packageId: string;
      name?: string;
      idempotencyKey?: string;
    }) => evalHubApi.run(packageId, { name, idempotencyKey }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['eval-hub-runs'] });
    },
  });
}

export function useEvalHubRuns() {
  return useQuery({
    queryKey: ['eval-hub-runs'],
    queryFn: () => evalHubApi.listRuns({ limit: 100 }),
  });
}
