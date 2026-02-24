import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Multi-Agent Debugger | AgentTrace",
  description: "Debug multi-agent orchestration with step-through execution and message flow visualization",
};

export default function OrchestrationPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Multi-Agent Debugger" description="Debug multi-agent orchestration with step-through execution and message flow visualization" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <OrchestrationContent />
      </Suspense>
    </div>
  );
}

function OrchestrationContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Multi-Agent Debugger</p>
      <p className="text-sm mt-2">Step-through debugging for multi-agent workflows with message flow visualization and breakpoints.</p>
    </div>
  );
}
