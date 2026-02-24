import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Regression Suite | AgentTrace",
  description: "Golden datasets and regression testing for agent outputs",
};

export default function RegressionPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Regression Suite"
        description="Create golden datasets and run regression tests to catch output quality degradation"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <RegressionDashboardContent />
      </Suspense>
    </div>
  );
}

function RegressionDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Regression Suite Dashboard</p>
      <p className="text-sm mt-2">Build golden datasets with expected outputs and run automated regression tests</p>
      <p className="text-sm mt-1">Supports baseline comparisons, pass rate tracking, and diff analysis</p>
    </div>
  );
}
