import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { CostAutopilotDashboard } from "@/components/cost-optimizer/cost-autopilot-dashboard";

export const metadata = {
  title: "Cost Autopilot | AgentTrace",
  description: "ML-powered cost optimization with hotspot detection, predictions, and automated rules",
};

export default function CostAutopilotPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Agent Cost Autopilot"
        description="ML-powered cost optimization with hotspot detection, budget predictions, and automated savings rules"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CostAutopilotDashboard />
      </Suspense>
    </div>
  );
}
