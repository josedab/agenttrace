import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Trace Reviews | AgentTrace",
  description: "Collaborative trace review workflow with PR-style annotations",
};

export default function ReviewsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Trace Reviews"
        description="GitHub PR-style review system for agent traces — annotate, flag issues, approve or reject"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <ReviewsContent />
      </Suspense>
    </div>
  );
}

function ReviewsContent() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="Open Reviews" value="0" description="Awaiting review" />
        <StatCard title="In Review" value="0" description="Being reviewed" />
        <StatCard title="Approved" value="0" description="Total approved" />
        <StatCard title="Overdue SLA" value="0" description="Past due date" />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-2">Review Queue</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Reviews sorted by priority and SLA deadlines.
          </p>
          <div className="text-center text-sm text-muted-foreground py-6">
            No reviews in queue.
          </div>
        </div>

        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-2">Review Features</h3>
          <div className="space-y-3 text-sm">
            <div className="flex gap-2 items-center">
              <span className="w-2 h-2 rounded-full bg-green-500" />
              Annotate specific observations in traces
            </div>
            <div className="flex gap-2 items-center">
              <span className="w-2 h-2 rounded-full bg-yellow-500" />
              Flag quality issues with priority levels
            </div>
            <div className="flex gap-2 items-center">
              <span className="w-2 h-2 rounded-full bg-blue-500" />
              Request re-runs with different parameters
            </div>
            <div className="flex gap-2 items-center">
              <span className="w-2 h-2 rounded-full bg-purple-500" />
              Approve or reject agent outputs
            </div>
            <div className="flex gap-2 items-center">
              <span className="w-2 h-2 rounded-full bg-red-500" />
              SLA tracking with escalation rules
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function StatCard({ title, value, description }: { title: string; value: string; description: string }) {
  return (
    <div className="rounded-lg border bg-card p-4">
      <p className="text-sm text-muted-foreground">{title}</p>
      <p className="text-2xl font-bold">{value}</p>
      <p className="text-xs text-muted-foreground">{description}</p>
    </div>
  );
}
