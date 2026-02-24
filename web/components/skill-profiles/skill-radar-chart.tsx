"use client";

import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface SkillRadarChartProps {
  skills: Record<string, { score: number; confidence: number }>;
  agentName: string;
}

export function SkillRadarChart({ skills, agentName }: SkillRadarChartProps) {
  const dimensions = Object.entries(skills);
  const maxScore = 100;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm">{agentName}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {dimensions.map(([name, { score, confidence }]) => (
            <div key={name}>
              <div className="flex justify-between text-xs mb-1">
                <span className="capitalize">{name.replace(/_/g, " ")}</span>
                <span className="font-medium">{Math.round(score)}</span>
              </div>
              <div className="h-2 bg-muted rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full bg-primary transition-all"
                  style={{
                    width: `${(score / maxScore) * 100}%`,
                    opacity: 0.5 + confidence * 0.5,
                  }}
                />
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
