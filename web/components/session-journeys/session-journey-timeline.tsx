"use client";

import * as React from "react";
import {
  Search,
  Clock,
  DollarSign,
  Layers,
  Brain,
  Route,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface WorkflowPhase {
  id: string;
  name: string;
  type: "planning" | "implementation" | "testing" | "debugging" | "review";
  durationSeconds: number;
  cost: number;
  tokenCount: number;
  confidenceScore: number;
}

interface JourneySummary {
  totalDuration: number;
  totalCost: number;
  phaseCount: number;
}

interface SessionJourneyTimelineProps {
  sessionId?: string;
}

const phaseConfig = {
  planning: { color: "bg-blue-500", label: "Planning", textColor: "text-blue-700", bgLight: "bg-blue-500/10" },
  implementation: { color: "bg-green-500", label: "Implementation", textColor: "text-green-700", bgLight: "bg-green-500/10" },
  testing: { color: "bg-purple-500", label: "Testing", textColor: "text-purple-700", bgLight: "bg-purple-500/10" },
  debugging: { color: "bg-red-500", label: "Debugging", textColor: "text-red-700", bgLight: "bg-red-500/10" },
  review: { color: "bg-yellow-500", label: "Review", textColor: "text-yellow-700", bgLight: "bg-yellow-500/10" },
};

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return secs > 0 ? `${minutes}m ${secs}s` : `${minutes}m`;
}

export function SessionJourneyTimeline({
  sessionId: initialSessionId,
}: SessionJourneyTimelineProps) {
  const [sessionId, setSessionId] = React.useState(initialSessionId ?? "");
  const [submittedId, setSubmittedId] = React.useState(initialSessionId ?? "");

  // TODO: Replace with useSessionJourneys(submittedId) hook when available
  const isLoading = false;
  const phases: WorkflowPhase[] = [];
  const summary: JourneySummary | null = null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setSubmittedId(sessionId);
  };

  const totalDuration = phases.reduce((acc, p) => acc + p.durationSeconds, 0);

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit} className="flex items-center gap-2">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Enter session ID..."
            value={sessionId}
            onChange={(e) => setSessionId(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button type="submit" disabled={!sessionId}>
          Load Journey
        </Button>
      </form>

      {summary && (
        <div className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Duration
              </CardTitle>
              <Clock className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {formatDuration(summary.totalDuration)}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Cost
              </CardTitle>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                ${summary.totalCost.toFixed(4)}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Phases
              </CardTitle>
              <Layers className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{summary.phaseCount}</div>
            </CardContent>
          </Card>
        </div>
      )}

      {isLoading && (
        <div className="space-y-4">
          <div className="h-12 bg-muted animate-pulse rounded-lg" />
          <div className="space-y-3">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="h-20 bg-muted animate-pulse rounded-lg" />
            ))}
          </div>
        </div>
      )}

      {phases.length > 0 && (
        <>
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Timeline</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex h-10 rounded-lg overflow-hidden">
                {phases.map((phase) => {
                  const config = phaseConfig[phase.type];
                  const widthPercent =
                    totalDuration > 0
                      ? (phase.durationSeconds / totalDuration) * 100
                      : 0;
                  return (
                    <div
                      key={phase.id}
                      className={cn(
                        "flex items-center justify-center text-xs font-medium text-white transition-all",
                        config.color
                      )}
                      style={{ width: `${Math.max(widthPercent, 3)}%` }}
                      title={`${config.label}: ${formatDuration(phase.durationSeconds)}`}
                    >
                      {widthPercent > 10 && config.label}
                    </div>
                  );
                })}
              </div>
              <div className="flex flex-wrap gap-3 mt-3">
                {Object.entries(phaseConfig).map(([key, config]) => (
                  <div key={key} className="flex items-center gap-1.5 text-xs">
                    <div className={cn("h-2.5 w-2.5 rounded-sm", config.color)} />
                    <span className="text-muted-foreground">{config.label}</span>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          <div className="space-y-3">
            {phases.map((phase) => {
              const config = phaseConfig[phase.type];
              return (
                <Card key={phase.id}>
                  <CardContent className="flex items-center justify-between p-4">
                    <div className="flex items-center gap-3 min-w-0">
                      <div className={cn("h-10 w-1.5 rounded-full shrink-0", config.color)} />
                      <div className="min-w-0">
                        <p className="text-sm font-medium">{phase.name}</p>
                        <div className="flex items-center gap-2 mt-1">
                          <Badge
                            variant="outline"
                            className={cn("text-xs", config.bgLight, config.textColor)}
                          >
                            {config.label}
                          </Badge>
                          <span className="text-xs text-muted-foreground flex items-center gap-1">
                            <Brain className="h-3 w-3" />
                            {(phase.confidenceScore * 100).toFixed(0)}% confidence
                          </span>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-4 shrink-0 text-sm text-muted-foreground">
                      <div className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        <span>{formatDuration(phase.durationSeconds)}</span>
                      </div>
                      <div className="flex items-center gap-1">
                        <DollarSign className="h-3 w-3" />
                        <span>${phase.cost.toFixed(4)}</span>
                      </div>
                      <span className="text-xs">
                        {phase.tokenCount.toLocaleString()} tokens
                      </span>
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </>
      )}

      {submittedId && !isLoading && phases.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          <Route className="h-12 w-12 mx-auto mb-4 opacity-50" />
          <p className="text-lg font-medium">No journey data found</p>
          <p className="text-sm mt-2">
            Enter a valid session ID to view the workflow journey timeline
          </p>
        </div>
      )}
    </div>
  );
}
