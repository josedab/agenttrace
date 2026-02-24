import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Prompt Optimization | AgentTrace",
  description: "Optimize prompts with automated variant generation and failure analysis",
};

export default function PromptOptimizationPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Prompt Optimization"
        description="Generate and evaluate prompt variants to improve agent performance automatically"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <PromptOptimizationDashboardContent />
      </Suspense>
    </div>
  );
}

function PromptOptimizationDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Prompt Optimization Dashboard</p>
      <p className="text-sm mt-2">Analyze failure patterns and generate optimized prompt variants</p>
      <p className="text-sm mt-1">Supports automated scoring, approval workflows, and A/B testing</p>
    </div>
  );
}
