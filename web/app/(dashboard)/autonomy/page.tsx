import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Autonomy Gradient | AgentTrace",
  description: "Configure agent autonomy levels and track trust evolution",
};

export default function AutonomyPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Autonomy Gradient"
        description="Configure agent autonomy levels and track trust evolution"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AutonomyContent />
      </Suspense>
    </div>
  );
}

function AutonomyContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Autonomy Gradient Dashboard</p>
      <p className="text-sm mt-2">Configure autonomy levels per agent and monitor trust scores over time</p>
      <p className="text-sm mt-1">Supports graduated autonomy from supervised to fully autonomous modes</p>
    </div>
  );
}
