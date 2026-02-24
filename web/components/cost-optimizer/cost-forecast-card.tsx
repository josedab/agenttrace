"use client";

import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { CostForecast } from "@/hooks/use-cost-optimizer";
import { TrendingDown, DollarSign } from "lucide-react";

interface CostForecastCardProps {
  forecast: CostForecast;
}

export function CostForecastCard({ forecast }: CostForecastCardProps) {
  const formatCurrency = (amount: number) => `$${amount.toFixed(2)}`;

  const statusColors: Record<string, string> = {
    within: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
    warning: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400",
    exceeded: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400",
  };

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center justify-between text-sm">
          <span className="flex items-center gap-1">
            <DollarSign className="h-4 w-4" /> Cost Forecast
          </span>
          <Badge className={cn("text-xs", statusColors[forecast.budgetStatus] || statusColors.within)}>
            {forecast.budgetStatus === "within" ? "On Budget" : forecast.budgetStatus === "warning" ? "Warning" : "Over Budget"}
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 gap-4 mb-4">
          <div>
            <p className="text-xs text-muted-foreground">Current Daily</p>
            <p className="text-xl font-bold">{formatCurrency(forecast.currentDailyCost)}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Projected Monthly</p>
            <p className="text-xl font-bold">{formatCurrency(forecast.projectedMonthlyCost)}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Projected Yearly</p>
            <p className="text-lg font-semibold">{formatCurrency(forecast.projectedYearlyCost)}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Savings Potential</p>
            <p className="text-lg font-semibold flex items-center gap-1">
              <TrendingDown className="h-4 w-4 text-green-500" />
              {forecast.optimizationPotential.toFixed(0)}%
            </p>
          </div>
        </div>
        <div className="pt-3 border-t">
          <p className="text-xs text-muted-foreground mb-1">95% Confidence Range (Monthly)</p>
          <div className="flex items-center gap-2 text-sm">
            <span>{formatCurrency(forecast.confidenceInterval[0])}</span>
            <div className="flex-1 h-1.5 bg-muted rounded-full overflow-hidden">
              <div className="h-full bg-blue-500 rounded-full" style={{ width: "60%" }} />
            </div>
            <span>{formatCurrency(forecast.confidenceInterval[1])}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
