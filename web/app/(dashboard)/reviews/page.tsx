import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { TraceReviewPanel } from "@/components/collaboration/trace-review-panel";

export const metadata = {
  title: "Trace Reviews | AgentTrace",
  description: "Collaborative trace review workflow with threaded comments, approvals, and integrations",
};

export default function ReviewsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Trace Reviews"
        description="Collaborative review system for agent traces — threaded comments, @mentions, approval workflows, and Slack/Teams integration"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <TraceReviewPanel />
      </Suspense>
    </div>
  );
}
