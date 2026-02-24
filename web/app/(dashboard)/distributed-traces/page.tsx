import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Distributed Tracing | AgentTrace",
  description: "Trace agent actions across downstream services and infrastructure",
};

export default function DistributedTracingPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Distributed Tracing" description="Trace agent actions across downstream services and infrastructure" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <DistributedTracingContent />
      </Suspense>
    </div>
  );
}

function DistributedTracingContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Distributed Tracing</p>
      <p className="text-sm mt-2">Trace agent actions across downstream services and infrastructure</p>
    </div>
  );
}
