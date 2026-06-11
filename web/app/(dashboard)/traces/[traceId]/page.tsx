import { Suspense } from 'react';
import { notFound } from 'next/navigation';
import { TraceDetail } from '@/components/traces/trace-detail';
import { TraceDetailSkeleton } from '@/components/traces/trace-detail-skeleton';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Play } from 'lucide-react';
import Link from 'next/link';

export const metadata = {
  title: 'Trace Detail | AgentTrace',
  description: 'View detailed trace information',
};

interface TraceDetailPageProps {
  params: Promise<{
    traceId: string;
  }>;
}

export default async function TraceDetailPage({ params }: TraceDetailPageProps) {
  const { traceId } = await params;

  if (!traceId) {
    notFound();
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/traces">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Traces
          </Link>
        </Button>
        <Button variant="outline" size="sm" asChild>
          <Link href={`/replay?traceId=${encodeURIComponent(traceId)}`}>
            <Play className="mr-2 h-4 w-4" />
            Open replay debugger
          </Link>
        </Button>
      </div>
      <Suspense fallback={<TraceDetailSkeleton />}>
        <TraceDetail traceId={traceId} />
      </Suspense>
    </div>
  );
}
