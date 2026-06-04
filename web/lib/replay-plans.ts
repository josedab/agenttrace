import { api } from '@/lib/api';

export type ReplayExecutionMode = 'recorded_generation' | 'sandbox';
export type ReplayPlanStatus =
  | 'planned'
  | 'ready'
  | 'running'
  | 'completed'
  | 'failed'
  | 'unsupported';

export interface ReplayPlanInput {
  checkpointId?: string;
  mode: ReplayExecutionMode;
}

export interface ReplayCapabilityReport {
  canInspectTimeline: boolean;
  canReplayRecordedGeneration: boolean;
  canExecuteInSandbox: boolean;
  hasCheckpoint: boolean;
  hasFileOperations: boolean;
  hasTerminalCommands: boolean;
  generationCount: number;
  unsupportedReasons: string[];
  safetyNotice: string;
}

export interface ReplayPlanComparison {
  originalGenerationCount: number;
  replayGenerationCount: number;
  originalTokens: number;
  replayTokens: number;
  originalCost: number;
  replayProviderCost: number;
  equivalent: boolean;
  verdict: string;
  notes: string[];
}

export interface ReplayGenerationResult {
  eventId: string;
  model?: string;
  outputSha256: string;
  tokens: number;
  referenceCost: number;
}

export interface ReplayPlan {
  id: string;
  projectId: string;
  traceId: string;
  checkpointId?: string;
  status: ReplayPlanStatus;
  request: ReplayPlanInput;
  capabilities: ReplayCapabilityReport;
  result?: {
    startedAt: string;
    completedAt: string;
    generations: ReplayGenerationResult[];
    nonGenerationEventsSkipped: number;
    comparison: ReplayPlanComparison;
  };
  failureReason?: string;
  createdAt: string;
  updatedAt: string;
}

function replayQuery(input: ReplayPlanInput) {
  const params = new URLSearchParams({ mode: input.mode });
  if (input.checkpointId) {
    params.set('checkpointId', input.checkpointId);
  }
  return params.toString();
}

export const replayPlansApi = {
  getCapabilities: (traceId: string, input: ReplayPlanInput) =>
    api.get<ReplayCapabilityReport>(
      `/api/public/traces/${encodeURIComponent(traceId)}/replay-capabilities?${replayQuery(input)}`
    ),
  create: (traceId: string, input: ReplayPlanInput) =>
    api.post<ReplayPlan>(`/api/public/traces/${encodeURIComponent(traceId)}/replay-plans`, input),
  get: (planId: string) =>
    api.get<ReplayPlan>(`/api/public/replay-plans/${encodeURIComponent(planId)}`),
  execute: (planId: string) =>
    api.post<ReplayPlan>(`/api/public/replay-plans/${encodeURIComponent(planId)}/execute`),
  retry: (planId: string) =>
    api.post<ReplayPlan>(`/api/public/replay-plans/${encodeURIComponent(planId)}/retry`),
  getComparison: (planId: string) =>
    api.get<ReplayPlanComparison>(
      `/api/public/replay-plans/${encodeURIComponent(planId)}/comparison`
    ),
};
