import { api } from '@/lib/api';

export type ShareResourceType = 'trace' | 'replay_plan';

export interface ShareLinkCreated {
  id: string;
  resourceType: ShareResourceType;
  resourceId: string;
  redactionVersion: number;
  expiresAt: string;
  createdBy: string;
  createdAt: string;
  token: string;
  url: string;
}

export interface SharedResourceView {
  resourceType: ShareResourceType;
  expiresAt: string;
  trace?: {
    traceId: string;
    name: string;
    startTime: string;
    endTime?: string;
    durationMs: number;
    summary: {
      totalEvents: number;
      llmCalls: number;
      toolCalls: number;
      fileOperations: number;
      terminalCommands: number;
      checkpoints: number;
      errors: number;
      totalTokens: number;
      totalCost: number;
    };
    events: Array<{
      type: string;
      timestamp: string;
      durationMs?: number;
      title: string;
      status: string;
      model?: string;
      tokens?: number;
    }>;
  };
  replayPlan?: {
    planId: string;
    traceId: string;
    status: string;
    capabilities: {
      canInspectTimeline: boolean;
      canReplayRecordedGeneration: boolean;
      canExecuteInSandbox: boolean;
      generationCount: number;
      unsupportedReasons: string[];
      safetyNotice: string;
    };
    comparison?: {
      originalGenerationCount: number;
      replayGenerationCount: number;
      originalTokens: number;
      replayTokens: number;
      originalCost: number;
      replayProviderCost: number;
      equivalent: boolean;
      verdict: string;
      notes: string[];
    };
  };
}

export const shareLinksApi = {
  createTrace: (traceId: string, expiresInSeconds = 7 * 24 * 60 * 60) =>
    api.post<ShareLinkCreated>(`/api/public/traces/${encodeURIComponent(traceId)}/share-links`, {
      expiresInSeconds,
    }),
  createReplayPlan: (planId: string, expiresInSeconds = 7 * 24 * 60 * 60) =>
    api.post<ShareLinkCreated>(
      `/api/public/replay-plans/${encodeURIComponent(planId)}/share-links`,
      { expiresInSeconds }
    ),
};
