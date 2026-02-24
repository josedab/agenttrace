import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "AI Debugger | AgentTrace",
  description: "AI-powered debugging with root cause analysis and suggested fixes",
};

export default function AIDebuggerPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="AI Debugger"
        description="Ask questions about traces and get AI-powered root cause analysis with fix suggestions"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AIDebuggerDashboardContent />
      </Suspense>
    </div>
  );
}

function AIDebuggerDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">AI Debugger Dashboard</p>
      <p className="text-sm mt-2">Debug traces with natural language queries and automated root cause analysis</p>
      <p className="text-sm mt-1">Get actionable fix suggestions with impact and effort estimates</p>
    </div>
  );
}
