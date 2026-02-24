import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Root Cause Analysis | AgentTrace",
  description: "AI-powered analysis of agent failures with automated remediation suggestions",
};

export default function RCAPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Root Cause Analysis" description="AI-powered analysis of agent failures with automated remediation suggestions" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <RCAContent />
      </Suspense>
    </div>
  );
}

function RCAContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Root Cause Analysis</p>
      <p className="text-sm mt-2">AI-powered failure analysis with causal chains, remediation suggestions, and pattern detection.</p>
    </div>
  );
}
