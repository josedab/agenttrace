import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { SelfHealingDashboard } from "@/components/guardrails/self-healing-dashboard";

export const metadata = {
  title: "Self-Healing Guardrails | AgentTrace",
  description: "Automatic remediation with retry, fallback, and circuit breaker policies for agent guardrails",
};

export default function SelfHealingGuardrailsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Self-Healing Guardrails"
        description="Automatic remediation with retry, fallback, and circuit breaker policies for agent guardrails"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <SelfHealingDashboard />
      </Suspense>
    </div>
  );
}
