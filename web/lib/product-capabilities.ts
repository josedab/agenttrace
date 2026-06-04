export type ProductCapabilityId =
  | 'trace-replay'
  | 'eval-hub'
  | 'prompts'
  | 'cost-center'
  | 'collaboration';

export type ProductCapabilityStatus = 'available' | 'preview' | 'hidden';

export interface ProductCapability {
  id: ProductCapabilityId;
  name: string;
  shortName: string;
  description: string;
  canonicalPath: string;
  activePaths: readonly string[];
  icon: 'trace' | 'eval' | 'prompt' | 'cost' | 'collaboration';
  status: ProductCapabilityStatus;
}

export interface CompatibilityRedirect {
  source: string;
  destination: string;
  permanent: false;
}

export const productCapabilities: readonly ProductCapability[] = [
  {
    id: 'trace-replay',
    name: 'Trace & Replay Explorer',
    shortName: 'Trace & Replay',
    description:
      'Inspect agent runs, checkpoints, file operations, terminal activity, and replay plans.',
    canonicalPath: '/traces',
    activePaths: ['/traces', '/replay'],
    icon: 'trace',
    status: 'available',
  },
  {
    id: 'eval-hub',
    name: 'Eval Hub',
    shortName: 'Eval Hub',
    description:
      'Build evaluators, curate datasets, run experiments, and share versioned eval assets.',
    canonicalPath: '/evals',
    activePaths: ['/evals', '/datasets'],
    icon: 'eval',
    status: 'available',
  },
  {
    id: 'prompts',
    name: 'Prompts Workspace',
    shortName: 'Prompts',
    description: 'Version, test, and promote prompts without splitting work across separate labs.',
    canonicalPath: '/prompts',
    activePaths: ['/prompts'],
    icon: 'prompt',
    status: 'available',
  },
  {
    id: 'cost-center',
    name: 'Cost Center',
    shortName: 'Cost Center',
    description: 'Understand spend, attribution, budgets, and cost per successful outcome.',
    canonicalPath: '/analytics/cost',
    activePaths: ['/analytics/cost'],
    icon: 'cost',
    status: 'available',
  },
  {
    id: 'collaboration',
    name: 'Collaboration',
    shortName: 'Collaboration',
    description:
      'Review project outcomes and share team-ready reports from real trace and CI data.',
    canonicalPath: '/analytics/outcomes',
    activePaths: ['/analytics/outcomes'],
    icon: 'collaboration',
    status: 'available',
  },
] as const;

export const primaryNavigationCapabilities = productCapabilities.filter(
  (capability) => capability.status === 'available'
);

export const compatibilityRedirects: readonly CompatibilityRedirect[] = [
  { source: '/eval-marketplace', destination: '/evals?view=library', permanent: false },
  { source: '/eval-playground', destination: '/evals?view=playground', permanent: false },
  { source: '/datasets', destination: '/evals?view=datasets', permanent: false },
  { source: '/test-suites', destination: '/evals?view=runs', permanent: false },
  { source: '/agent-benchmarks', destination: '/evals?view=benchmarks', permanent: false },
  { source: '/prompt-lab', destination: '/prompts', permanent: false },
  { source: '/prompt-optimization', destination: '/prompts?view=optimization', permanent: false },
  { source: '/prompt-ci', destination: '/prompts?view=ci', permanent: false },
  { source: '/cost-optimizer', destination: '/analytics/cost?view=optimizer', permanent: false },
  {
    source: '/cost-attribution',
    destination: '/analytics/cost?view=attribution',
    permanent: false,
  },
  { source: '/cost-guardrails', destination: '/analytics/cost?view=guardrails', permanent: false },
  { source: '/cost-alerts', destination: '/analytics/cost?view=alerts', permanent: false },
  { source: '/cost-forecast', destination: '/analytics/cost?view=forecast', permanent: false },
  { source: '/team', destination: '/analytics/outcomes?view=team', permanent: false },
  { source: '/collab', destination: '/analytics/outcomes?view=collaboration', permanent: false },
  {
    source: '/collab-patterns',
    destination: '/analytics/outcomes?view=collaboration',
    permanent: false,
  },
  { source: '/cross-org', destination: '/analytics/outcomes?view=collaboration', permanent: false },
  { source: '/code-impact', destination: '/analytics/outcomes?view=code', permanent: false },
  { source: '/session-journeys', destination: '/analytics/outcomes?view=runs', permanent: false },
  { source: '/trace-diff', destination: '/replay?view=compare', permanent: false },
] as const;

export function getCapabilityForPath(pathname: string): ProductCapability | undefined {
  return productCapabilities.find((capability) => isCapabilityPath(capability, pathname));
}

export function isCapabilityPath(capability: ProductCapability, pathname: string): boolean {
  return capability.activePaths.some(
    (path) => pathname === path || pathname.startsWith(`${path}/`)
  );
}
