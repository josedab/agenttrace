import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Cost Predictions | AgentTrace",
  description: "Predict cost, latency, and quality before running agent tasks",
};

export default function PredictionsPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Cost Predictions" description="Predict cost, latency, and quality before running agent tasks" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <PredictionsContent />
      </Suspense>
    </div>
  );
}

function PredictionsContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Cost Predictions</p>
      <p className="text-sm mt-2">Predict cost, latency, and quality metrics before running agent tasks with budget approval workflows.</p>
    </div>
  );
}
