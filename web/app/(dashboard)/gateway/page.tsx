import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "LLM Gateway | AgentTrace",
  description: "Route LLM API calls with smart routing, fallback chains, and cost optimization",
};

export default function GatewayPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="LLM Gateway"
        description="Reverse proxy for LLM API calls with automatic tracing, cost-optimized routing, and fallback chains"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <GatewayContent />
      </Suspense>
    </div>
  );
}

function GatewayContent() {
  return (
    <div className="space-y-6">
      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="Total Requests" value="0" description="Proxied through gateway" />
        <StatCard title="Avg Latency" value="—" description="Response time" />
        <StatCard title="Total Cost" value="$0.00" description="Estimated spend" />
        <StatCard title="Fallback Rate" value="0%" description="Requests using fallback" />
      </div>

      {/* Gateway Configs */}
      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Gateway Configurations</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Configure routing strategies, provider priorities, and fallback chains for your LLM API calls.
        </p>

        <div className="space-y-4">
          <div className="rounded-md border p-4">
            <h4 className="font-medium mb-2">Supported Providers</h4>
            <div className="flex gap-2 flex-wrap">
              {["OpenAI", "Anthropic", "Google", "Local Models"].map((provider) => (
                <span key={provider} className="px-2 py-1 rounded-full bg-primary/10 text-primary text-xs font-medium">
                  {provider}
                </span>
              ))}
            </div>
          </div>

          <div className="rounded-md border p-4">
            <h4 className="font-medium mb-2">Routing Strategies</h4>
            <div className="grid gap-2 md:grid-cols-3 text-sm">
              <div><span className="font-mono text-xs bg-muted px-1 rounded">cheapest</span> — Route to lowest cost provider</div>
              <div><span className="font-mono text-xs bg-muted px-1 rounded">fastest</span> — Route to lowest latency provider</div>
              <div><span className="font-mono text-xs bg-muted px-1 rounded">fallback</span> — Chain through providers on failure</div>
              <div><span className="font-mono text-xs bg-muted px-1 rounded">round_robin</span> — Distribute evenly across providers</div>
              <div><span className="font-mono text-xs bg-muted px-1 rounded">priority</span> — Use highest priority provider</div>
            </div>
          </div>
        </div>

        <div className="mt-6 text-center text-sm text-muted-foreground">
          No gateway configurations yet. Create one to start routing LLM requests.
        </div>
      </div>

      {/* Smart Routing Rules */}
      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-2">Smart Routing Rules</h3>
        <p className="text-sm text-muted-foreground">
          Define rules to automatically route requests based on token count, task complexity, or cost thresholds.
        </p>
        <div className="mt-4 text-center text-sm text-muted-foreground">
          Configure a gateway first to add routing rules.
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
