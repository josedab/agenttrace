import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Live Streaming | AgentTrace",
  description: "Real-time agent activity monitoring",
};

export default function StreamingPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Live Streaming"
        description="Monitor your AI coding agents in real-time"
      />
      <Suspense
        fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}
      >
        <StreamingDashboardWrapper />
      </Suspense>
    </div>
  );
}

function StreamingDashboardWrapper() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Select a trace to begin streaming</p>
      <p className="text-sm mt-2">
        Active agent sessions will appear here with real-time metrics and
        activity feeds
      </p>
    </div>
  );
}
