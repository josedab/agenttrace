"use client";

import * as React from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { Play, FileUp, Plus, AlertCircle } from "lucide-react";

import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { DatasetItemTable } from "@/components/datasets/dataset-item-table";
import { DatasetRunList } from "@/components/datasets/dataset-run-list";
import { AddItemDialog } from "@/components/datasets/add-item-dialog";
import { ImportItemsDialog } from "@/components/datasets/import-items-dialog";
import { RunExperimentDialog } from "@/components/datasets/run-experiment-dialog";

interface DatasetDetailProps {
  datasetId: string;
}

export function DatasetDetail({ datasetId }: DatasetDetailProps) {
  const { data: dataset, isLoading, error } = useQuery({
    queryKey: ["dataset", datasetId],
    queryFn: () => api.datasets.get(datasetId),
  });

  if (isLoading) {
    return <DatasetDetailSkeleton />;
  }

  if (error || !dataset) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <AlertCircle className="h-12 w-12 text-destructive mb-4" />
        <p className="text-destructive">Failed to load dataset</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold">{dataset.name}</h1>
          {dataset.description && (
            <p className="text-muted-foreground mt-1">{dataset.description}</p>
          )}
          <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
            <span>
              <Badge variant="secondary">{dataset.itemCount || 0}</Badge> items
            </span>
            <span>
              <Badge variant="outline">{dataset.runCount || 0}</Badge> runs
            </span>
            <span>
              Updated {formatDistanceToNow(new Date(dataset.updatedAt), { addSuffix: true })}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <ImportItemsDialog datasetId={datasetId} />
          <AddItemDialog datasetId={datasetId} />
          <RunExperimentDialog datasetId={datasetId} datasetName={dataset.name} />
        </div>
      </div>

      {/* Content */}
      <Tabs defaultValue="items">
        <TabsList>
          <TabsTrigger value="items">Items</TabsTrigger>
          <TabsTrigger value="runs">Experiment Runs</TabsTrigger>
        </TabsList>

        <TabsContent value="items" className="mt-4">
          <DatasetItemTable datasetId={datasetId} />
        </TabsContent>

        <TabsContent value="runs" className="mt-4">
          <DatasetRunList datasetId={datasetId} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function DatasetDetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <div className="h-8 w-48 bg-muted animate-pulse rounded" />
          <div className="h-4 w-64 bg-muted animate-pulse rounded mt-2" />
        </div>
        <div className="flex gap-2">
          <div className="h-10 w-32 bg-muted animate-pulse rounded" />
          <div className="h-10 w-32 bg-muted animate-pulse rounded" />
        </div>
      </div>
      <div className="h-96 bg-muted animate-pulse rounded" />
    </div>
  );
}
