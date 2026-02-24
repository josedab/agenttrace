import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Synthetic Data | AgentTrace",
  description: "Generate test data and adversarial inputs for agent evaluation",
};

export default function SyntheticDataPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Synthetic Data" description="Generate test data and adversarial inputs for agent evaluation" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <SyntheticDataContent />
      </Suspense>
    </div>
  );
}

function SyntheticDataContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Synthetic Data</p>
      <p className="text-sm mt-2">Generate test data and adversarial inputs for agent evaluation</p>
    </div>
  );
}
