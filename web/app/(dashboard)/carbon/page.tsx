import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Carbon Tracking | AgentTrace",
  description: "Monitor energy consumption and CO2 emissions from AI agent operations",
};

export default function CarbonTrackingPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Carbon Tracking" description="Monitor energy consumption and CO2 emissions from AI agent operations" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CarbonTrackingContent />
      </Suspense>
    </div>
  );
}

function CarbonTrackingContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Carbon Tracking</p>
      <p className="text-sm mt-2">Monitor energy consumption and CO2 emissions from AI agent operations</p>
    </div>
  );
}
