import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Agent Benchmarks | AgentTrace",
  description: "Performance benchmarks and leaderboards for agent evaluation",
};

export default function AgentBenchmarksPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Agent Benchmarks"
        description="Create benchmark suites, run evaluations, and compare agent performance on leaderboards"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AgentBenchmarksDashboardContent />
      </Suspense>
    </div>
  );
}

function AgentBenchmarksDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Agent Benchmarks Dashboard</p>
      <p className="text-sm mt-2">Define benchmark suites with tasks, run agents against them, and track rankings</p>
      <p className="text-sm mt-1">Supports custom scoring, version comparison, and public leaderboards</p>
    </div>
  );
}
