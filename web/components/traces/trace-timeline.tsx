"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Cpu, MessageSquare, Zap, Timer, Wrench } from "lucide-react";

interface Observation {
  id: string;
  type: "SPAN" | "GENERATION" | "EVENT";
  name: string | null;
  startTime: string;
  endTime: string | null;
  latency: number | null;
  level: string;
  parentObservationId: string | null;
  model: string | null;
  totalCost: number | null;
  usage: {
    promptTokens: number | null;
    completionTokens: number | null;
    totalTokens: number | null;
  } | null;
}

interface Trace {
  id: string;
  startTime: string;
  endTime: string | null;
  latency: number | null;
}

interface TraceTimelineProps {
  trace: Trace;
  observations: Observation[];
  selectedId: string | null;
  onSelect: (id: string | null) => void;
}

export function TraceTimeline({
  trace,
  observations,
  selectedId,
  onSelect,
}: TraceTimelineProps) {
  const traceStart = new Date(trace.startTime).getTime();
  const traceDuration = trace.latency || 1;

  // Sort observations by start time
  const sortedObservations = [...observations].sort(
    (a, b) => new Date(a.startTime).getTime() - new Date(b.startTime).getTime()
  );

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="space-y-2">
          {sortedObservations.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-8">
              No observations found for this trace
            </p>
          ) : (
            sortedObservations.map((observation) => {
              const startOffset = new Date(observation.startTime).getTime() - traceStart;
              const duration = observation.latency || 0;
              const leftPercent = (startOffset / traceDuration) * 100;
              const widthPercent = Math.max((duration / traceDuration) * 100, 1);

              return (
                <div
                  key={observation.id}
                  className={cn(
                    "flex items-center gap-4 p-3 rounded-lg cursor-pointer transition-colors",
                    selectedId === observation.id
                      ? "bg-primary/10 border border-primary"
                      : "hover:bg-muted"
                  )}
                  onClick={() => onSelect(observation.id)}
                >
                  {/* Icon */}
                  <div className="flex-shrink-0">
                    <ObservationIcon type={observation.type} />
                  </div>

                  {/* Name and details */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium truncate">
                        {observation.name || observation.type}
                      </span>
                      <TypeBadge type={observation.type} />
                      {observation.model && (
                        <Badge variant="outline" className="text-xs">
                          {observation.model}
                        </Badge>
                      )}
                    </div>
                    <div className="flex items-center gap-4 text-xs text-muted-foreground mt-1">
                      {observation.latency && (
                        <span className="flex items-center gap-1">
                          <Timer className="h-3 w-3" />
                          {formatLatency(observation.latency)}
                        </span>
                      )}
                      {observation.totalCost !== null && (
                        <span>${observation.totalCost.toFixed(4)}</span>
                      )}
                      {observation.usage?.totalTokens && (
                        <span>{observation.usage.totalTokens.toLocaleString()} tokens</span>
                      )}
                    </div>
                  </div>

                  {/* Timeline bar */}
                  <div className="w-48 flex-shrink-0">
                    <div className="relative h-6 bg-muted rounded">
                      <div
                        className={cn(
                          "absolute h-full rounded",
                          observation.type === "GENERATION"
                            ? "bg-blue-500"
                            : observation.type === "EVENT"
                            ? "bg-yellow-500"
                            : "bg-green-500",
                          observation.level === "ERROR" && "bg-red-500"
                        )}
                        style={{
                          left: `${Math.min(leftPercent, 100)}%`,
                          width: `${Math.min(widthPercent, 100 - leftPercent)}%`,
                        }}
                      />
                    </div>
                  </div>
                </div>
              );
            })
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function ObservationIcon({ type }: { type: Observation["type"] }) {
  switch (type) {
    case "GENERATION":
      return <Cpu className="h-5 w-5 text-blue-500" />;
    case "EVENT":
      return <Zap className="h-5 w-5 text-yellow-500" />;
    default:
      return <Wrench className="h-5 w-5 text-green-500" />;
  }
}

function TypeBadge({ type }: { type: Observation["type"] }) {
  return (
    <Badge
      variant="outline"
      className={cn(
        "text-xs",
        type === "GENERATION" && "border-blue-500 text-blue-500",
        type === "EVENT" && "border-yellow-500 text-yellow-500",
        type === "SPAN" && "border-green-500 text-green-500"
      )}
    >
      {type}
    </Badge>
  );
}

function formatLatency(ms: number): string {
  if (ms >= 1000) {
    return (ms / 1000).toFixed(2) + "s";
  }
  return ms.toFixed(0) + "ms";
}
