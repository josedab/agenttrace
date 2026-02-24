import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Security Scanner | AgentTrace",
  description: "Scan traces for security vulnerabilities and enforce security policies",
};

export default function SecurityPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Security Scanner"
        description="Scan traces for PII leaks, prompt injection, and data exfiltration risks"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <SecurityDashboardContent />
      </Suspense>
    </div>
  );
}

function SecurityDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Security Scanner Dashboard</p>
      <p className="text-sm mt-2">Detect security vulnerabilities including PII leaks and prompt injection attacks</p>
      <p className="text-sm mt-1">Define security policies with configurable rules and automated responses</p>
    </div>
  );
}
