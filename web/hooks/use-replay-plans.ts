'use client';

import { useMutation, useQuery } from '@tanstack/react-query';

import { replayPlansApi, type ReplayExecutionMode, type ReplayPlanInput } from '@/lib/replay-plans';

export function useReplayCapabilities(
  traceId: string,
  checkpointId: string | undefined,
  mode: ReplayExecutionMode
) {
  const input: ReplayPlanInput = { mode, checkpointId };
  return useQuery({
    queryKey: ['replay-capabilities', traceId, checkpointId, mode],
    queryFn: () => replayPlansApi.getCapabilities(traceId, input),
    enabled: Boolean(traceId),
  });
}

export function useCreateReplayPlan(traceId: string) {
  return useMutation({
    mutationFn: (input: ReplayPlanInput) => replayPlansApi.create(traceId, input),
  });
}

export function useExecuteReplayPlan() {
  return useMutation({
    mutationFn: (planId: string) => replayPlansApi.execute(planId),
  });
}

export function useRetryReplayPlan() {
  return useMutation({
    mutationFn: (planId: string) => replayPlansApi.retry(planId),
  });
}
