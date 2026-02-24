import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Observability Copilot | AgentTrace",
  description: "AI-powered trace analysis and proactive optimization suggestions",
};

export default function CopilotPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Observability Copilot"
        description="AI-powered trace analysis and proactive optimization suggestions"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CopilotContent />
      </Suspense>
    </div>
  );
}

function CopilotContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Observability Copilot</p>
      <p className="text-sm mt-2">Ask questions about your traces and get AI-powered analysis and suggestions</p>
      <p className="text-sm mt-1">Proactive insights, anomaly explanations, and optimization recommendations</p>
    </div>
  );
}
