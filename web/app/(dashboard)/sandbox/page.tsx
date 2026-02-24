import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Security Sandbox | AgentTrace",
  description: "Review and approve agent actions before they're applied",
};

export default function SandboxPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Security Sandbox"
        description="Review and approve agent actions before they're applied"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <SandboxContent />
      </Suspense>
    </div>
  );
}

function SandboxContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Security Sandbox</p>
      <p className="text-sm mt-2">
        Review pending agent actions, assess risk levels, and approve or reject proposed changes before execution.
      </p>
    </div>
  );
}
