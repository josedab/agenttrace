"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import { formatDistanceToNow, format } from "date-fns";
import { Clock, DollarSign, AlertCircle, Copy, ExternalLink, Tag } from "lucide-react";

import { cn } from "@/lib/utils";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { TraceTimeline } from "@/components/traces/trace-timeline";
import { TraceTree } from "@/components/traces/trace-tree";
import { ObservationPanel } from "@/components/traces/observation-panel";
import { TraceDetailSkeleton } from "@/components/traces/trace-detail-skeleton";
import { toast } from "sonner";

interface TraceDetailProps {
  traceId: string;
}

export function TraceDetail({ traceId }: TraceDetailProps) {
  const [selectedObservationId, setSelectedObservationId] = React.useState<string | null>(null);

  const { data: trace, isLoading, error } = useQuery({
    queryKey: ["trace", traceId],
    queryFn: () => api.traces.get(traceId),
  });

  const { data: observations } = useQuery({
    queryKey: ["trace-observations", traceId],
    queryFn: () => api.observations.listByTrace(traceId),
    enabled: !!trace,
  });

  const selectedObservation = React.useMemo(() => {
    if (!selectedObservationId || !observations) return null;
    return observations.find((o) => o.id === selectedObservationId) || null;
  }, [selectedObservationId, observations]);

  if (isLoading) {
    return <TraceDetailSkeleton />;
  }

  if (error || !trace) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <AlertCircle className="h-12 w-12 text-destructive mb-4" />
        <p className="text-destructive">Failed to load trace</p>
      </div>
    );
  }

  const copyTraceId = () => {
    navigator.clipboard.writeText(traceId);
    toast.success("Trace ID copied to clipboard");
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{trace.name || "Unnamed Trace"}</h1>
            <LevelBadge level={trace.level} />
          </div>
          <div className="flex items-center gap-2 mt-2 text-sm text-muted-foreground">
            <span className="font-mono">{traceId.slice(0, 16)}...</span>
            <Button variant="ghost" size="sm" onClick={copyTraceId}>
              <Copy className="h-3 w-3" />
            </Button>
          </div>
        </div>
        <div className="flex items-center gap-4">
          {trace.sessionId && (
            <Button variant="outline" size="sm" asChild>
              <a href={`/sessions/${trace.sessionId}`}>
                <ExternalLink className="h-4 w-4 mr-2" />
                View Session
              </a>
            </Button>
          )}
        </div>
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Duration</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-muted-foreground" />
              <span className="text-xl font-semibold">
                {trace.latency ? formatLatency(trace.latency) : "-"}
              </span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Cost</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <DollarSign className="h-4 w-4 text-muted-foreground" />
              <span className="text-xl font-semibold">
                {trace.totalCost !== null ? `$${trace.totalCost.toFixed(4)}` : "-"}
              </span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Start Time</CardDescription>
          </CardHeader>
          <CardContent>
            <span className="text-sm font-medium">
              {format(new Date(trace.startTime), "PPpp")}
            </span>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Tokens</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-semibold">
              {trace.usage?.totalTokens?.toLocaleString() ?? "-"}
            </div>
            {trace.usage && (
              <div className="text-xs text-muted-foreground">
                {trace.usage.promptTokens?.toLocaleString()} in / {trace.usage.completionTokens?.toLocaleString()} out
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Tags */}
      {trace.tags && trace.tags.length > 0 && (
        <div className="flex items-center gap-2">
          <Tag className="h-4 w-4 text-muted-foreground" />
          <div className="flex flex-wrap gap-1">
            {trace.tags.map((tag) => (
              <Badge key={tag} variant="secondary">
                {tag}
              </Badge>
            ))}
          </div>
        </div>
      )}

      {/* Main content */}
      <div className="grid lg:grid-cols-3 gap-6">
        {/* Left panel - Tree/Timeline */}
        <div className="lg:col-span-2">
          <Tabs defaultValue="timeline">
            <TabsList>
              <TabsTrigger value="timeline">Timeline</TabsTrigger>
              <TabsTrigger value="tree">Tree View</TabsTrigger>
              <TabsTrigger value="input">Input</TabsTrigger>
              <TabsTrigger value="output">Output</TabsTrigger>
            </TabsList>
            <TabsContent value="timeline" className="mt-4">
              <TraceTimeline
                trace={trace}
                observations={observations || []}
                selectedId={selectedObservationId}
                onSelect={setSelectedObservationId}
              />
            </TabsContent>
            <TabsContent value="tree" className="mt-4">
              <TraceTree
                observations={observations || []}
                selectedId={selectedObservationId}
                onSelect={setSelectedObservationId}
              />
            </TabsContent>
            <TabsContent value="input" className="mt-4">
              <Card>
                <CardContent className="pt-6">
                  <pre className="whitespace-pre-wrap text-sm font-mono bg-muted p-4 rounded-lg overflow-auto max-h-[500px]">
                    {trace.input ? formatJson(trace.input) : "No input"}
                  </pre>
                </CardContent>
              </Card>
            </TabsContent>
            <TabsContent value="output" className="mt-4">
              <Card>
                <CardContent className="pt-6">
                  <pre className="whitespace-pre-wrap text-sm font-mono bg-muted p-4 rounded-lg overflow-auto max-h-[500px]">
                    {trace.output ? formatJson(trace.output) : "No output"}
                  </pre>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>

        {/* Right panel - Observation detail */}
        <div>
          <ObservationPanel observation={selectedObservation} />
        </div>
      </div>
    </div>
  );
}

function LevelBadge({ level }: { level: string }) {
  return (
    <Badge
      variant="outline"
      className={cn(
        level === "ERROR" && "border-destructive text-destructive",
        level === "WARNING" && "border-yellow-500 text-yellow-500",
        level === "DEBUG" && "border-blue-500 text-blue-500"
      )}
    >
      {level}
    </Badge>
  );
}

function formatLatency(ms: number): string {
  if (ms >= 1000) {
    return (ms / 1000).toFixed(2) + "s";
  }
  return ms.toFixed(0) + "ms";
}

function formatJson(value: string): string {
  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch {
    return value;
  }
}
