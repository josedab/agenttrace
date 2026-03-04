import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { CodeImpactMap } from "@/components/code-impact/code-impact-map";

export const metadata = {
  title: "Code Impact Map | AgentTrace",
  description: "Visualize the code impact of AI agent traces across your codebase",
};

export default function CodeImpactPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Code Impact Map"
        description="Visualize file changes, lines added/removed, and complexity across traces"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CodeImpactMap />
      </Suspense>
    </div>
  );
}
