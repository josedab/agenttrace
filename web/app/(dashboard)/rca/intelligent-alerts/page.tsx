import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { RCADashboard } from "@/components/rca/rca-dashboard";

export const metadata = {
  title: "Intelligent Alerts | AgentTrace",
  description: "Context-aware alerting with root cause analysis and multi-channel delivery",
};

export default function IntelligentAlertsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Intelligent Alerts"
        description="Context-aware alerting with root cause analysis, anomaly correlation, and multi-channel delivery"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <RCADashboard />
      </Suspense>
    </div>
  );
}
