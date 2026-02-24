import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Prompt Cache | AgentTrace",
  description: "Analyze and configure intelligent prompt caching for cost reduction",
};

export default function PromptCachePage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Prompt Cache" description="Analyze and configure intelligent prompt caching for cost reduction" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <PromptCacheContent />
      </Suspense>
    </div>
  );
}

function PromptCacheContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Prompt Cache</p>
      <p className="text-sm mt-2">Analyze and configure intelligent prompt caching for cost reduction</p>
    </div>
  );
}
