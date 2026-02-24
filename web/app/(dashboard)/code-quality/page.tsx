import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { CodeQualityDashboard } from "@/components/code-quality/code-quality-dashboard";

export const metadata = {
  title: "Code Quality | AgentTrace",
  description: "Automated code evaluation pipeline for agent-generated code",
};

export default function CodeQualityPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Code Quality Scoring"
        description="Automated code evaluation pipeline analyzing agent-generated code with static analyzers and quality metrics"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CodeQualityDashboard />
      </Suspense>
    </div>
  );
}
