import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { ReplayDashboard } from "./replay-dashboard";

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
        <ReplayDashboard />
      </Suspense>
    </div>
  );
}
