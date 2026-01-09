import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { TraceList } from "@/components/traces/trace-list";
import { TraceListSkeleton } from "@/components/traces/trace-list-skeleton";
import { TraceFilters } from "@/components/traces/trace-filters";

export const metadata = {
  title: "Traces | AgentTrace",
  description: "Explore and analyze your AI agent traces",
};

interface TracesPageProps {
  searchParams: {
    q?: string;
    level?: string;
    startDate?: string;
    endDate?: string;
    minLatency?: string;
    maxLatency?: string;
    minCost?: string;
    maxCost?: string;
  };
}

export default function TracesPage({ searchParams }: TracesPageProps) {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Traces"
        description="Explore and analyze your AI agent traces"
      />
      <TraceFilters />
      <Suspense fallback={<TraceListSkeleton />}>
        <TraceList searchParams={searchParams} />
      </Suspense>
    </div>
  );
}
