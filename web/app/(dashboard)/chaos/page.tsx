import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Chaos Testing | AgentTrace",
  description: "Run resilience experiments and measure agent fault tolerance",
};

export default function ChaosTestingPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Chaos Testing" description="Run resilience experiments and measure agent fault tolerance" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <ChaosTestingContent />
      </Suspense>
    </div>
  );
}

function ChaosTestingContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Chaos Testing</p>
      <p className="text-sm mt-2">Run resilience experiments and measure agent fault tolerance</p>
    </div>
  );
}
