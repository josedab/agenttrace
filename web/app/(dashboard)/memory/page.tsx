import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Agent Memory | AgentTrace",
  description: "Visualize context window usage, memory retention, and optimization opportunities",
};

export default function AgentMemoryPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Agent Memory" description="Visualize context window usage, memory retention, and optimization opportunities" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AgentMemoryContent />
      </Suspense>
    </div>
  );
}

function AgentMemoryContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Agent Memory</p>
      <p className="text-sm mt-2">Visualize context window usage, memory retention, and optimization opportunities</p>
    </div>
  );
}
