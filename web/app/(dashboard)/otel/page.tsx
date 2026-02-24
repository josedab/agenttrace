import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "OpenTelemetry | AgentTrace",
  description: "OpenTelemetry compatibility, export destinations, and collector configuration",
};

export default function OTelPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="OpenTelemetry Compatibility"
        description="Export traces to OpenTelemetry-compatible backends and generate collector configs"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <OTelDashboardContent />
      </Suspense>
    </div>
  );
}

function OTelDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">OpenTelemetry Dashboard</p>
      <p className="text-sm mt-2">Configure export destinations for OTLP, Jaeger, Zipkin, and Prometheus</p>
      <p className="text-sm mt-1">Generate collector configs and manage field mappings</p>
    </div>
  );
}
