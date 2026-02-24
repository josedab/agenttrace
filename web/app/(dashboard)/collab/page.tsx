import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Collaboration Hub | AgentTrace",
  description: "Review queues, quality standards, and team activity feed",
};

export default function CollabPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Collaboration Hub"
        description="Manage review queues, enforce quality standards, and track team activity"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CollabDashboardContent />
      </Suspense>
    </div>
  );
}

function CollabDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Collaboration Hub Dashboard</p>
      <p className="text-sm mt-2">Create review queues, assign trace reviews, and define quality standards</p>
      <p className="text-sm mt-1">Includes real-time activity feed and team performance tracking</p>
    </div>
  );
}
