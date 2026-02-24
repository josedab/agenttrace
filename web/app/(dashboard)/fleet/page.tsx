import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Fleet Management | AgentTrace",
  description: "Centralized management for all your AI agents across projects",
};

export default function FleetPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Fleet Management" description="Centralized management for all your AI agents across projects" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <FleetContent />
      </Suspense>
    </div>
  );
}

function FleetContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Fleet Management</p>
      <p className="text-sm mt-2">Centralized management for all your AI agents with policies, bulk updates, and scaling controls.</p>
    </div>
  );
}
