import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { AdapterManager } from "@/components/adapters/adapter-manager";

export const metadata = {
  title: "Protocol Adapters | AgentTrace",
  description: "Manage agent framework adapters for zero-config trace integration",
};

export default function AdaptersPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Protocol Adapters"
        description="Zero-config integration with any agent framework — LangChain, CrewAI, AutoGen, LangGraph, and more"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AdapterManager />
      </Suspense>
    </div>
  );
}
