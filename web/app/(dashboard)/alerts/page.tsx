import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Smart Alerts | AgentTrace",
  description: "Trace-aware smart alerting with behavioral drift detection",
};

export default function AlertsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Smart Alerts"
        description="ML-powered alerting with trace awareness, behavioral drift detection, and multi-channel delivery"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AlertsDashboardContent />
      </Suspense>
    </div>
  );
}

function AlertsDashboardContent() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="Active Alerts" value="0" description="Currently firing" />
        <StatCard title="Rules Configured" value="0" description="Alert rules defined" />
        <StatCard title="Drift Detected" value="0" description="Behavioral drifts this week" />
        <StatCard title="Deliveries Sent" value="0" description="Notifications sent today" />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-4">Alert Rules</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Configure trace-aware alerting rules with threshold, behavioral drift, and pattern matching conditions.
          </p>
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">No alert rules configured yet.</p>
            <p className="text-xs mt-1">Create rules to monitor agent behavior and detect anomalies.</p>
          </div>
        </div>

        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-4">Recent Alert Events</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Track triggered alerts with trace context, drift details, and delivery status.
          </p>
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">No alert events yet.</p>
            <p className="text-xs mt-1">Events will appear here when alert rules fire.</p>
          </div>
        </div>
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Delivery Channels</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Configure Slack, PagerDuty, webhook, email, and Teams delivery channels.
        </p>
        <div className="grid gap-3 md:grid-cols-5">
          <ChannelCard name="Slack" icon="💬" configured={false} />
          <ChannelCard name="PagerDuty" icon="🔔" configured={false} />
          <ChannelCard name="Webhook" icon="🔗" configured={false} />
          <ChannelCard name="Email" icon="📧" configured={false} />
          <ChannelCard name="Teams" icon="👥" configured={false} />
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

function ChannelCard({ name, icon, configured }: { name: string; icon: string; configured: boolean }) {
  return (
    <div className="rounded-lg border p-3 text-center">
      <p className="text-2xl mb-1">{icon}</p>
      <p className="text-sm font-medium">{name}</p>
      <p className={`text-xs ${configured ? "text-green-600" : "text-muted-foreground"}`}>
        {configured ? "Connected" : "Not configured"}
      </p>
    </div>
  );
}
