import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Canary Deployments | AgentTrace",
  description: "Progressive rollout system for agent configurations",
};

export default function CanaryPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Canary Deployments"
        description="Progressive rollout system for agent configurations with automated rollback on quality regression"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CanaryContent />
      </Suspense>
    </div>
  );
}

function CanaryContent() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="Active Deployments" value="0" description="Currently running" />
        <StatCard title="Completed" value="0" description="Successfully rolled out" />
        <StatCard title="Rolled Back" value="0" description="Automatically reverted" />
        <StatCard title="Success Rate" value="—" description="Deployment success" />
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Progressive Rollout Stages</h3>
        <div className="flex items-center gap-2">
          {[5, 25, 50, 100].map((pct, i) => (
            <div key={pct} className="flex items-center gap-2">
              <div className="rounded-full px-3 py-1 bg-muted text-sm font-medium">{pct}%</div>
              {i < 3 && <span className="text-muted-foreground">→</span>}
            </div>
          ))}
        </div>
        <p className="text-sm text-muted-foreground mt-3">
          Route traffic progressively to new agent versions. Automated rollback triggers on quality regression.
        </p>
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-2">Promotion Criteria</h3>
        <div className="grid gap-3 md:grid-cols-2 text-sm">
          <div className="p-3 rounded-md border">
            <span className="font-medium">Evaluator Scores</span>
            <p className="text-muted-foreground">Minimum quality score threshold</p>
          </div>
          <div className="p-3 rounded-md border">
            <span className="font-medium">Cost Thresholds</span>
            <p className="text-muted-foreground">Maximum cost increase percentage</p>
          </div>
          <div className="p-3 rounded-md border">
            <span className="font-medium">Latency Budgets</span>
            <p className="text-muted-foreground">Maximum response time</p>
          </div>
          <div className="p-3 rounded-md border">
            <span className="font-medium">Error Rates</span>
            <p className="text-muted-foreground">Maximum acceptable error rate</p>
          </div>
        </div>
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-2">Deployments</h3>
        <div className="text-center text-sm text-muted-foreground py-8">
          No canary deployments yet. Create one to start progressive rollout.
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
