import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Federated Learning | AgentTrace",
  description: "Privacy-preserving insights from cross-organization data",
};

export default function FederatedPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Federated Learning"
        description="Privacy-preserving insights from cross-organization data"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <FederatedContent />
      </Suspense>
    </div>
  );
}

function FederatedContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Federated Learning Dashboard</p>
      <p className="text-sm mt-2">Join federation rings and derive insights without sharing raw data</p>
      <p className="text-sm mt-1">Differential privacy and secure aggregation ensure data never leaves your org</p>
    </div>
  );
}
