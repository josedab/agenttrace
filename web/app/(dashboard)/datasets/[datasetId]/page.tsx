import { Suspense } from 'react';
import { notFound } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { DatasetDetail } from '@/components/datasets/dataset-detail';
import { Skeleton } from '@/components/ui/skeleton';

export const metadata = {
  title: 'Dataset Detail | AgentTrace',
  description: 'View and manage dataset items',
};

interface DatasetDetailPageProps {
  params: Promise<{
    datasetId: string;
  }>;
}

export default async function DatasetDetailPage({ params }: DatasetDetailPageProps) {
  const { datasetId } = await params;

  if (!datasetId) {
    notFound();
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/datasets">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Datasets
          </Link>
        </Button>
      </div>
      <Suspense fallback={<DatasetDetailSkeleton />}>
        <DatasetDetail datasetId={datasetId} />
      </Suspense>
    </div>
  );
}

function DatasetDetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <Skeleton className="h-8 w-48" />
          <Skeleton className="mt-2 h-4 w-64" />
        </div>
        <div className="flex gap-2">
          <Skeleton className="h-10 w-32" />
          <Skeleton className="h-10 w-32" />
        </div>
      </div>
      <Skeleton className="h-96 w-full" />
    </div>
  );
}
