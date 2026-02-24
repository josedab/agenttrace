import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Cost Guardrails | AgentTrace",
  description: "Budget enforcement, cost policies, and spend forecasting for agents",
};

export default function CostGuardrailsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Cost Guardrails"
        description="Enforce budget policies, monitor spend, and forecast costs across your agents"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CostGuardrailsDashboardContent />
      </Suspense>
    </div>
  );
}

function CostGuardrailsDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Cost Guardrails Dashboard</p>
      <p className="text-sm mt-2">Configure budget policies with hard limits, soft limits, and rate controls</p>
      <p className="text-sm mt-1">Supports real-time spend tracking, violation alerts, and cost forecasting</p>
    </div>
  );
}
