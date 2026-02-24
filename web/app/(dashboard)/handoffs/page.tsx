import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Agent Handoffs | AgentTrace",
  description: "Track inter-agent task handoffs and context preservation",
};

export default function AgentHandoffsPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Agent Handoffs" description="Track inter-agent task handoffs and context preservation" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AgentHandoffsContent />
      </Suspense>
    </div>
  );
}

function AgentHandoffsContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Agent Handoffs</p>
      <p className="text-sm mt-2">Track inter-agent task handoffs and context preservation</p>
    </div>
  );
}
