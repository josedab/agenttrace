import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Semantic Search | AgentTrace",
  description: "Natural language trace search with clustering and anomaly detection",
};

export default function SemanticSearchPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Semantic Trace Search"
        description="Search traces using natural language, explore clusters, and discover anomaly patterns"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <SemanticSearchDashboardContent />
      </Suspense>
    </div>
  );
}

function SemanticSearchDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Semantic Search Dashboard</p>
      <p className="text-sm mt-2">Use natural language queries to find traces, explore auto-generated clusters</p>
      <p className="text-sm mt-1">Supports vector embeddings, trace clustering, and anomaly pattern detection</p>
    </div>
  );
}
