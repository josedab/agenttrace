import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Agent Replay | AgentTrace",
  description: "Time-travel debugging and session replay for agent traces",
};

export default function ReplayPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Agent Replay"
        description="Time-travel debugging with session replay, branching, and collaborative sharing"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <ReplayDashboardContent />
      </Suspense>
    </div>
  );
}

function ReplayDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Agent Replay Dashboard</p>
      <p className="text-sm mt-2">Record, replay, and branch agent sessions for time-travel debugging</p>
      <p className="text-sm mt-1">Supports timeline scrubbing, event inspection, and session sharing</p>
    </div>
  );
}
