import { Suspense } from "react";
import { Metadata } from "next";
import { TraceDiffViewer } from "@/components/trace-diff/trace-diff-viewer";

export const metadata: Metadata = {
  title: "Trace Diff | AgentTrace",
  description: "Compare traces side-by-side with structural diff and regression bisect",
};

export default function TraceDiffPage({
  searchParams,
}: {
  searchParams: { left?: string; right?: string };
}) {
  return (
    <div className="container mx-auto py-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Trace Diff & Regression Bisect</h1>
        <p className="text-muted-foreground mt-1">
          Compare two traces side-by-side or bisect through trace history to find regressions
        </p>
      </div>
      <Suspense fallback={<div className="animate-pulse h-96 bg-muted rounded-lg" />}>
        <TraceDiffViewer
          leftTraceId={searchParams.left}
          rightTraceId={searchParams.right}
        />
      </Suspense>
    </div>
  );
}
