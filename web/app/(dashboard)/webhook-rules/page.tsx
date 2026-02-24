import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Webhook Orchestration | AgentTrace",
  description: "Create event-driven rules for automated alerting",
};

export default function WebhookRulesPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Webhook Orchestration"
        description="Create event-driven rules for automated alerting"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <WebhookRulesContent />
      </Suspense>
    </div>
  );
}

function WebhookRulesContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Webhook Orchestration</p>
      <p className="text-sm mt-2">
        Create event-driven webhook rules for automated alerting and integrate with external services.
      </p>
    </div>
  );
}
