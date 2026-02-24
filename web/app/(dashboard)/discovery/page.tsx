import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Auto-Discovery | AgentTrace",
  description: "Automatically discover and instrument AI frameworks and components",
};

export default function DiscoveryPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Auto-Discovery"
        description="Automatically detect AI frameworks and instrument components in your project"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <DiscoveryDashboardContent />
      </Suspense>
    </div>
  );
}

function DiscoveryDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Auto-Discovery Dashboard</p>
      <p className="text-sm mt-2">Scan your project to discover LLM frameworks, agents, and tool integrations</p>
      <p className="text-sm mt-1">Supports LangChain, CrewAI, AutoGen, and more</p>
    </div>
  );
}
