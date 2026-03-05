import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { EvalMarketplaceBrowser } from "@/components/eval-marketplace/eval-marketplace-browser";

export const metadata = {
  title: "Eval Marketplace | AgentTrace",
  description: "Browse and import evaluation datasets for AI agent testing",
};

export default function EvalMarketplacePage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Eval Marketplace"
        description="Browse, search, and import evaluation datasets for agent testing"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <EvalMarketplaceBrowser />
      </Suspense>
    </div>
  );
}
