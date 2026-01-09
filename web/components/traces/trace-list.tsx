"use client";

import * as React from "react";
import Link from "next/link";
import { useInfiniteQuery } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { Clock, DollarSign, Layers, AlertCircle, ChevronRight } from "lucide-react";

import { cn } from "@/lib/utils";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

interface TraceListProps {
  searchParams: {
    q?: string;
    level?: string;
    startDate?: string;
    endDate?: string;
    minLatency?: string;
    maxLatency?: string;
    minCost?: string;
    maxCost?: string;
  };
}

export function TraceList({ searchParams }: TraceListProps) {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
    error,
  } = useInfiniteQuery({
    queryKey: ["traces", searchParams],
    queryFn: async ({ pageParam }) => {
      const result = await api.traces.list({
        cursor: pageParam,
        limit: 50,
        search: searchParams.q,
        level: searchParams.level,
        startDate: searchParams.startDate,
        endDate: searchParams.endDate,
        minLatency: searchParams.minLatency ? parseInt(searchParams.minLatency) : undefined,
        maxLatency: searchParams.maxLatency ? parseInt(searchParams.maxLatency) : undefined,
        minCost: searchParams.minCost ? parseFloat(searchParams.minCost) : undefined,
        maxCost: searchParams.maxCost ? parseFloat(searchParams.maxCost) : undefined,
      });
      return result;
    },
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    initialPageParam: undefined as string | undefined,
  });

  const traces = data?.pages.flatMap((page) => page.data) ?? [];

  if (isLoading) {
    return <TraceListLoading />;
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <AlertCircle className="h-12 w-12 text-destructive mb-4" />
        <p className="text-destructive">Failed to load traces</p>
      </div>
    );
  }

  if (traces.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 border rounded-lg bg-card">
        <Layers className="h-12 w-12 text-muted-foreground mb-4" />
        <p className="text-lg font-medium">No traces found</p>
        <p className="text-sm text-muted-foreground mt-1">
          Start instrumenting your agents to see traces here
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[300px]">Name / ID</TableHead>
              <TableHead>Input</TableHead>
              <TableHead>Output</TableHead>
              <TableHead className="w-[100px]">Level</TableHead>
              <TableHead className="w-[100px]">Latency</TableHead>
              <TableHead className="w-[100px]">Cost</TableHead>
              <TableHead className="w-[150px]">Time</TableHead>
              <TableHead className="w-[50px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {traces.map((trace) => (
              <TableRow key={trace.id}>
                <TableCell>
                  <Link
                    href={`/traces/${trace.id}`}
                    className="font-medium hover:underline"
                  >
                    {trace.name || trace.id.slice(0, 8)}
                  </Link>
                  <p className="text-xs text-muted-foreground font-mono">
                    {trace.id.slice(0, 16)}...
                  </p>
                </TableCell>
                <TableCell className="max-w-[200px]">
                  <p className="truncate text-sm text-muted-foreground">
                    {truncate(trace.input || "-", 50)}
                  </p>
                </TableCell>
                <TableCell className="max-w-[200px]">
                  <p className="truncate text-sm text-muted-foreground">
                    {truncate(trace.output || "-", 50)}
                  </p>
                </TableCell>
                <TableCell>
                  <LevelBadge level={trace.level} />
                </TableCell>
                <TableCell>
                  {trace.latency ? (
                    <div className="flex items-center gap-1 text-sm">
                      <Clock className="h-3 w-3 text-muted-foreground" />
                      {formatLatency(trace.latency)}
                    </div>
                  ) : (
                    "-"
                  )}
                </TableCell>
                <TableCell>
                  {trace.totalCost !== null && trace.totalCost !== undefined ? (
                    <div className="flex items-center gap-1 text-sm">
                      <DollarSign className="h-3 w-3 text-muted-foreground" />
                      {trace.totalCost.toFixed(4)}
                    </div>
                  ) : (
                    "-"
                  )}
                </TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {formatDistanceToNow(new Date(trace.startTime), {
                    addSuffix: true,
                  })}
                </TableCell>
                <TableCell>
                  <Link href={`/traces/${trace.id}`}>
                    <ChevronRight className="h-4 w-4 text-muted-foreground" />
                  </Link>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {hasNextPage && (
        <div className="flex justify-center">
          <Button
            variant="outline"
            onClick={() => fetchNextPage()}
            disabled={isFetchingNextPage}
          >
            {isFetchingNextPage ? "Loading..." : "Load more"}
          </Button>
        </div>
      )}
    </div>
  );
}

function TraceListLoading() {
  return (
    <div className="border rounded-lg">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[300px]">Name / ID</TableHead>
            <TableHead>Input</TableHead>
            <TableHead>Output</TableHead>
            <TableHead className="w-[100px]">Level</TableHead>
            <TableHead className="w-[100px]">Latency</TableHead>
            <TableHead className="w-[100px]">Cost</TableHead>
            <TableHead className="w-[150px]">Time</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {[...Array(10)].map((_, i) => (
            <TableRow key={i}>
              <TableCell>
                <div className="h-4 w-32 bg-muted animate-pulse rounded" />
                <div className="h-3 w-24 bg-muted animate-pulse rounded mt-1" />
              </TableCell>
              <TableCell>
                <div className="h-4 w-40 bg-muted animate-pulse rounded" />
              </TableCell>
              <TableCell>
                <div className="h-4 w-40 bg-muted animate-pulse rounded" />
              </TableCell>
              <TableCell>
                <div className="h-5 w-16 bg-muted animate-pulse rounded" />
              </TableCell>
              <TableCell>
                <div className="h-4 w-16 bg-muted animate-pulse rounded" />
              </TableCell>
              <TableCell>
                <div className="h-4 w-16 bg-muted animate-pulse rounded" />
              </TableCell>
              <TableCell>
                <div className="h-4 w-24 bg-muted animate-pulse rounded" />
              </TableCell>
              <TableCell></TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function LevelBadge({ level }: { level: string }) {
  return (
    <Badge
      variant="outline"
      className={cn(
        "text-xs",
        level === "ERROR" && "border-destructive text-destructive",
        level === "WARNING" && "border-yellow-500 text-yellow-500",
        level === "DEBUG" && "border-blue-500 text-blue-500"
      )}
    >
      {level}
    </Badge>
  );
}

function truncate(str: string, length: number): string {
  if (str.length <= length) return str;
  return str.slice(0, length) + "...";
}

function formatLatency(ms: number): string {
  if (ms >= 1000) {
    return (ms / 1000).toFixed(2) + "s";
  }
  return ms.toFixed(0) + "ms";
}
