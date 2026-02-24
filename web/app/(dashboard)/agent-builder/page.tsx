import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Agent Builder | AgentTrace",
  description: "Generate optimal agent configurations from natural language task descriptions",
};

export default function AgentBuilderPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Agent Builder" description="Generate optimal agent configurations from natural language task descriptions" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AgentBuilderContent />
      </Suspense>
    </div>
  );
}

function AgentBuilderContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Agent Builder</p>
      <p className="text-sm mt-2">Generate optimal agent configurations from natural language task descriptions and deploy them instantly.</p>
    </div>
  );
}
