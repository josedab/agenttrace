import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Prompt Optimization Lab | AgentTrace",
  description: "A/B test prompt variants and get optimization suggestions",
};

export default function PromptLabPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Prompt Optimization Lab"
        description="A/B test prompt variants and get optimization suggestions"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <PromptLabContent />
      </Suspense>
    </div>
  );
}

function PromptLabContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Prompt Optimization Lab</p>
      <p className="text-sm mt-2">
        Create A/B test experiments for prompt variants and receive AI-powered optimization suggestions.
      </p>
    </div>
  );
}
