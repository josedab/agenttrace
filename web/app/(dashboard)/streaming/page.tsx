import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { RealtimeStreamPanel } from "@/components/streaming/realtime-stream-panel";

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
        <RealtimeStreamPanel />
      </Suspense>
    </div>
  );
}
