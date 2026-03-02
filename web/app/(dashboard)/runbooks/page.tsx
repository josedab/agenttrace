import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Runbooks | AgentTrace",
  description: "Automated trace-to-runbook remediation",
};

export default function RunbooksPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Trace-to-Runbook Automation"
        description="Define YAML-based runbooks that automatically respond to trace failure patterns"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <RunbooksContent />
      </Suspense>
    </div>
  );
}

function RunbooksContent() {
  return (
    <div className="space-y-6">
      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="Active Runbooks" value="0" description="Currently active" />
        <StatCard title="Executions Today" value="0" description="Triggered today" />
        <StatCard title="Success Rate" value="—" description="Execution success" />
        <StatCard title="Pending Approval" value="0" description="Awaiting review" />
      </div>

      {/* Runbook editor */}
      <div className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-lg border bg-card overflow-hidden">
          <div className="border-b p-4 bg-muted/50 flex items-center justify-between">
            <h3 className="text-sm font-semibold">YAML Editor</h3>
            <div className="flex gap-2">
              <button className="rounded-md border px-3 py-1 text-xs hover:bg-accent">
                📋 Template
              </button>
              <button className="rounded-md bg-primary px-3 py-1 text-xs font-medium text-primary-foreground hover:bg-primary/90">
                Save & Validate
              </button>
            </div>
          </div>
          <div className="p-4">
            <pre className="text-xs bg-muted p-4 rounded-md font-mono overflow-x-auto min-h-[300px] whitespace-pre">
{`# Runbook Definition
name: high-cost-retry
description: Retry with cheaper model when cost exceeds threshold

triggers:
  - type: threshold
    conditions:
      metric: cost
      operator: gt
      value: 1.0
    description: Cost exceeds $1.00

  - type: error_match
    conditions:
      pattern: "rate_limit_exceeded"
    description: Rate limit hit

actions:
  - name: retry-with-cheaper-model
    type: retry_with_model
    parameters:
      model: gpt-4o-mini
      max_retries: "2"
    timeout: 60s
    on_failure: continue

  - name: notify-team
    type: send_notification
    parameters:
      channel: slack
      message: "High-cost trace remediated"
    on_failure: continue`}
            </pre>
            <p className="text-xs text-muted-foreground mt-2">
              💡 Use Monaco editor for syntax highlighting in production
            </p>
          </div>
        </div>

        {/* Trigger testing simulator */}
        <div className="space-y-4">
          <div className="rounded-lg border bg-card p-4">
            <h3 className="text-sm font-semibold mb-3">Trigger Testing Simulator</h3>
            <p className="text-sm text-muted-foreground mb-3">
              Test your runbook against a trace to see which triggers match.
            </p>
            <div className="space-y-3">
              <div>
                <label className="text-xs font-medium text-muted-foreground block mb-1">Trace ID</label>
                <input
                  type="text"
                  placeholder="Enter trace ID to test against..."
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm"
                />
              </div>
              <div className="flex gap-2">
                <button className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 flex-1">
                  🧪 Dry Run
                </button>
                <button className="rounded-md border px-4 py-2 text-sm font-medium hover:bg-accent flex-1">
                  ▶ Execute
                </button>
              </div>
            </div>
            <div className="mt-4 p-3 rounded-md bg-muted text-center text-sm text-muted-foreground">
              Run a test to see trigger matches and planned actions
            </div>
          </div>

          {/* Action types reference */}
          <div className="rounded-lg border bg-card p-4">
            <h3 className="text-sm font-semibold mb-3">Available Actions</h3>
            <div className="space-y-2">
              <ActionType icon="🔄" name="retry_with_model" desc="Retry with a different/cheaper model" />
              <ActionType icon="👤" name="escalate_to_human" desc="Escalate to human operator" />
              <ActionType icon="↩️" name="rollback_prompt" desc="Roll back to previous prompt version" />
              <ActionType icon="🌡️" name="adjust_temperature" desc="Adjust model temperature" />
              <ActionType icon="📢" name="send_notification" desc="Send Slack/email notification" />
              <ActionType icon="🔗" name="webhook" desc="Call custom webhook" />
            </div>
          </div>
        </div>
      </div>

      {/* Execution history */}
      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Execution History</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b">
                <th className="text-left py-3 px-4 font-medium">Runbook</th>
                <th className="text-left py-3 px-4 font-medium">Trace</th>
                <th className="text-left py-3 px-4 font-medium">Trigger</th>
                <th className="text-center py-3 px-4 font-medium">Status</th>
                <th className="text-right py-3 px-4 font-medium">Actions</th>
                <th className="text-right py-3 px-4 font-medium">Time</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td colSpan={6} className="text-center py-8 text-muted-foreground">
                  <p className="text-sm">No executions yet</p>
                  <p className="text-xs mt-1">Runbook executions will appear here with full audit trail</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function StatCard({ title, value, description }: { title: string; value: string; description: string }) {
  return (
    <div className="rounded-lg border bg-card p-4">
      <p className="text-sm font-medium text-muted-foreground">{title}</p>
      <p className="text-2xl font-bold">{value}</p>
      <p className="text-xs text-muted-foreground">{description}</p>
    </div>
  );
}

function ActionType({ icon, name, desc }: { icon: string; name: string; desc: string }) {
  return (
    <div className="flex items-center gap-2 p-2 rounded-md hover:bg-muted/50">
      <span>{icon}</span>
      <div>
        <code className="text-xs font-mono bg-muted px-1 rounded">{name}</code>
        <p className="text-xs text-muted-foreground">{desc}</p>
      </div>
    </div>
  );
}
