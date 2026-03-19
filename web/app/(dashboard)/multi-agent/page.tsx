import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { TopologyDashboard } from "@/components/agent-graph/topology-dashboard";

export const metadata = {
  title: "Multi-Agent Topology Dashboard | AgentTrace",
  description: "Visualize and analyze multi-agent collaboration patterns, message flows, and delegation chains",
};

export default function MultiAgentPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Multi-Agent Topology Dashboard"
        description="Visualize agent collaboration, message flows, delegation chains, and identify bottlenecks"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <TopologyDashboard />
      </Suspense>
    </div>
  );
}
