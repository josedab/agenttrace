import { Suspense } from "react";
import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DatasetRunResults } from "@/components/datasets/dataset-run-results";
import { Skeleton } from "@/components/ui/skeleton";

export const metadata = {
  title: "Experiment Run | AgentTrace",
  description: "View experiment run results",
};

interface RunDetailPageProps {
  params: {
    datasetId: string;
    runId: string;
  };
}

export default function RunDetailPage({ params }: RunDetailPageProps) {
  const { datasetId, runId } = params;

  if (!datasetId || !runId) {
    notFound();
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href={`/datasets/${datasetId}`}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Dataset
          </Link>
        </Button>
      </div>
      <Suspense fallback={<RunDetailSkeleton />}>
        <DatasetRunResults datasetId={datasetId} runId={runId} />
      </Suspense>
    </div>
  );
}

function RunDetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-4 w-64 mt-2" />
        </div>
      </div>
      <div className="grid grid-cols-4 gap-4">
        {[...Array(4)].map((_, i) => (
          <Skeleton key={i} className="h-24 w-full" />
        ))}
      </div>
      <Skeleton className="h-96 w-full" />
    </div>
  );
}
