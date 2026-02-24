import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Workflow Simulator | AgentTrace",
  description: "Design, simulate, and validate agent workflows visually",
};

export default function WorkflowsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Workflow Simulator"
        description="Design agent workflows with a visual editor and simulate execution paths"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <WorkflowsDashboardContent />
      </Suspense>
    </div>
  );
}

function WorkflowsDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Workflow Simulator Dashboard</p>
      <p className="text-sm mt-2">Create and simulate agent workflows with visual node-based editing</p>
      <p className="text-sm mt-1">Supports agents, tools, conditions, and branching logic</p>
    </div>
  );
}
