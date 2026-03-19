import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { PromptCIDashboard } from "@/components/prompts/prompt-ci-dashboard";

export const metadata = {
  title: "Prompt CI Testing | AgentTrace",
  description: "Continuous integration regression testing for prompt quality",
};

export default function PromptCIPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Prompt CI Testing"
        description="Run regression tests on prompt changes in your CI/CD pipeline"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <PromptCIDashboard />
      </Suspense>
    </div>
  );
}
