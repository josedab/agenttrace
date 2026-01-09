"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import { format, subDays } from "date-fns";
import { DollarSign, TrendingUp, TrendingDown, Calendar } from "lucide-react";

import { api } from "@/lib/api";
import { cn } from "@/lib/utils";
import { PageHeader } from "@/components/layout/page-header";
import { Button } from "@/components/ui/button";
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
import { CostBreakdownChart } from "@/components/dashboard/cost-breakdown-chart";
import { CostTrendChart } from "@/components/analytics/cost-trend-chart";

export default function CostAnalyticsPage() {
  const [dateRange, setDateRange] = React.useState("7d");
  const [groupBy, setGroupBy] = React.useState("model");

  const { data: costData, isLoading } = useQuery({
    queryKey: ["cost-analytics", dateRange, groupBy],
    queryFn: () => api.analytics.getCostAnalytics({ dateRange, groupBy }),
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Cost Analytics"
        description="Track and analyze your LLM spending."
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
            <SelectItem value="user">By User</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Cost</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-bold">
                  ${costData?.totalCost?.toFixed(2) ?? "0.00"}
                </span>
                {costData?.costChange !== undefined && (
                  <span
                    className={cn(
                      "text-xs flex items-center",
                      costData.costChange >= 0 ? "text-red-500" : "text-green-500"
                    )}
                  >
                    {costData.costChange >= 0 ? (
                      <TrendingUp className="h-3 w-3 mr-0.5" />
                    ) : (
                      <TrendingDown className="h-3 w-3 mr-0.5" />
                    )}
                    {Math.abs(costData.costChange).toFixed(1)}%
                  </span>
                )}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Input Tokens Cost</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <span className="text-2xl font-bold">
                ${costData?.inputCost?.toFixed(2) ?? "0.00"}
              </span>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Output Tokens Cost</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <span className="text-2xl font-bold">
                ${costData?.outputCost?.toFixed(2) ?? "0.00"}
              </span>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Avg Cost per Trace</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-24" />
            ) : (
              <span className="text-2xl font-bold">
                ${costData?.avgCostPerTrace?.toFixed(4) ?? "0.0000"}
              </span>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Cost Over Time</CardTitle>
            <CardDescription>Daily cost breakdown</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-64 w-full" />
            ) : (
              <CostTrendChart data={costData?.costOverTime || []} />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Cost by {groupBy === "model" ? "Model" : groupBy === "project" ? "Project" : "User"}</CardTitle>
            <CardDescription>Cost distribution</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-64 w-full" />
            ) : (
              <CostBreakdownChart data={costData?.costByGroup || []} />
            )}
          </CardContent>
        </Card>
      </div>

      {/* Detailed breakdown table */}
      <Card>
        <CardHeader>
          <CardTitle>Detailed Breakdown</CardTitle>
          <CardDescription>
            Cost breakdown by {groupBy === "model" ? "model" : groupBy === "project" ? "project" : "user"}
          </CardDescription>
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
                      {groupBy === "model" ? "Model" : groupBy === "project" ? "Project" : "User"}
                    </TableHead>
                    <TableHead className="text-right">Traces</TableHead>
                    <TableHead className="text-right">Input Tokens</TableHead>
                    <TableHead className="text-right">Output Tokens</TableHead>
                    <TableHead className="text-right">Total Cost</TableHead>
                    <TableHead className="text-right">% of Total</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {costData?.breakdown?.map((item: any) => (
                    <TableRow key={item.name}>
                      <TableCell className="font-medium">{item.name}</TableCell>
                      <TableCell className="text-right">
                        {item.traces?.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        {item.inputTokens?.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        {item.outputTokens?.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        ${item.cost?.toFixed(4)}
                      </TableCell>
                      <TableCell className="text-right text-muted-foreground">
                        {item.percentage?.toFixed(1)}%
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
