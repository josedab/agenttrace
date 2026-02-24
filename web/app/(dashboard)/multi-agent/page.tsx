import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Multi-Agent Graph | AgentTrace",
  description: "Visualize and analyze multi-agent collaboration patterns",
};

export default function MultiAgentPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Multi-Agent Graph"
        description="Visualize agent collaboration, message flows, and identify bottlenecks"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <MultiAgentDashboardContent />
      </Suspense>
    </div>
  );
}

function MultiAgentDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Multi-Agent Graph Dashboard</p>
      <p className="text-sm mt-2">Explore agent interactions, message dependencies, and collaboration topology</p>
      <p className="text-sm mt-1">Supports bottleneck detection, latency analysis, and communication pattern visualization</p>
    </div>
  );
}
