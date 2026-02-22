"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  TrendingDown,
  DollarSign,
  ArrowRight,
  Check,
  X,
  Sparkles,
} from "lucide-react";

interface CostRecommendation {
  id: string;
  currentModel: string;
  recommendedModel: string;
  traceCount: number;
  estimatedSavingsPerMonth: number;
  qualityImpactEstimate: number;
  confidence: number;
  status: string;
}

interface CostAnalysis {
  projectId: string;
  totalCostPeriod: number;
  modelBreakdown: {
    model: string;
    traceCount: number;
    totalCost: number;
    avgCostPerTrace: number;
  }[];
  recommendations: CostRecommendation[];
  potentialSavings: number;
}

export function CostOptimizerPanel() {
  const queryClient = useQueryClient();

  const { data: analysis, isLoading } = useQuery<CostAnalysis>({
    queryKey: ["cost-optimizer"],
    queryFn: () => api.costOptimizer.analyze({ dateRange: "30d" }),
  });

  const applyMutation = useMutation({
    mutationFn: (id: string) => api.costOptimizer.apply(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["cost-optimizer"] });
    },
  });

  const dismissMutation = useMutation({
    mutationFn: (id: string) => api.costOptimizer.dismiss(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["cost-optimizer"] });
    },
  });

  if (isLoading || !analysis) {
    return <CostOptimizerSkeleton />;
  }

  const activeRecommendations = analysis.recommendations.filter(
    (r) => r.status === "PENDING"
  );

  return (
    <div className="space-y-6">
      {/* Savings Summary */}
      {analysis.potentialSavings > 0 && (
        <Card className="border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-950">
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-green-100 dark:bg-green-900 rounded-full">
                <Sparkles className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <div className="text-sm text-green-700 dark:text-green-300">
                  Potential Monthly Savings
                </div>
                <div className="text-2xl font-bold text-green-700 dark:text-green-300">
                  ${analysis.potentialSavings.toFixed(2)}
                </div>
              </div>
              <div className="ml-auto text-sm text-green-600 dark:text-green-400">
                {activeRecommendations.length} recommendation
                {activeRecommendations.length !== 1 ? "s" : ""} available
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Model Cost Breakdown */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <DollarSign className="h-4 w-4" />
            Cost by Model (Last 30 days)
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Model</TableHead>
                <TableHead className="text-right">Traces</TableHead>
                <TableHead className="text-right">Total Cost</TableHead>
                <TableHead className="text-right">Avg/Trace</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {analysis.modelBreakdown.map((entry) => (
                <TableRow key={entry.model}>
                  <TableCell className="font-mono text-sm">
                    {entry.model}
                  </TableCell>
                  <TableCell className="text-right">
                    {entry.traceCount.toLocaleString()}
                  </TableCell>
                  <TableCell className="text-right font-medium">
                    ${entry.totalCost.toFixed(2)}
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground">
                    ${entry.avgCostPerTrace.toFixed(4)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Recommendations */}
      {activeRecommendations.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <TrendingDown className="h-4 w-4" />
              Optimization Recommendations
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {activeRecommendations.map((rec) => (
              <div
                key={rec.id}
                className="flex items-center gap-4 p-3 border rounded-lg"
              >
                <div className="flex items-center gap-2 min-w-0 flex-1">
                  <span className="font-mono text-sm truncate">
                    {rec.currentModel}
                  </span>
                  <ArrowRight className="h-4 w-4 shrink-0 text-muted-foreground" />
                  <span className="font-mono text-sm truncate text-green-600">
                    {rec.recommendedModel}
                  </span>
                </div>
                <div className="text-right shrink-0">
                  <div className="text-sm font-medium text-green-600">
                    -${rec.estimatedSavingsPerMonth.toFixed(2)}/mo
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {rec.traceCount} traces · {(rec.confidence * 100).toFixed(0)}%
                    confidence
                  </div>
                </div>
                <div className="flex gap-1 shrink-0">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => applyMutation.mutate(rec.id)}
                    disabled={applyMutation.isPending}
                  >
                    <Check className="h-3 w-3" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => dismissMutation.mutate(rec.id)}
                    disabled={dismissMutation.isPending}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function CostOptimizerSkeleton() {
  return (
    <div className="space-y-4">
      <div className="h-24 bg-muted animate-pulse rounded-lg" />
      <div className="h-64 bg-muted animate-pulse rounded-lg" />
    </div>
  );
}
