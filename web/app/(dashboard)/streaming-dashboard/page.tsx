import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { StreamingDashboard } from "@/components/streaming-dashboard/streaming-dashboard";

export const metadata = {
  title: "Live Dashboard | AgentTrace",
  description: "Real-time streaming metrics for active AI agent sessions",
};

export default function StreamingDashboardPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Live Dashboard"
        description="Real-time streaming metrics, active sessions, and cost tracking"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <StreamingDashboard />
      </Suspense>
    </div>
  );
}
