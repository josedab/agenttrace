import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Agent Versioning | AgentTrace",
  description: "Version control for agent configurations with one-click rollback",
};

export default function AgentVersionsPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Agent Versioning" description="Version control for agent configurations with one-click rollback" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AgentVersionsContent />
      </Suspense>
    </div>
  );
}

function AgentVersionsContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Agent Versioning</p>
      <p className="text-sm mt-2">Version control for agent configurations with diff comparison and one-click rollback.</p>
    </div>
  );
}
