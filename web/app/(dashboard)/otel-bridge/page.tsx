import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "OpenTelemetry Bridge | AgentTrace",
  description: "Configure OpenTelemetry span export and import with external observability platforms",
};

export default function OTelBridgePage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="OpenTelemetry Bridge"
        description="Export and import spans with Jaeger, Tempo, Datadog, and other OTel-compatible platforms"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <OTelBridgeContent />
      </Suspense>
    </div>
  );
}

function OTelBridgeContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">OpenTelemetry Bridge</p>
      <p className="text-sm mt-2">Configure bidirectional span export and import with OTel-compatible platforms</p>
      <p className="text-sm mt-1">Supports Jaeger, Grafana Tempo, Datadog, OTLP, and Zipkin destinations</p>
    </div>
  );
}
