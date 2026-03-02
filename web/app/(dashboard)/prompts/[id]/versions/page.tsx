import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Prompt Version History | AgentTrace",
  description: "View prompt version history with diffs and impact analysis",
};

interface VersionsPageProps {
  params: { id: string };
}

export default function PromptVersionsPage({ params }: VersionsPageProps) {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Version History"
        description="Compare prompt versions with semantic diffs and impact analysis"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <PromptVersionsContent promptId={params.id} />
      </Suspense>
    </div>
  );
}

function PromptVersionsContent({ promptId }: { promptId: string }) {
  return (
    <div className="space-y-6">
      {/* Version comparison selector */}
      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Compare Versions</h3>
        <div className="flex items-center gap-4">
          <div className="flex-1">
            <label className="text-sm font-medium text-muted-foreground block mb-1">Base Version</label>
            <select className="w-full rounded-md border bg-background px-3 py-2 text-sm">
              <option>Select version...</option>
            </select>
          </div>
          <div className="flex items-center pt-5">
            <span className="text-muted-foreground">→</span>
          </div>
          <div className="flex-1">
            <label className="text-sm font-medium text-muted-foreground block mb-1">Compare Version</label>
            <select className="w-full rounded-md border bg-background px-3 py-2 text-sm">
              <option>Select version...</option>
            </select>
          </div>
          <div className="pt-5">
            <button className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
              Compare
            </button>
          </div>
        </div>
      </div>

      {/* Diff viewer */}
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Semantic Diff</h3>
          <div className="flex gap-2">
            <button className="rounded-md border px-3 py-1 text-sm hover:bg-accent">
              Side by Side
            </button>
            <button className="rounded-md border px-3 py-1 text-sm hover:bg-accent bg-accent">
              Unified
            </button>
          </div>
        </div>
        <div className="text-center py-12 text-muted-foreground">
          <p className="text-sm">Select two versions to see the diff</p>
          <p className="text-xs mt-1">Detects variable changes, instruction modifications, and system prompt alterations</p>
        </div>
      </div>

      {/* Change summary and impact */}
      <div className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-4">Change Summary</h3>
          <div className="space-y-3">
            <SummaryRow label="Variables" icon="🔤" description="No changes" />
            <SummaryRow label="Instructions" icon="📝" description="No changes" />
            <SummaryRow label="System Prompt" icon="⚙️" description="No changes" />
            <SummaryRow label="Configuration" icon="🔧" description="No changes" />
          </div>
          <div className="mt-4 p-3 rounded-md bg-muted">
            <p className="text-sm font-medium">Risk Level: —</p>
            <p className="text-xs text-muted-foreground mt-1">Select versions to assess risk</p>
          </div>
        </div>

        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-4">Impact Analysis</h3>
          <div className="space-y-3">
            <MetricRow label="Latency" before="—" after="—" delta="—" />
            <MetricRow label="Cost" before="—" after="—" delta="—" />
            <MetricRow label="Tokens" before="—" after="—" delta="—" />
            <MetricRow label="Error Rate" before="—" after="—" delta="—" />
            <MetricRow label="Quality Score" before="—" after="—" delta="—" />
          </div>
        </div>
      </div>

      {/* Review & approval gates */}
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Review & Approval</h3>
          <div className="flex gap-2">
            <button className="rounded-md bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-700">
              ✓ Approve
            </button>
            <button className="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700">
              ✗ Reject
            </button>
            <button className="rounded-md border px-4 py-2 text-sm font-medium hover:bg-accent">
              ↩ Rollback
            </button>
          </div>
        </div>
        <div className="text-center py-6 text-muted-foreground">
          <p className="text-sm">No reviews yet for this version</p>
          <p className="text-xs mt-1">Approve or reject prompt changes before they reach production</p>
        </div>
      </div>
    </div>
  );
}

function SummaryRow({ label, icon, description }: { label: string; icon: string; description: string }) {
  return (
    <div className="flex items-center justify-between py-2 border-b last:border-0">
      <div className="flex items-center gap-2">
        <span>{icon}</span>
        <span className="text-sm font-medium">{label}</span>
      </div>
      <span className="text-sm text-muted-foreground">{description}</span>
    </div>
  );
}

function MetricRow({ label, before, after, delta }: { label: string; before: string; after: string; delta: string }) {
  return (
    <div className="flex items-center justify-between py-2 border-b last:border-0">
      <span className="text-sm font-medium w-24">{label}</span>
      <span className="text-sm text-muted-foreground">{before}</span>
      <span className="text-muted-foreground">→</span>
      <span className="text-sm text-muted-foreground">{after}</span>
      <span className="text-sm font-medium w-20 text-right">{delta}</span>
    </div>
  );
}
