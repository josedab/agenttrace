import { PageHeader } from '@/components/layout/page-header';
import { OutcomeDashboard } from '@/components/outcomes/outcome-dashboard';

export const metadata = {
  title: 'Agent Outcomes | AgentTrace',
  description: 'Project-scoped trace, git, CI, pull request, and cost outcome analytics',
};

export default function OutcomeAnalyticsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Agent outcomes"
        description="Follow agent runs through commits and CI without filling gaps with synthetic metrics."
      />
      <OutcomeDashboard />
    </div>
  );
}
