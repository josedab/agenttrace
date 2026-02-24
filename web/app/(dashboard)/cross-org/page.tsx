import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Cross-Org Benchmarks | AgentTrace",
  description: "Compare agent performance anonymously across organizations",
};

export default function CrossOrgPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Cross-Org Benchmarks"
        description="Compare agent performance anonymously across organizations"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <CrossOrgContent />
      </Suspense>
    </div>
  );
}

function CrossOrgContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Cross-Org Benchmarks Dashboard</p>
      <p className="text-sm mt-2">Submit anonymized metrics and compare against industry benchmarks</p>
      <p className="text-sm mt-1">Privacy-preserving comparisons across organizations and industries</p>
    </div>
  );
}
