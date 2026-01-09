import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { DatasetList } from "@/components/datasets/dataset-list";
import { DatasetListSkeleton } from "@/components/datasets/dataset-list-skeleton";
import { CreateDatasetDialog } from "@/components/datasets/create-dataset-dialog";

export const metadata = {
  title: "Datasets | AgentTrace",
  description: "Manage datasets for evaluation experiments",
};

export default function DatasetsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Datasets"
        description="Manage datasets for evaluation experiments"
      >
        <CreateDatasetDialog />
      </PageHeader>
      <Suspense fallback={<DatasetListSkeleton />}>
        <DatasetList />
      </Suspense>
    </div>
  );
}
