"use client";

import * as React from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { Play, Clock, CheckCircle, XCircle, Loader2 } from "lucide-react";

import { cn } from "@/lib/utils";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";

interface DatasetRunListProps {
  datasetId: string;
}

export function DatasetRunList({ datasetId }: DatasetRunListProps) {
  const { data: runs, isLoading } = useQuery({
    queryKey: ["dataset-runs", datasetId],
    queryFn: () => api.datasets.getRuns(datasetId),
    refetchInterval: 5000, // Poll for updates
  });

  if (isLoading) {
    return <RunListSkeleton />;
  }

  if (!runs || runs.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 border rounded-lg bg-card">
        <Play className="h-12 w-12 text-muted-foreground mb-4" />
        <p className="text-lg font-medium">No experiment runs yet</p>
        <p className="text-sm text-muted-foreground mt-1">
          Run an experiment to see results here
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {runs.map((run) => (
        <Card key={run.id}>
          <CardHeader className="pb-3">
            <div className="flex items-start justify-between">
              <div>
                <CardTitle className="text-base">{run.name}</CardTitle>
                <CardDescription>
                  Started {formatDistanceToNow(new Date(run.createdAt), { addSuffix: true })}
                </CardDescription>
              </div>
              <RunStatusBadge status={run.status} />
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {/* Progress */}
              {run.status === "RUNNING" && (
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Progress</span>
                    <span>{run.completedCount} / {run.totalCount}</span>
                  </div>
                  <Progress value={(run.completedCount / run.totalCount) * 100} />
                </div>
              )}

              {/* Stats */}
              <div className="grid grid-cols-4 gap-4">
                <div>
                  <p className="text-sm text-muted-foreground">Total</p>
                  <p className="text-lg font-semibold">{run.totalCount}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Completed</p>
                  <p className="text-lg font-semibold text-green-500">{run.completedCount}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Failed</p>
                  <p className="text-lg font-semibold text-red-500">{run.failedCount || 0}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Avg Score</p>
                  <p className="text-lg font-semibold">
                    {run.avgScore !== null ? run.avgScore.toFixed(2) : "-"}
                  </p>
                </div>
              </div>

              {/* Actions */}
              <div className="flex justify-end">
                <Button variant="outline" size="sm" asChild>
                  <Link href={`/datasets/${datasetId}/runs/${run.id}`}>
                    View Results
                  </Link>
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function RunStatusBadge({ status }: { status: string }) {
  switch (status) {
    case "PENDING":
      return (
        <Badge variant="outline">
          <Clock className="h-3 w-3 mr-1" />
          Pending
        </Badge>
      );
    case "RUNNING":
      return (
        <Badge variant="secondary">
          <Loader2 className="h-3 w-3 mr-1 animate-spin" />
          Running
        </Badge>
      );
    case "COMPLETED":
      return (
        <Badge variant="default" className="bg-green-500">
          <CheckCircle className="h-3 w-3 mr-1" />
          Completed
        </Badge>
      );
    case "FAILED":
      return (
        <Badge variant="destructive">
          <XCircle className="h-3 w-3 mr-1" />
          Failed
        </Badge>
      );
    default:
      return <Badge variant="outline">{status}</Badge>;
  }
}

function RunListSkeleton() {
  return (
    <div className="space-y-4">
      {[...Array(3)].map((_, i) => (
        <Card key={i}>
          <CardHeader className="pb-3">
            <Skeleton className="h-5 w-48" />
            <Skeleton className="h-4 w-32" />
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-4 gap-4">
              {[...Array(4)].map((_, j) => (
                <div key={j}>
                  <Skeleton className="h-4 w-16 mb-1" />
                  <Skeleton className="h-6 w-12" />
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
