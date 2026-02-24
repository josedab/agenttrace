import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Training Pipeline | AgentTrace",
  description: "Curate traces into fine-tuning datasets and detect failure patterns",
};

export default function TrainingPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Training Pipeline"
        description="Curate traces into fine-tuning datasets and detect failure patterns"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <TrainingContent />
      </Suspense>
    </div>
  );
}

function TrainingContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Training Pipeline</p>
      <p className="text-sm mt-2">
        Curate trace data into fine-tuning datasets, export in standard formats, and detect recurring failure patterns.
      </p>
    </div>
  );
}
