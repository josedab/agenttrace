import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Anomaly Detection | AgentTrace",
  description: "Monitor and detect anomalies in agent behavior",
};

export default function AnomalyPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Anomaly Detection"
        description="Monitor agent behavior and detect anomalies in cost, latency, and error patterns"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AnomalyDashboardContent />
      </Suspense>
    </div>
  );
}

function AnomalyDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Anomaly Detection Dashboard</p>
      <p className="text-sm mt-2">Configure detection rules and alert channels to monitor your agents</p>
      <p className="text-sm mt-1">Supports Z-Score, IQR, MAD, and Moving Average detection methods</p>
    </div>
  );
}
