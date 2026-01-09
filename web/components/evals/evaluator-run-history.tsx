"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import { formatDistanceToNow, format } from "date-fns";
import { CheckCircle, XCircle, Clock, Loader2 } from "lucide-react";

import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

interface EvaluatorRun {
  id: string;
  status: "PENDING" | "RUNNING" | "COMPLETED" | "FAILED";
  startedAt: string;
  completedAt?: string;
  totalCount: number;
  completedCount: number;
  failedCount: number;
  avgScore?: number;
  error?: string;
}

interface EvaluatorRunHistoryProps {
  evaluatorId: string;
}

export function EvaluatorRunHistory({ evaluatorId }: EvaluatorRunHistoryProps) {
  const { data: runs, isLoading, error } = useQuery({
    queryKey: ["evaluator-runs", evaluatorId],
    queryFn: () => api.evaluators.listRuns(evaluatorId),
    refetchInterval: (query) => {
      // Keep polling if any run is in progress
      const hasRunning = query.state.data?.some(
        (run: EvaluatorRun) => run.status === "RUNNING" || run.status === "PENDING"
      );
      return hasRunning ? 3000 : false;
    },
  });

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Run History</CardTitle>
          <CardDescription>Recent evaluation runs</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {[...Array(5)].map((_, i) => (
              <Skeleton key={i} className="h-12 w-full" />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Run History</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-destructive">Failed to load run history</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Run History</CardTitle>
        <CardDescription>Recent evaluation runs and their results</CardDescription>
      </CardHeader>
      <CardContent>
        {runs && runs.length > 0 ? (
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Status</TableHead>
                  <TableHead>Started</TableHead>
                  <TableHead>Duration</TableHead>
                  <TableHead>Progress</TableHead>
                  <TableHead>Avg Score</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {runs.map((run: EvaluatorRun) => (
                  <TableRow key={run.id}>
                    <TableCell>
                      <RunStatusBadge status={run.status} />
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">
                        {formatDistanceToNow(new Date(run.startedAt), {
                          addSuffix: true,
                        })}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {format(new Date(run.startedAt), "MMM d, HH:mm")}
                      </div>
                    </TableCell>
                    <TableCell>
                      {run.completedAt ? (
                        <span className="text-sm">
                          {Math.round(
                            (new Date(run.completedAt).getTime() -
                              new Date(run.startedAt).getTime()) /
                              1000
                          )}
                          s
                        </span>
                      ) : run.status === "RUNNING" ? (
                        <span className="text-sm text-muted-foreground">
                          In progress...
                        </span>
                      ) : (
                        <span className="text-sm text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <span className="text-sm">
                          {run.completedCount}/{run.totalCount}
                        </span>
                        {run.failedCount > 0 && (
                          <span className="text-xs text-destructive">
                            ({run.failedCount} failed)
                          </span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      {run.avgScore !== undefined ? (
                        <Badge
                          variant={
                            run.avgScore >= 0.7
                              ? "default"
                              : run.avgScore >= 0.4
                              ? "secondary"
                              : "destructive"
                          }
                        >
                          {run.avgScore.toFixed(2)}
                        </Badge>
                      ) : (
                        <span className="text-sm text-muted-foreground">-</span>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground text-center py-8">
            No runs yet. Click "Run Now" to start an evaluation.
          </p>
        )}
      </CardContent>
    </Card>
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
