"use client";

import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface QualityScoreCardProps {
  score: number;
  dimensionScores: Record<string, number>;
  trend?: string;
}

const dimensionLabels: Record<string, string> = {
  security: "Security",
  complexity: "Complexity",
  readability: "Readability",
  maintainability: "Maintainability",
  test_coverage: "Test Coverage",
  performance: "Performance",
};

function getScoreColor(score: number): string {
  if (score >= 80) return "text-green-600 dark:text-green-400";
  if (score >= 60) return "text-yellow-600 dark:text-yellow-400";
  if (score >= 40) return "text-orange-600 dark:text-orange-400";
  return "text-red-600 dark:text-red-400";
}

function getScoreBar(score: number): string {
  if (score >= 80) return "bg-green-500";
  if (score >= 60) return "bg-yellow-500";
  if (score >= 40) return "bg-orange-500";
  return "bg-red-500";
}

export function QualityScoreCard({
  score,
  dimensionScores,
  trend,
}: QualityScoreCardProps) {
  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center justify-between">
          <span>Code Quality</span>
          {trend && (
            <span className="text-xs font-normal text-muted-foreground">
              {trend === "improving"
                ? "📈 Improving"
                : trend === "declining"
                  ? "📉 Declining"
                  : "→ Stable"}
            </span>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-center mb-6">
          <div className="relative h-28 w-28">
            <svg
              className="h-28 w-28 -rotate-90"
              viewBox="0 0 100 100"
            >
              <circle
                cx="50"
                cy="50"
                r="45"
                fill="none"
                stroke="currentColor"
                strokeWidth="8"
                className="text-muted/20"
              />
              <circle
                cx="50"
                cy="50"
                r="45"
                fill="none"
                strokeWidth="8"
                strokeDasharray={`${(score / 100) * 283} 283`}
                strokeLinecap="round"
                className={getScoreBar(score)}
              />
            </svg>
            <div className="absolute inset-0 flex items-center justify-center">
              <span
                className={cn("text-3xl font-bold", getScoreColor(score))}
              >
                {Math.round(score)}
              </span>
            </div>
          </div>
        </div>

        <div className="space-y-3">
          {Object.entries(dimensionScores).map(([key, value]) => (
            <div key={key}>
              <div className="flex justify-between text-sm mb-1">
                <span className="text-muted-foreground">
                  {dimensionLabels[key] || key}
                </span>
                <span className={cn("font-medium", getScoreColor(value))}>
                  {Math.round(value)}
                </span>
              </div>
              <div className="h-1.5 bg-muted rounded-full overflow-hidden">
                <div
                  className={cn(
                    "h-full rounded-full transition-all",
                    getScoreBar(value),
                  )}
                  style={{ width: `${value}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
