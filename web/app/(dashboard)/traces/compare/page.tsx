import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Trace Comparison | AgentTrace",
  description: "Compare traces side-by-side with metric analysis",
};

export default function TraceComparePage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Trace Comparison Matrix"
        description="Select 2-10 traces and view structured comparisons: latency, cost, tool usage, and decision paths"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <TraceCompareContent />
      </Suspense>
    </div>
  );
}

function TraceCompareContent() {
  return (
    <div className="space-y-6">
      {/* Trace selector */}
      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Select Traces</h3>
        <div className="flex gap-4 items-end">
          <div className="flex-1">
            <label className="text-sm font-medium text-muted-foreground block mb-1">
              Enter trace IDs (comma-separated)
            </label>
            <input
              type="text"
              placeholder="trace-id-1, trace-id-2, ..."
              className="w-full rounded-md border bg-background px-3 py-2 text-sm"
            />
          </div>
          <button className="rounded-md bg-primary px-6 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
            Compare
          </button>
        </div>
      </div>

      {/* Metric comparison grid */}
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Metric Comparison</h3>
          <div className="flex gap-2">
            <button className="rounded-md border px-3 py-1 text-sm hover:bg-accent">
              📊 Export CSV
            </button>
            <button className="rounded-md border px-3 py-1 text-sm hover:bg-accent">
              🔗 Share Link
            </button>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b">
                <th className="text-left py-3 px-4 font-medium">Metric</th>
                <th className="text-right py-3 px-4 font-medium text-muted-foreground">Min</th>
                <th className="text-right py-3 px-4 font-medium text-muted-foreground">Max</th>
                <th className="text-right py-3 px-4 font-medium text-muted-foreground">Avg</th>
              </tr>
            </thead>
            <tbody>
              <ComparisonRow metric="Latency" unit="ms" />
              <ComparisonRow metric="Total Cost" unit="USD" />
              <ComparisonRow metric="Total Tokens" unit="tokens" />
              <ComparisonRow metric="Span Count" unit="spans" />
              <ComparisonRow metric="Error Rate" unit="%" />
            </tbody>
          </table>
        </div>
        <p className="text-center text-sm text-muted-foreground mt-6">
          Select traces above to populate the comparison matrix
        </p>
      </div>

      {/* Side-by-side panels */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Tool usage comparison */}
        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-4">Tool Usage Differences</h3>
          <p className="text-sm text-muted-foreground">
            Compare which tools each trace used, call counts, and execution times.
          </p>
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">No data yet</p>
          </div>
        </div>

        {/* Topology diff */}
        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-4">Topology Diff</h3>
          <p className="text-sm text-muted-foreground">
            Visualize structural differences between trace spans using a flow diagram.
          </p>
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">No data yet</p>
          </div>
        </div>
      </div>
    </div>
  );
}

function ComparisonRow({ metric, unit }: { metric: string; unit: string }) {
  return (
    <tr className="border-b last:border-0 hover:bg-muted/50">
      <td className="py-3 px-4 font-medium">
        {metric} <span className="text-muted-foreground text-xs">({unit})</span>
      </td>
      <td className="text-right py-3 px-4 text-muted-foreground">—</td>
      <td className="text-right py-3 px-4 text-muted-foreground">—</td>
      <td className="text-right py-3 px-4 text-muted-foreground">—</td>
    </tr>
  );
}
