import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    {
      type: 'doc',
      id: 'introduction',
      label: 'Introduction',
    },
    {
      type: 'category',
      label: 'Getting Started',
      collapsed: false,
      items: [
        'getting-started/quickstart',
        'getting-started/installation',
        'getting-started/first-trace',
        'getting-started/concepts',
      ],
    },
    {
      type: 'category',
      label: 'Core Features',
      items: [
        'features/tracing',
        'features/observations',
        'features/sessions',
        'features/scores',
        'features/cost-tracking',
        'features/latency-analysis',
        'features/anomaly-detection',
        'features/prompt-library',
      ],
    },
    {
      type: 'category',
      label: 'Prompt Management',
      items: [
        'prompts/overview',
        'prompts/versioning',
        'prompts/labels',
        'prompts/playground',
        'prompts/variables',
      ],
    },
    {
      type: 'category',
      label: 'Datasets & Experiments',
      items: [
        'datasets/overview',
        'datasets/creating-datasets',
        'datasets/running-experiments',
        'datasets/comparing-results',
      ],
    },
    {
      type: 'category',
      label: 'Evaluation',
      items: [
        'evaluation/overview',
        'evaluation/llm-as-judge',
        'evaluation/human-annotation',
        'evaluation/custom-evaluators',
      ],
    },
    {
      type: 'category',
      label: 'Agent Features',
      items: [
        'agent-features/checkpoints',
        'agent-features/git-linking',
        'agent-features/file-operations',
        'agent-features/terminal-commands',
      ],
    },
    {
      type: 'category',
      label: 'Integrations',
      items: [
        'integrations/opentelemetry',
        'integrations/github-actions',
        'integrations/gitlab-ci',
        'integrations/vscode',
        'integrations/jetbrains',
        'integrations/jupyter',
        'integrations/github-app',
        'integrations/notifications',
        'integrations/grafana',
      ],
    },
    {
      type: 'category',
      label: 'SDKs',
      link: {
        type: 'doc',
        id: 'sdks/index',
      },
      items: [
        'sdks/python',
        'sdks/typescript',
        'sdks/go',
        'sdks/cli',
      ],
    },
    {
      type: 'category',
      label: 'API Reference',
      link: {
        type: 'doc',
        id: 'api-reference/index',
      },
      items: [
        'api-reference/authentication',
        'api-reference/traces',
        'api-reference/observations',
        'api-reference/sessions',
        'api-reference/scores',
        'api-reference/prompts',
        'api-reference/datasets',
        'api-reference/evaluators',
        'api-reference/checkpoints',
        'api-reference/git-links',
        'api-reference/ci-runs',
        'api-reference/webhooks',
        'api-reference/graphql',
      ],
    },
    {
      type: 'category',
      label: 'Self-Hosting',
      items: [
        'self-hosting/overview',
        'self-hosting/docker-compose',
        'self-hosting/kubernetes',
        'self-hosting/configuration',
        'self-hosting/scaling',
        'self-hosting/backup',
      ],
    },
    {
      type: 'category',
      label: 'Enterprise',
      items: [
        'enterprise/sso',
        'enterprise/audit-logs',
        'enterprise/rbac',
        'enterprise/compliance',
      ],
    },
    {
      type: 'doc',
      id: 'troubleshooting',
      label: 'Troubleshooting',
    },
  ],
};

export default sidebars;
