import { Suspense } from 'react';
import { PageHeader } from '@/components/layout/page-header';
import { ReplayDashboard } from './replay-dashboard';

export const metadata = {
  title: 'Trace Replay Debugger | AgentTrace',
  description: 'Inspect trace timelines and construct safe replay plans',
};

export default function ReplayPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Trace replay debugger"
        description="Inspect recorded execution, select a checkpoint, and compare a safe replay branch without running captured code on the API host."
      />
      <Suspense fallback={<div className="h-96 animate-pulse rounded-lg bg-muted" />}>
        <ReplayDashboard />
      </Suspense>
    </div>
  );
}
