import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Cost Forecast & Budget Simulator | AgentTrace",
  description: "Forecast costs, simulate model routing changes, and manage budgets",
};

export default function CostForecastPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Cost Forecast & Budget Simulator"
        description="Predict future costs, simulate routing changes, and set budget alerts"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CostForecastContent />
      </Suspense>
    </div>
  );
}

function CostForecastContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Cost Forecast & Budget Simulator</p>
      <p className="text-sm mt-2">View cost predictions, simulate model routing changes, and configure budget plans</p>
      <p className="text-sm mt-1">Supports daily, weekly, and monthly forecasting with confidence intervals</p>
    </div>
  );
}
