import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Knowledge Base | AgentTrace",
  description: "Browse and manage the trace annotation knowledge base",
};

export default function KnowledgeBasePage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Knowledge Base"
        description="Browse, search, and manage trace annotation patterns, root causes, and fixes"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <KnowledgeBaseContent />
      </Suspense>
    </div>
  );
}

function KnowledgeBaseContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Trace Annotation Knowledge Base</p>
      <p className="text-sm mt-2">Search and browse annotated patterns, root causes, fixes, and optimizations</p>
      <p className="text-sm mt-1">Categorized entries with full-text search and tag filtering</p>
    </div>
  );
}
