import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Team Intelligence | AgentTrace",
  description: "Team-wide analytics, cost tracking, and AI ROI calculations",
};

export default function TeamPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Team Intelligence"
        description="Team-wide analytics, cost tracking, and AI ROI calculations"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <TeamContent />
      </Suspense>
    </div>
  );
}

function TeamContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Team Intelligence</p>
      <p className="text-sm mt-2">
        Track team-wide AI usage analytics, monitor costs, and calculate return on investment for AI agent adoption.
      </p>
    </div>
  );
}
