import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Diff Intelligence | AgentTrace",
  description: "Analyze code quality of agent-produced changes",
};

export default function DiffIntelligencePage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Diff Intelligence"
        description="Analyze and score code changes produced by your AI coding agents"
      />
      <Suspense
        fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}
      >
        <DiffIntelligenceContent />
      </Suspense>
    </div>
  );
}

function DiffIntelligenceContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Agent Code Quality Analysis</p>
      <p className="text-sm mt-2">
        Select a trace to analyze the code changes made by your AI agent
      </p>
      <p className="text-sm mt-1">
        Includes security scanning, complexity analysis, and quality scoring
      </p>
    </div>
  );
}
