import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Knowledge Graph | AgentTrace",
  description: "Explore relationships between files, tools, and agents",
};

export default function KnowledgeGraphPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Knowledge Graph"
        description="Explore relationships between files, tools, and agents"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <KnowledgeGraphContent />
      </Suspense>
    </div>
  );
}

function KnowledgeGraphContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Knowledge Graph Explorer</p>
      <p className="text-sm mt-2">Visualize relationships between files, tools, agents, and traces</p>
      <p className="text-sm mt-1">Query the graph to discover hidden dependencies and patterns</p>
    </div>
  );
}
