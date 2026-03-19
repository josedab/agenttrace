import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { ABTestingDashboard } from "@/components/prompts/ab-testing-dashboard";

export const metadata = {
  title: "Prompt A/B Testing | AgentTrace",
  description: "Run A/B tests on prompt variants with statistical significance analysis",
};

export default function ABTestingPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Prompt A/B Testing"
        description="Production traffic splitting for prompt variants with statistical significance calculation and automated winner selection"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <ABTestingDashboard />
      </Suspense>
    </div>
  );
}
