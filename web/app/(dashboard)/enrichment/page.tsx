import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Enrichment Pipeline | AgentTrace",
  description: "Configure webhook enrichment rules to automatically enhance traces and spans",
};

export default function EnrichmentPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Enrichment Pipeline"
        description="Create and manage enrichment rules that automatically enhance traces with additional metadata"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <EnrichmentPipelineContent />
      </Suspense>
    </div>
  );
}

function EnrichmentPipelineContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Webhook Enrichment Pipeline</p>
      <p className="text-sm mt-2">Define rules to enrich traces with metadata from external sources</p>
      <p className="text-sm mt-1">Supports trigger-based conditions with custom transforms and dry-run testing</p>
    </div>
  );
}
