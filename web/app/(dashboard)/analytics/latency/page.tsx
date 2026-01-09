"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import { Clock, TrendingUp, TrendingDown, Calendar, Gauge } from "lucide-react";

import { api } from "@/lib/api";
import { cn } from "@/lib/utils";
import { PageHeader } from "@/components/layout/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { LatencyPercentileChart } from "@/components/dashboard/latency-percentile-chart";
import { LatencyDistributionChart } from "@/components/analytics/latency-distribution-chart";

export default function LatencyAnalyticsPage() {
  const [dateRange, setDateRange] = React.useState("7d");
  const [groupBy, setGroupBy] = React.useState("model");

  const { data: latencyData, isLoading } = useQuery({
    queryKey: ["latency-analytics", dateRange, groupBy],
    queryFn: () => api.analytics.getLatencyAnalytics({ dateRange, groupBy }),
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Latency Analytics"
        description="Monitor and optimize your LLM response times."
      />

      {/* Filters */}
      <div className="flex items-center gap-4">
        <Select value={dateRange} onValueChange={setDateRange}>
          <SelectTrigger className="w-[150px]">
            <Calendar className="h-4 w-4 mr-2" />
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="24h">Last 24 hours</SelectItem>
            <SelectItem value="7d">Last 7 days</SelectItem>
            <SelectItem value="30d">Last 30 days</SelectItem>
            <SelectItem value="90d">Last 90 days</SelectItem>
          </SelectContent>
        </Select>

        <Select value={groupBy} onValueChange={setGroupBy}>
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="Group by" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="model">By Model</SelectItem>
            <SelectItem value="project">By Project</SelectItem>
            <SelectItem value="endpoint">By Endpoint</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>P50 Latency</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-bold">
                  {latencyData?.p50?.toFixed(0) ?? 0}ms
                </span>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>P95 Latency</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <span className="text-2xl font-bold">
                {latencyData?.p95?.toFixed(0) ?? 0}ms
              </span>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>P99 Latency</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <span className="text-2xl font-bold text-orange-500">
                {latencyData?.p99?.toFixed(0) ?? 0}ms
              </span>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Average</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-bold">
                  {latencyData?.avg?.toFixed(0) ?? 0}ms
                </span>
                {latencyData?.avgChange !== undefined && (
                  <span
                    className={cn(
                      "text-xs flex items-center",
                      latencyData.avgChange <= 0
                        ? "text-green-500"
                        : "text-red-500"
                    )}
                  >
                    {latencyData.avgChange <= 0 ? (
                      <TrendingDown className="h-3 w-3 mr-0.5" />
                    ) : (
                      <TrendingUp className="h-3 w-3 mr-0.5" />
                    )}
                    {Math.abs(latencyData.avgChange).toFixed(1)}%
                  </span>
                )}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Requests</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <span className="text-2xl font-bold">
                {latencyData?.totalRequests?.toLocaleString() ?? 0}
              </span>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Latency Over Time</CardTitle>
            <CardDescription>P50, P95, P99 percentiles</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-64 w-full" />
            ) : (
              <LatencyPercentileChart data={latencyData?.latencyOverTime || []} />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Latency Distribution</CardTitle>
            <CardDescription>Request latency histogram</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-64 w-full" />
            ) : (
              <LatencyDistributionChart data={latencyData?.distribution || []} />
            )}
          </CardContent>
        </Card>
      </div>

      {/* Detailed breakdown table */}
      <Card>
        <CardHeader>
          <CardTitle>
            Latency by {groupBy === "model" ? "Model" : groupBy === "project" ? "Project" : "Endpoint"}
          </CardTitle>
          <CardDescription>Detailed latency breakdown</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-2">
              {[...Array(5)].map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : (
            <div className="border rounded-lg">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>
                      {groupBy === "model" ? "Model" : groupBy === "project" ? "Project" : "Endpoint"}
                    </TableHead>
                    <TableHead className="text-right">Requests</TableHead>
                    <TableHead className="text-right">P50</TableHead>
                    <TableHead className="text-right">P95</TableHead>
                    <TableHead className="text-right">P99</TableHead>
                    <TableHead className="text-right">Average</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {latencyData?.breakdown?.map((item: any) => (
                    <TableRow key={item.name}>
                      <TableCell className="font-medium">{item.name}</TableCell>
                      <TableCell className="text-right">
                        {item.requests?.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        {item.p50?.toFixed(0)}ms
                      </TableCell>
                      <TableCell className="text-right">
                        {item.p95?.toFixed(0)}ms
                      </TableCell>
                      <TableCell className="text-right text-orange-500">
                        {item.p99?.toFixed(0)}ms
                      </TableCell>
                      <TableCell className="text-right text-muted-foreground">
                        {item.avg?.toFixed(0)}ms
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
