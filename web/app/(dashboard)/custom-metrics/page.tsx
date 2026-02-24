import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Custom Metrics | AgentTrace",
  description: "Define custom metrics, build dashboards, and set alerts",
};

export default function CustomMetricsPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Custom Metrics" description="Define custom metrics, build dashboards, and set alerts" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CustomMetricsContent />
      </Suspense>
    </div>
  );
}

function CustomMetricsContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Custom Metrics</p>
      <p className="text-sm mt-2">Define custom metrics, build dashboards, and set alerts</p>
    </div>
  );
}
