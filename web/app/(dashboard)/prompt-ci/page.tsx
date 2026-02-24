import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

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
        <PromptCIDashboardContent />
      </Suspense>
    </div>
  );
}

function PromptCIDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Prompt CI Testing Dashboard</p>
      <p className="text-sm mt-2">Create baselines, compare prompt versions, and catch regressions before deployment</p>
      <p className="text-sm mt-1">Supports GitHub Actions, GitLab CI, Jenkins, and CircleCI integrations</p>
    </div>
  );
}
