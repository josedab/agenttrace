import { Suspense } from "react";
import { notFound } from "next/navigation";
import { PageHeader } from "@/components/layout/page-header";
import { TraceDetail } from "@/components/traces/trace-detail";
import { TraceDetailSkeleton } from "@/components/traces/trace-detail-skeleton";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";
import Link from "next/link";

export const metadata = {
  title: "Trace Detail | AgentTrace",
  description: "View detailed trace information",
};

interface TraceDetailPageProps {
  params: {
    traceId: string;
  };
}

export default function TraceDetailPage({ params }: TraceDetailPageProps) {
  const { traceId } = params;

  if (!traceId) {
    notFound();
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/traces">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Traces
          </Link>
        </Button>
      </div>
      <Suspense fallback={<TraceDetailSkeleton />}>
        <TraceDetail traceId={traceId} />
      </Suspense>
    </div>
  );
}
