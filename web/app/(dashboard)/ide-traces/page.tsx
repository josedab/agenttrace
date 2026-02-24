import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "IDE Trace Viewer | AgentTrace",
  description: "View trace data as inline code annotations in your IDE",
};

export default function IDETracesPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="IDE Trace Viewer"
        description="Map traces to source code with inline annotations for calls, errors, and performance"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <IDETracesDashboardContent />
      </Suspense>
    </div>
  );
}

function IDETracesDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">IDE Trace Viewer Dashboard</p>
      <p className="text-sm mt-2">Browse file-level trace mappings and line-by-line annotations</p>
      <p className="text-sm mt-1">Supports VS Code extension integration, batch file mapping, and coverage visualization</p>
    </div>
  );
}
