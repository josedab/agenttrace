import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Compliance Reports | AgentTrace",
  description: "Generate EU AI Act and SOC 2 compliance reports",
};

export default function ComplianceReportsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Compliance Reports"
        description="Generate EU AI Act and SOC 2 compliance reports"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <ComplianceReportsContent />
      </Suspense>
    </div>
  );
}

function ComplianceReportsContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Compliance Reports</p>
      <p className="text-sm mt-2">
        Generate and manage EU AI Act, SOC 2, and other regulatory compliance reports from your trace data.
      </p>
    </div>
  );
}
