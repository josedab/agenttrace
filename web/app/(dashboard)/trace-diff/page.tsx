import { Suspense } from 'react';
import { Metadata } from 'next';
import { TraceDiffViewer } from '@/components/trace-diff/trace-diff-viewer';

export const metadata: Metadata = {
  title: 'Trace Diff | AgentTrace',
  description: 'Compare traces side-by-side with structural diff and regression bisect',
};

export default async function TraceDiffPage({
  searchParams,
}: {
  searchParams: Promise<{ left?: string; right?: string }>;
}) {
  const { left, right } = await searchParams;

  return (
    <div className="container mx-auto space-y-6 py-6">
      <div>
        <h1 className="text-2xl font-bold">Trace Diff & Regression Bisect</h1>
        <p className="mt-1 text-muted-foreground">
          Compare two traces side-by-side or bisect through trace history to find regressions
        </p>
      </div>
      <Suspense fallback={<div className="h-96 animate-pulse rounded-lg bg-muted" />}>
        <TraceDiffViewer leftTraceId={left} rightTraceId={right} />
      </Suspense>
    </div>
  );
}
