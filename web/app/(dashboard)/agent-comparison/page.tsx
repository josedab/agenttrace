import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { AgentComparisonDashboard } from "@/components/agent-comparison/agent-comparison-dashboard";

export const metadata = {
  title: "Agent Comparison | AgentTrace",
  description: "Compare AI coding agents across cost, latency, quality, and efficiency metrics",
};

export default function AgentComparisonPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Agent Comparison"
        description="Compare AI coding agents side-by-side across performance, cost, and quality metrics"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AgentComparisonDashboard />
      </Suspense>
    </div>
  );
}
