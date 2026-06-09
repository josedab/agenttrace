'use client';

import { useQuery } from '@tanstack/react-query';

import { outcomesApi, type OutcomeWindow } from '@/lib/outcomes';

export function useOutcomeOverview(window: OutcomeWindow) {
  return useQuery({
    queryKey: ['outcome-overview', window],
    queryFn: () => outcomesApi.getOverview(window),
  });
}

export function useOutcomeDigest(window: OutcomeWindow, enabled: boolean) {
  return useQuery({
    queryKey: ['outcome-digest', window],
    queryFn: () => outcomesApi.getDigest(window),
    enabled,
  });
}
