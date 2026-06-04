import { api } from '@/lib/api';

export type OutcomeWindow = '24h' | '7d' | '30d' | '90d';

export interface OutcomeOptionalMetric {
  value: number | null;
  available: boolean;
}

export interface OutcomeBreakdown {
  name: string;
  runs: number;
  successfulRuns: number;
  successRate: OutcomeOptionalMetric;
  totalCost: number;
  costPerSuccessfulOutcome: OutcomeOptionalMetric;
}

export interface LinkedOutcome {
  commitSha: string;
  commitMessage?: string;
  branch?: string;
  committedAt: string;
  traceCount: number;
  ciStatus: string | null;
  ciConclusion: string | null;
  ciProviderUrl: string | null;
  prNumber: number | null;
  prTitle: string | null;
}

export interface OutcomeOverview {
  projectId: string;
  period: {
    from: string;
    to: string;
  };
  runs: {
    total: number;
    successful: number;
    failed: number;
    inProgress: number;
    successRate: OutcomeOptionalMetric;
  };
  ci: {
    total: number;
    passed: number;
    failed: number;
    cancelled: number;
    inProgress: number;
    linkedRuns: number;
    passRate: OutcomeOptionalMetric;
  };
  sourceControl: {
    linkedCommits: number;
    linkedTraces: number;
    linkedPullRequests: number;
    regressionSignals: number;
    revertSignals: number;
  };
  cost: {
    totalCost: number;
    costPerSuccessfulOutcome: OutcomeOptionalMetric;
  };
  byAgent: OutcomeBreakdown[];
  byModel: OutcomeBreakdown[];
  recentOutcomes: LinkedOutcome[];
  availability: {
    traceData: boolean;
    gitData: boolean;
    ciData: boolean;
    pullRequestData: boolean;
    agentAttribution: boolean;
    modelAttribution: boolean;
    unavailable: string[];
  };
  generatedAt: string;
}

export interface OutcomeDigest {
  projectId: string;
  period: {
    from: string;
    to: string;
  };
  title: string;
  summary: string;
  highlights: string[];
  attention: string[];
  generatedAt: string;
}

function outcomeSearchParams(window: OutcomeWindow) {
  return new URLSearchParams({ window }).toString();
}

export const outcomesApi = {
  getOverview: (window: OutcomeWindow) =>
    api.get<OutcomeOverview>(`/api/public/outcomes?${outcomeSearchParams(window)}`),
  getDigest: (window: OutcomeWindow) =>
    api.get<OutcomeDigest>(`/api/public/outcomes/digest?${outcomeSearchParams(window)}`),
};
