import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { FederatedAnalyticsDashboardComponent } from "@/components/federated/federated-analytics-dashboard";

export const metadata = {
  title: "Federated Trace Analytics | AgentTrace",
  description: "Cross-organization trace analytics with differential privacy benchmarking",
};

export default function FederatedAnalyticsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Federated Trace Analytics"
        description="Cross-organization benchmarking with differential privacy — compare metrics without sharing raw data"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <FederatedAnalyticsDashboardComponent />
      </Suspense>
    </div>
  );
}
