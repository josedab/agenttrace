import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Cost Attribution | AgentTrace",
  description: "Attribute AI costs to business outcomes and calculate ROI",
};

export default function CostAttributionPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Cost Attribution"
        description="Attribute AI costs to business outcomes and calculate ROI"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CostAttributionContent />
      </Suspense>
    </div>
  );
}

function CostAttributionContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Cost Attribution Dashboard</p>
      <p className="text-sm mt-2">Map AI spending to business outcomes and measure return on investment</p>
      <p className="text-sm mt-1">Supports per-feature, per-team, and per-agent cost breakdowns</p>
    </div>
  );
}
