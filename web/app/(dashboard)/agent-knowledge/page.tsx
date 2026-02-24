import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Knowledge Graph | AgentTrace",
  description: "Visualize agent codebase knowledge as an interactive graph",
};

export default function AgentKnowledgePage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Agent Knowledge Graph"
        description="Explore how agents understand your codebase through an interactive knowledge graph"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AgentKnowledgeDashboardContent />
      </Suspense>
    </div>
  );
}

function AgentKnowledgeDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Knowledge Graph Dashboard</p>
      <p className="text-sm mt-2">Visualize relationships between agents, tools, APIs, and concepts</p>
      <p className="text-sm mt-1">Supports graph evolution tracking, node exploration, and dependency analysis</p>
    </div>
  );
}
