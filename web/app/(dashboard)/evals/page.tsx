import { PageHeader } from '@/components/layout/page-header';
import { EvalHubDashboard } from '@/components/eval-hub/eval-hub-dashboard';

export const metadata = {
  title: 'Eval Hub | AgentTrace',
  description: 'Project evaluators, datasets, experiments, benchmarks, and versioned packages',
};

export default function EvalHubPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Eval Hub"
        description="Build, version, fork, and run evaluation assets without splitting work across separate labs."
      />
      <EvalHubDashboard />
    </div>
  );
}
