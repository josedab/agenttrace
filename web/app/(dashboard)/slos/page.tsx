import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "SLOs | AgentTrace",
  description: "Define and track Service Level Objectives for agent performance",
};

export default function SLOsPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="SLOs" description="Define and track Service Level Objectives for agent performance" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <SLOsContent />
      </Suspense>
    </div>
  );
}

function SLOsContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">SLOs</p>
      <p className="text-sm mt-2">Define and track Service Level Objectives for agent performance</p>
    </div>
  );
}
