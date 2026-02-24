import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Intent Verification | AgentTrace",
  description: "Verify that agents do what they say they'll do",
};

export default function IntentsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Intent Verification"
        description="Verify that agents do what they say they'll do"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <IntentsContent />
      </Suspense>
    </div>
  );
}

function IntentsContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Intent Verification Dashboard</p>
      <p className="text-sm mt-2">Declare intents before execution and verify outcomes match declarations</p>
      <p className="text-sm mt-1">Track intent alignment scores and drift over time</p>
    </div>
  );
}
