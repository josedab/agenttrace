import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { DashboardOverview } from "@/components/dashboard/dashboard-overview";
import { DashboardSkeleton } from "@/components/dashboard/dashboard-skeleton";

export const metadata = {
  title: "Dashboard | AgentTrace",
  description: "Overview of your AI agent observability metrics",
};

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Dashboard"
        description="Overview of your AI agent performance and costs"
      />
      <Suspense fallback={<DashboardSkeleton />}>
        <DashboardOverview />
      </Suspense>
    </div>
  );
}
