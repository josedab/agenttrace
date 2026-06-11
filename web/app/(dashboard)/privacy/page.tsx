import { PageHeader } from '@/components/layout/page-header';
import { PrivacyModeDashboard } from '@/components/privacy/privacy-mode-dashboard';

export const metadata = {
  title: 'Privacy Center | AgentTrace',
  description: 'Effective no-egress, redaction, and external capability status',
};

export default function PrivacyPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Privacy center"
        description="See the behavior AgentTrace is actually enforcing for egress and redaction."
      />
      <PrivacyModeDashboard />
    </div>
  );
}
