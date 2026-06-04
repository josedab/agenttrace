// Compatibility facade for the AgentTrace web API client.
//
// The implementation is split into cohesive modules under `lib/api/`:
//   - transport:  transport, auth/project state, error + JSON helpers
//   - contracts:  public DTOs/types exposed to consumers
//   - normalizers: response DTOs and payload -> contract mappers
//   - clients/*:  domain clients grouped into coherent areas
//
// This file re-exports the public surface and composes the `api` object so
// that existing `@/lib/api` imports keep working unchanged.

export * from './api/contracts';
export {
  API_URL,
  ApiError,
  setApiAccessToken,
  getApiAccessToken,
  setApiProjectId,
  getApiProjectId,
  createApiClient,
  createGraphQLClient,
} from './api/transport';

import { coreApi } from './api/clients/core';
import { identityApi } from './api/clients/identity';
import { observabilityApi } from './api/clients/observability';
import { promptsEvalsDataApi } from './api/clients/prompts-evals-data';
import { intelligenceGovernanceApi } from './api/clients/intelligence-governance';
import { automationCollaborationPlatformApi } from './api/clients/automation-collaboration-platform';

/**
 * Aggregated API client. Composed from the per-area domain clients; the object
 * shape (and every method) is identical to the previous single-module export.
 */
export const api = {
  ...coreApi,
  ...identityApi,
  ...observabilityApi,
  ...promptsEvalsDataApi,
  ...intelligenceGovernanceApi,
  ...automationCollaborationPlatformApi,
};
