import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Cost Optimizer | AgentTrace",
  description: "Optimize AI agent costs with autopilot recommendations",
};

export default function CostOptimizerPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Cost Optimizer"
        description="Analyze costs, get optimization recommendations, and configure autopilot savings"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CostOptimizerContent />
      </Suspense>
    </div>
  );
}

function CostOptimizerContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Cost Optimization Autopilot</p>
      <p className="text-sm mt-2">View cost forecasts, model recommendations, and configure automatic optimization</p>
      <p className="text-sm mt-1">Supports daily/monthly budgets with conservative, balanced, or aggressive optimization</p>
    </div>
  );
}
