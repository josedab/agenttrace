import { api } from '@/lib/api';

// Benchmarks are intentionally absent: they have no owning project, so the API
// rejects publishing or forking them as packages.
export type EvalHubAssetKind = 'dataset' | 'evaluator' | 'prompt' | 'experiment';
export type EvalHubVisibility = 'private' | 'organization' | 'public';
export type EvalHubRunStatus = 'ready' | 'running' | 'completed' | 'unsupported' | 'failed';

export interface EvalHubVersion {
  id: string;
  packageId: string;
  version: number;
  sourceResourceId: string;
  manifest: unknown;
  checksum: string;
  versionNote?: string;
  createdBy: string;
  createdAt: string;
}

export interface EvalHubPackage {
  id: string;
  ownerProjectId: string;
  organizationId: string;
  kind: EvalHubAssetKind;
  name: string;
  description?: string;
  visibility: EvalHubVisibility;
  latestVersion: number;
  forkedFromPackageId?: string;
  forkedFromVersion?: number;
  publishedBy: string;
  createdAt: string;
  updatedAt: string;
  version?: EvalHubVersion;
}

export interface EvalHubRun {
  id: string;
  projectId: string;
  packageId: string;
  packageVersion: number;
  status: EvalHubRunStatus;
  datasetRunId?: string;
  experimentId?: string;
  result?: unknown;
  capabilityMessage?: string;
  idempotencyKey?: string;
  createdBy: string;
  startedAt: string;
  completedAt?: string;
}

export interface EvalHubPackageList {
  packages: EvalHubPackage[];
  totalCount: number;
  hasMore: boolean;
}

export interface EvalHubRunList {
  runs: EvalHubRun[];
  totalCount: number;
  hasMore: boolean;
}

export interface PublishEvalHubPackageInput {
  packageId?: string;
  kind: EvalHubAssetKind;
  sourceResourceId: string;
  name?: string;
  description?: string;
  visibility: EvalHubVisibility;
  versionNote?: string;
}

export const evalHubApi = {
  listPackages: (
    params: {
      query?: string;
      kind?: EvalHubAssetKind;
      visibility?: EvalHubVisibility;
      limit?: number;
      offset?: number;
    } = {}
  ) => {
    const searchParams = new URLSearchParams();
    if (params.query) searchParams.set('q', params.query);
    if (params.kind) searchParams.set('kind', params.kind);
    if (params.visibility) searchParams.set('visibility', params.visibility);
    if (params.limit) searchParams.set('limit', String(params.limit));
    if (params.offset) searchParams.set('offset', String(params.offset));
    return api.get<EvalHubPackageList>(`/api/public/eval-hub/packages?${searchParams}`);
  },
  getPackage: (packageId: string) =>
    api.get<EvalHubPackage>(`/api/public/eval-hub/packages/${encodeURIComponent(packageId)}`),
  publish: (input: PublishEvalHubPackageInput) =>
    api.post<EvalHubPackage>('/api/public/eval-hub/packages', input),
  fork: (packageId: string, input: { name?: string; visibility?: EvalHubVisibility } = {}) =>
    api.post<EvalHubPackage>(
      `/api/public/eval-hub/packages/${encodeURIComponent(packageId)}/fork`,
      input
    ),
  run: (
    packageId: string,
    input: {
      name?: string;
      version?: number;
      idempotencyKey?: string;
      variables?: Record<string, string>;
    } = {}
  ) =>
    api.post<EvalHubRun>(
      `/api/public/eval-hub/packages/${encodeURIComponent(packageId)}/runs`,
      input
    ),
  listRuns: (params: { limit?: number; offset?: number } = {}) => {
    const searchParams = new URLSearchParams();
    if (params.limit) searchParams.set('limit', String(params.limit));
    if (params.offset) searchParams.set('offset', String(params.offset));
    return api.get<EvalHubRunList>(`/api/public/eval-hub/runs?${searchParams}`);
  },
};
