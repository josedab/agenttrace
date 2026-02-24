import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Cost Alerts | AgentTrace",
  description: "Cost alerting rules, circuit breakers, and spend monitoring",
};

export default function CostAlertsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Cost Alerts"
        description="Configure alert rules, circuit breakers, and real-time cost monitoring"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CostAlertsDashboardContent />
      </Suspense>
    </div>
  );
}

function CostAlertsDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Cost Alerts Dashboard</p>
      <p className="text-sm mt-2">Set up cost alert rules with configurable thresholds and notification channels</p>
      <p className="text-sm mt-1">Includes circuit breaker protection to prevent runaway spending</p>
    </div>
  );
}
