"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import { formatDistanceToNow, format } from "date-fns";
import { CheckCircle, XCircle, Clock, AlertCircle, BarChart3 } from "lucide-react";

import { cn } from "@/lib/utils";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";

interface DatasetRunResultsProps {
  datasetId: string;
  runId: string;
}

export function DatasetRunResults({ datasetId, runId }: DatasetRunResultsProps) {
  const { data: run, isLoading, error } = useQuery({
    queryKey: ["dataset-run", datasetId, runId],
    queryFn: () => api.datasets.getRun(datasetId, runId),
    refetchInterval: (query) => {
      // Keep polling while running
      if (query.state.data?.status === "RUNNING") {
        return 3000;
      }
      return false;
    },
  });

  if (isLoading) {
    return <RunResultsSkeleton />;
  }

  if (error || !run) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <AlertCircle className="h-12 w-12 text-destructive mb-4" />
        <p className="text-destructive">Failed to load run results</p>
      </div>
    );
  }

  const progress = (run.completedCount / run.totalCount) * 100;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold">{run.name}</h1>
          <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
            <span>Started {formatDistanceToNow(new Date(run.createdAt), { addSuffix: true })}</span>
            <RunStatusBadge status={run.status} />
          </div>
        </div>
      </div>

      {/* Progress (if running) */}
      {run.status === "RUNNING" && (
        <Card>
          <CardContent className="pt-6">
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>Progress</span>
                <span>{run.completedCount} / {run.totalCount} items</span>
              </div>
              <Progress value={progress} />
            </div>
          </CardContent>
        </Card>
      )}

      {/* Summary stats */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Items</CardDescription>
          </CardHeader>
          <CardContent>
            <span className="text-2xl font-bold">{run.totalCount}</span>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Completed</CardDescription>
          </CardHeader>
          <CardContent>
            <span className="text-2xl font-bold text-green-500">{run.completedCount}</span>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Failed</CardDescription>
          </CardHeader>
          <CardContent>
            <span className="text-2xl font-bold text-red-500">{run.failedCount || 0}</span>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Avg Score</CardDescription>
          </CardHeader>
          <CardContent>
            <span className="text-2xl font-bold">
              {run.avgScore !== null ? run.avgScore.toFixed(2) : "-"}
            </span>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Cost</CardDescription>
          </CardHeader>
          <CardContent>
            <span className="text-2xl font-bold">
              ${run.totalCost?.toFixed(4) || "0.00"}
            </span>
          </CardContent>
        </Card>
      </div>

      {/* Score distribution */}
      {run.scores && run.scores.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              Score Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ScoreDistribution scores={run.scores} />
          </CardContent>
        </Card>
      )}

      {/* Results table */}
      <Card>
        <CardHeader>
          <CardTitle>Results</CardTitle>
          <CardDescription>
            Individual item results
          </CardDescription>
        </CardHeader>
        <CardContent>
          {run.results && run.results.length > 0 ? (
            <div className="border rounded-lg">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[50px]">#</TableHead>
                    <TableHead>Input</TableHead>
                    <TableHead>Output</TableHead>
                    <TableHead>Expected</TableHead>
                    <TableHead className="w-[100px]">Score</TableHead>
                    <TableHead className="w-[100px]">Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {run.results.map((result, index) => (
                    <TableRow key={result.id}>
                      <TableCell className="text-muted-foreground">
                        {index + 1}
                      </TableCell>
                      <TableCell className="max-w-[200px]">
                        <p className="truncate text-sm">
                          {truncateValue(result.input, 50)}
                        </p>
                      </TableCell>
                      <TableCell className="max-w-[200px]">
                        <p className="truncate text-sm">
                          {truncateValue(result.output, 50)}
                        </p>
                      </TableCell>
                      <TableCell className="max-w-[200px]">
                        <p className="truncate text-sm text-muted-foreground">
                          {result.expectedOutput
                            ? truncateValue(result.expectedOutput, 50)
                            : "-"}
                        </p>
                      </TableCell>
                      <TableCell>
                        {result.score !== null ? (
                          <Badge
                            variant={result.score >= 0.7 ? "default" : result.score >= 0.4 ? "secondary" : "destructive"}
                          >
                            {result.score.toFixed(2)}
                          </Badge>
                        ) : (
                          "-"
                        )}
                      </TableCell>
                      <TableCell>
                        <ResultStatusBadge status={result.status} />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground text-center py-8">
              No results yet
            </p>
          )}
        </CardContent>
      </Card>
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

function ResultStatusBadge({ status }: { status: string }) {
  switch (status) {
    case "SUCCESS":
      return (
        <div className="flex items-center gap-1 text-green-500">
          <CheckCircle className="h-4 w-4" />
        </div>
      );
    case "ERROR":
      return (
        <div className="flex items-center gap-1 text-red-500">
          <XCircle className="h-4 w-4" />
        </div>
      );
    case "PENDING":
      return (
        <div className="flex items-center gap-1 text-muted-foreground">
          <Clock className="h-4 w-4" />
        </div>
      );
    default:
      return <span>{status}</span>;
  }
}

function ScoreDistribution({ scores }: { scores: number[] }) {
  const bins = [0, 0, 0, 0, 0]; // 0-0.2, 0.2-0.4, 0.4-0.6, 0.6-0.8, 0.8-1.0
  scores.forEach((score) => {
    const index = Math.min(Math.floor(score * 5), 4);
    bins[index]++;
  });

  const maxCount = Math.max(...bins);

  return (
    <div className="flex items-end gap-2 h-32">
      {bins.map((count, index) => (
        <div key={index} className="flex-1 flex flex-col items-center gap-1">
          <div
            className={cn(
              "w-full rounded-t",
              index < 2 ? "bg-red-500" : index < 4 ? "bg-yellow-500" : "bg-green-500"
            )}
            style={{
              height: `${(count / maxCount) * 100}%`,
              minHeight: count > 0 ? "4px" : 0,
            }}
          />
          <span className="text-xs text-muted-foreground">
            {(index * 0.2).toFixed(1)}-{((index + 1) * 0.2).toFixed(1)}
          </span>
        </div>
      ))}
    </div>
  );
}

function truncateValue(value: any, length: number): string {
  const str = typeof value === "string" ? value : JSON.stringify(value);
  if (str.length <= length) return str;
  return str.slice(0, length) + "...";
}

function RunResultsSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-64 mt-2" />
      </div>
      <div className="grid grid-cols-5 gap-4">
        {[...Array(5)].map((_, i) => (
          <Skeleton key={i} className="h-24 w-full" />
        ))}
      </div>
      <Skeleton className="h-96 w-full" />
    </div>
  );
}
