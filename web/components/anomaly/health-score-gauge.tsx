"use client";

import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface HealthScoreGaugeProps {
  score: number;
  activeAlerts: number;
}

export function HealthScoreGauge({ score, activeAlerts }: HealthScoreGaugeProps) {
  const getColor = (s: number) => {
    if (s >= 90) return { text: "text-green-500", bar: "text-green-500", label: "Healthy" };
    if (s >= 70) return { text: "text-yellow-500", bar: "text-yellow-500", label: "Warning" };
    if (s >= 50) return { text: "text-orange-500", bar: "text-orange-500", label: "Degraded" };
    return { text: "text-red-500", bar: "text-red-500", label: "Critical" };
  };

  const config = getColor(score);

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">Agent Health</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center">
        <div className="relative h-24 w-24 mb-2">
          <svg className="h-24 w-24 -rotate-90" viewBox="0 0 100 100">
            <circle cx="50" cy="50" r="40" fill="none" stroke="currentColor" strokeWidth="10" className="text-muted/20" />
            <circle
              cx="50" cy="50" r="40" fill="none" strokeWidth="10"
              strokeDasharray={`${(score / 100) * 251} 251`}
              strokeLinecap="round"
              className={config.bar}
            />
          </svg>
          <div className="absolute inset-0 flex items-center justify-center">
            <span className={cn("text-2xl font-bold", config.text)}>{Math.round(score)}</span>
          </div>
        </div>
        <span className={cn("text-sm font-medium", config.text)}>{config.label}</span>
        {activeAlerts > 0 && (
          <span className="text-xs text-red-500 mt-1">{activeAlerts} active alert{activeAlerts > 1 ? "s" : ""}</span>
        )}
      </CardContent>
    </Card>
  );
}
