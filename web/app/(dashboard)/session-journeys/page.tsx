import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { SessionJourneyTimeline } from "@/components/session-journeys/session-journey-timeline";

export const metadata = {
  title: "Session Journeys | AgentTrace",
  description: "Visualize agent session workflow phases and timelines",
};

export default function SessionJourneysPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Session Journeys"
        description="Visualize workflow phases, durations, and costs across agent sessions"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <SessionJourneyTimeline />
      </Suspense>
    </div>
  );
}
