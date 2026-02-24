import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { RegressionDetectionDashboard } from "@/components/regression-detection/regression-detection-dashboard";

export const metadata = {
  title: "Regression Detection | AgentTrace",
  description:
    "ML-powered quality regression detection for monitoring agent performance degradation",
};

export default function RegressionDetectionPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Regression Detection"
        description="ML-powered quality regression detection across cost, latency, and error patterns"
      />
      <Suspense
        fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}
      >
        <RegressionDetectionDashboard />
      </Suspense>
    </div>
  );
}
