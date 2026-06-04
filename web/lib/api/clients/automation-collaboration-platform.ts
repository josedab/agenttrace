// automationCollaborationPlatform API client.
import { fetchWithAuth, unsupportedApiFeature } from '../transport';
import type {
  CreateOTelBridgeDestinationInput,
  OTelBridgeConfiguration,
  OTelBridgeDestination,
  OTelBridgeStats,
  OTelImportInput,
  OTelImportResult,
} from '../contracts';

export const automationCollaborationPlatformApi = {
  orchestration: {
    listSessions: () => fetchWithAuth<unknown>('/api/public/orchestration/sessions'),
    createSession: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/orchestration/sessions', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getSession: (id: string) => fetchWithAuth<unknown>(`/api/public/orchestration/sessions/${id}`),
    executeCommand: (id: string, cmd: unknown) =>
      fetchWithAuth<unknown>(`/api/public/orchestration/sessions/${id}/command`, {
        method: 'POST',
        body: JSON.stringify(cmd),
      }),
    addBreakpoint: (id: string, bp: unknown) =>
      fetchWithAuth<unknown>(`/api/public/orchestration/sessions/${id}/breakpoints`, {
        method: 'POST',
        body: JSON.stringify(bp),
      }),
  },
  // Multi-Agent
  multiAgent: {
    listSessions: () => fetchWithAuth<unknown>('/api/public/multi-agent/sessions'),
    analyze: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/multi-agent/analyze', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getSession: (id: string) => fetchWithAuth<unknown>(`/api/public/multi-agent/sessions/${id}`),
    getTopologyGraph: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/multi-agent/sessions/${id}/topology`),
    getTopologyAnalytics: (sessionId: string) =>
      fetchWithAuth<unknown>(`/api/public/multi-agent-graph/sessions/${sessionId}/analytics`),
    getDelegationChains: (sessionId: string) =>
      fetchWithAuth<unknown>(`/api/public/multi-agent-graph/sessions/${sessionId}/delegations`),
  },
  agentVersions: {
    list: (agentName?: string) =>
      fetchWithAuth<unknown>(`/api/public/agent-versions${agentName ? `?agentName=${agentName}` : ''}`),
    create: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/agent-versions', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getActive: (agentName?: string) =>
      fetchWithAuth<unknown>(
        `/api/public/agent-versions/active${agentName ? `?agentName=${agentName}` : ''}`
      ),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/agent-versions/${id}`),
    rollback: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/agent-versions/${id}/rollback`, { method: 'POST' }),
    diff: (data: { versionIdA: string; versionIdB: string }) =>
      fetchWithAuth<unknown>('/api/public/agent-versions/diff', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  agentBuilder: {
    generate: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/agent-builder/generate', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    list: () => fetchWithAuth<unknown>('/api/public/agent-builder/blueprints'),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/agent-builder/blueprints/${id}`),
    deploy: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/agent-builder/blueprints/${id}/deploy`, { method: 'POST' }),
  },
  fleet: {
    getDashboard: () => fetchWithAuth<unknown>('/api/public/fleet/dashboard'),
    listAgents: () => fetchWithAuth<unknown>('/api/public/fleet/agents'),
    listPolicies: () => fetchWithAuth<unknown>('/api/public/fleet/policies'),
    createPolicy: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/fleet/policies', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    bulkUpdate: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/fleet/bulk-update', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getScaling: () => fetchWithAuth<unknown>('/api/public/fleet/scaling'),
  },
  // Collaboration Hub
  collabHub: {
    createQueue: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/collab/queues', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listQueues: () => fetchWithAuth<unknown>('/api/public/collab/queues'),
    assignReview: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/collab/reviews', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    completeReview: (id: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/collab/reviews/${id}/complete`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    createStandard: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/collab/standards', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listStandards: () => fetchWithAuth<unknown>('/api/public/collab/standards'),
    getActivityFeed: () => fetchWithAuth<unknown>('/api/public/collab/activity'),
  },
  // Collaboration Patterns
  collabPatterns: {
    list: () => fetchWithAuth<unknown>('/api/public/collab-patterns'),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/collab-patterns/${id}`),
    deploy: (id: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/collab-patterns/${id}/deploy`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getDeployments: () => fetchWithAuth<unknown>('/api/public/collab-patterns/deployments'),
    getAnalytics: (id: string) => fetchWithAuth<unknown>(`/api/public/collab-patterns/${id}/analytics`),
  },
  // Observability Copilot
  copilot: {
    ask: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/copilot/ask', { method: 'POST', body: JSON.stringify(data) }),
    getSuggestions: () => fetchWithAuth<unknown>('/api/public/copilot/suggestions'),
    getInsights: () => fetchWithAuth<unknown>('/api/public/copilot/insights'),
  },
  // Trace Reviews (v12)
  traceReviews: {
    list: () => fetchWithAuth<unknown>('/api/public/reviews'),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/reviews/${id}`),
    create: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/reviews', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    approve: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/reviews/${id}/approve`, { method: 'POST' }),
    listComments: (reviewId: string) =>
      fetchWithAuth<unknown>(`/api/public/reviews/${reviewId}/comments`),
    addComment: (reviewId: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/reviews/${reviewId}/comments`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listQueues: () => fetchWithAuth<unknown>('/api/public/review-queues'),
    createQueue: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/review-queues', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listNotificationIntegrations: () => fetchWithAuth<unknown>('/api/public/notification-integrations'),
    addNotificationIntegration: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/notification-integrations', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Federated Learning
  federatedLearning: {
    listRings: () => fetchWithAuth<unknown>('/api/public/federated/rings'),
    joinRing: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/federated/rings/join', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getInsights: (ringId: string) =>
      fetchWithAuth<unknown>(`/api/public/federated/rings/${ringId}/insights`),
    getConfig: () => fetchWithAuth<unknown>('/api/public/federated/config'),
    updateConfig: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/federated/config', {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
  },
  // Federated Aggregation
  federatedAggregation: {
    getDashboard: () => fetchWithAuth<unknown>('/api/public/federated-aggregation/dashboard'),
    registerInstance: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/federated-aggregation/instances', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listInstances: () => fetchWithAuth<unknown>('/api/public/federated-aggregation/instances'),
    submitMetrics: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/federated-aggregation/metrics', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getBenchmarks: (metricType?: string) =>
      fetchWithAuth<unknown>(
        `/api/public/federated-aggregation/benchmarks${metricType ? `?metricType=${encodeURIComponent(metricType)}` : ''}`
      ),
    getInsights: () => fetchWithAuth<unknown>('/api/public/federated-aggregation/insights'),
  },
  // Federated Analytics (v12)
  federatedAnalytics: {
    getDashboard: () => fetchWithAuth<unknown>('/api/public/federated/dashboard'),
    query: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/federated/query', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Federated Mesh
  federatedMesh: {
    submitAnonymizedBenchmark: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/federated-aggregation/anonymized-benchmark', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getIndustryBaselines: () => fetchWithAuth<unknown>('/api/public/federated-aggregation/baselines'),
    getMeshStatus: () => fetchWithAuth<unknown>('/api/public/federated-aggregation/mesh-status'),
  },
  // Federation
  federation: {
    listPeers: () => fetchWithAuth<{ peers: unknown[]; count: number }>('/api/public/federation/peers'),
    addPeer: (data: { name: string; url: string; apiKey?: string }) =>
      fetchWithAuth<unknown>('/api/public/federation/peers', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    removePeer: (peerId: string) =>
      fetchWithAuth<unknown>(`/api/public/federation/peers/${peerId}`, { method: 'DELETE' }),
    query: (data: { query: string; peerIds?: string[] }) =>
      fetchWithAuth<unknown>('/api/public/federation/query', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listDestinations: () =>
      fetchWithAuth<{ destinations: unknown[]; count: number }>('/api/public/federation/destinations'),
    createDestination: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/federation/destinations', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Cross-Org Benchmarks
  crossOrg: {
    submit: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/cross-org/submit', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getReport: () => fetchWithAuth<unknown>('/api/public/cross-org/report'),
    getIndustry: (cat: string) => fetchWithAuth<unknown>(`/api/public/cross-org/industry/${cat}`),
  },
  // Workflow Simulator
  workflowSimulator: {
    list: () => fetchWithAuth<unknown>('/api/public/workflows'),
    create: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/workflows', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/workflows/${id}`),
    update: (id: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/workflows/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    delete: (id: string) => fetchWithAuth<unknown>(`/api/public/workflows/${id}`, { method: 'DELETE' }),
    validate: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/workflows/validate', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    simulate: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/workflows/simulate', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getSimulation: (id: string) => fetchWithAuth<unknown>(`/api/public/workflows/simulations/${id}`),
    listSimulations: (workflowId: string) =>
      fetchWithAuth<unknown>(`/api/public/workflows/${workflowId}/simulations`),
  },
  // Webhook Rules
  webhookRules: {
    list: () => fetchWithAuth<unknown>('/api/public/webhook-rules'),
    create: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/webhook-rules', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    delete: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/webhook-rules/${id}`, { method: 'DELETE' }),
    getTemplates: () => fetchWithAuth<unknown>('/api/public/webhook-rules/templates'),
    test: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/webhook-rules/${id}/test`, { method: 'POST' }),
  },
  plugins: {
    list: () => fetchWithAuth<unknown>('/api/public/plugins'),
    install: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/plugins', { method: 'POST', body: JSON.stringify(data) }),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/plugins/${id}`),
    activate: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/plugins/${id}/activate`, { method: 'POST' }),
    disable: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/plugins/${id}/disable`, { method: 'POST' }),
    execute: (id: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/plugins/${id}/execute`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    uninstall: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/plugins/${id}`, { method: 'DELETE' }),
  },
  // Agent Protocol Adapters
  adapters: {
    list: () => fetchWithAuth<unknown>('/api/public/adapters'),
    get: (id: string) => fetchWithAuth<unknown>(`/api/public/adapters/${id}`),
    register: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/adapters', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: unknown) =>
      fetchWithAuth<unknown>(`/api/public/adapters/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    delete: (id: string) => fetchWithAuth<unknown>(`/api/public/adapters/${id}`, { method: 'DELETE' }),
    test: (id: string) => fetchWithAuth<unknown>(`/api/public/adapters/${id}/test`, { method: 'POST' }),
    templates: () => fetchWithAuth<unknown>('/api/public/adapters/templates'),
    install: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/adapters', { method: 'POST', body: JSON.stringify(data) }),
    ingestEvent: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/adapters/events', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  mobile: {
    registerDevice: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/mobile/devices', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getDashboard: () => fetchWithAuth<unknown>('/api/public/mobile/dashboard'),
    listNotifications: () => fetchWithAuth<unknown>('/api/public/mobile/notifications'),
  },
  otelBridge: {
    getConfig: () => fetchWithAuth<OTelBridgeConfiguration>('/api/public/otel-bridge/config'),
    updateConfig: (data: Partial<OTelBridgeConfiguration>) =>
      fetchWithAuth<OTelBridgeConfiguration>('/api/public/otel-bridge/config', {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    listDestinations: () =>
      fetchWithAuth<OTelBridgeDestination[]>('/api/public/otel-bridge/destinations'),
    addDestination: (data: CreateOTelBridgeDestinationInput) =>
      fetchWithAuth<OTelBridgeDestination>('/api/public/otel-bridge/destinations', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    updateDestination: (_id: string, _data: Partial<OTelBridgeDestination>) =>
      unsupportedApiFeature<OTelBridgeDestination>('OpenTelemetry destination updates'),
    importSpans: (data: OTelImportInput) =>
      fetchWithAuth<OTelImportResult>('/api/public/otel-bridge/import', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getStats: () => fetchWithAuth<OTelBridgeStats>('/api/public/otel-bridge/stats'),
  },
  // OpenTelemetry Compatibility
  otelCompat: {
    createDestination: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/otel/destinations', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    listDestinations: () => fetchWithAuth<unknown>('/api/public/otel/destinations'),
    deleteDestination: (id: string) =>
      fetchWithAuth<unknown>(`/api/public/otel/destinations/${id}`, { method: 'DELETE' }),
    getMappings: () => fetchWithAuth<unknown>('/api/public/otel/mappings'),
    getDashboard: () => fetchWithAuth<unknown>('/api/public/otel/dashboard'),
    generateCollectorConfig: () =>
      fetchWithAuth<unknown>('/api/public/otel/collector-config', { method: 'POST' }),
  },
  // Skill Profiles
  skillProfiles: {
    list: () => fetchWithAuth<unknown>('/api/public/skill-profiles'),
    get: (agentName: string) => fetchWithAuth<unknown>(`/api/public/skill-profiles/${agentName}`),
    compare: (agents: string[]) =>
      fetchWithAuth<unknown>(`/api/public/skill-profiles/compare?agents=${agents.join(',')}`),
  },
  // Sandbox
  sandbox: {
    listPending: () => unsupportedApiFeature<never>('Sandbox reviews'),
    submitReview: (_data: unknown) => unsupportedApiFeature<never>('Sandbox reviews'),
    getReview: (_id: string) => unsupportedApiFeature<never>('Sandbox reviews'),
    decide: (_id: string, _decision: unknown) => unsupportedApiFeature<never>('Sandbox reviews'),
    listPolicies: () => unsupportedApiFeature<never>('Sandbox policies'),
    createPolicy: (_data: unknown) => unsupportedApiFeature<never>('Sandbox policies'),
    getStats: () => unsupportedApiFeature<never>('Sandbox reviews'),
  },
  // Cloud Onboarding
  cloudOnboarding: {
    get: () => fetchWithAuth<unknown>('/api/public/onboarding'),
    completeStep: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/onboarding/step', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    generateQuickstart: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/onboarding/quickstart', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    getUsage: () => fetchWithAuth<unknown>('/api/public/onboarding/usage'),
    checkQuota: (data: unknown) =>
      fetchWithAuth<unknown>('/api/public/onboarding/quota-check', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },
  // Cloud Sandbox
  cloudSandbox: {
    create: (data?: unknown) =>
      fetchWithAuth<unknown>('/api/public/cloud-sandbox', {
        method: 'POST',
        body: JSON.stringify(data || { preloadData: true }),
      }),
    get: (sessionId: string) => fetchWithAuth<unknown>(`/api/public/cloud-sandbox/${sessionId}`),
    extend: (sessionId: string) =>
      fetchWithAuth<unknown>(`/api/public/cloud-sandbox/${sessionId}/extend`, { method: 'POST' }),
  },
};
