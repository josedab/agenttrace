'use client';

import { useMutation } from '@tanstack/react-query';

import { shareLinksApi, type ShareResourceType } from '@/lib/share-links';

export function useCreateShareLink(resourceType: ShareResourceType, resourceId: string) {
  return useMutation({
    mutationFn: (expiresInSeconds?: number) =>
      resourceType === 'trace'
        ? shareLinksApi.createTrace(resourceId, expiresInSeconds)
        : shareLinksApi.createReplayPlan(resourceId, expiresInSeconds),
  });
}
