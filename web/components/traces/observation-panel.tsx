"use client";

import * as React from "react";
import { format } from "date-fns";
import { Clock, DollarSign, Cpu, Hash, Tag, Copy } from "lucide-react";
import { toast } from "sonner";

import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";

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
  modelParameters: Record<string, any> | null;
  input: string | null;
  output: string | null;
  metadata: Record<string, any> | null;
  totalCost: number | null;
  inputCost: number | null;
  outputCost: number | null;
  usage: {
    promptTokens: number | null;
    completionTokens: number | null;
    totalTokens: number | null;
  } | null;
}

interface ObservationPanelProps {
  observation: Observation | null;
}

export function ObservationPanel({ observation }: ObservationPanelProps) {
  if (!observation) {
    return (
      <Card className="h-full">
        <CardContent className="flex items-center justify-center h-96">
          <p className="text-sm text-muted-foreground">
            Select an observation to view details
          </p>
        </CardContent>
      </Card>
    );
  }

  const copyId = () => {
    navigator.clipboard.writeText(observation.id);
    toast.success("Observation ID copied to clipboard");
  };

  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <CardTitle className="text-lg">
              {observation.name || observation.type}
            </CardTitle>
            <CardDescription className="flex items-center gap-2 mt-1">
              <span className="font-mono text-xs">{observation.id.slice(0, 16)}...</span>
              <Button variant="ghost" size="sm" className="h-5 w-5 p-0" onClick={copyId}>
                <Copy className="h-3 w-3" />
              </Button>
            </CardDescription>
          </div>
          <TypeBadge type={observation.type} level={observation.level} />
        </div>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="details">
          <TabsList className="w-full">
            <TabsTrigger value="details" className="flex-1">Details</TabsTrigger>
            <TabsTrigger value="input" className="flex-1">Input</TabsTrigger>
            <TabsTrigger value="output" className="flex-1">Output</TabsTrigger>
          </TabsList>

          <TabsContent value="details" className="space-y-4 mt-4">
            {/* Metrics */}
            <div className="grid grid-cols-2 gap-4">
              <MetricItem
                icon={<Clock className="h-4 w-4" />}
                label="Duration"
                value={observation.latency ? formatLatency(observation.latency) : "-"}
              />
              <MetricItem
                icon={<DollarSign className="h-4 w-4" />}
                label="Cost"
                value={observation.totalCost !== null ? `$${observation.totalCost.toFixed(4)}` : "-"}
              />
            </div>

            {/* Model info for generations */}
            {observation.type === "GENERATION" && (
              <>
                <Separator />
                <div className="space-y-2">
                  <h4 className="text-sm font-medium">Model</h4>
                  <div className="flex items-center gap-2">
                    <Cpu className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm">{observation.model || "Unknown"}</span>
                  </div>
                  {observation.modelParameters && (
                    <div className="text-xs text-muted-foreground">
                      {Object.entries(observation.modelParameters).map(([key, value]) => (
                        <div key={key} className="flex justify-between">
                          <span>{key}:</span>
                          <span>{String(value)}</span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </>
            )}

            {/* Token usage */}
            {observation.usage && (
              <>
                <Separator />
                <div className="space-y-2">
                  <h4 className="text-sm font-medium">Token Usage</h4>
                  <div className="grid grid-cols-3 gap-2 text-center">
                    <div>
                      <div className="text-lg font-semibold">
                        {observation.usage.promptTokens?.toLocaleString() ?? "-"}
                      </div>
                      <div className="text-xs text-muted-foreground">Input</div>
                    </div>
                    <div>
                      <div className="text-lg font-semibold">
                        {observation.usage.completionTokens?.toLocaleString() ?? "-"}
                      </div>
                      <div className="text-xs text-muted-foreground">Output</div>
                    </div>
                    <div>
                      <div className="text-lg font-semibold">
                        {observation.usage.totalTokens?.toLocaleString() ?? "-"}
                      </div>
                      <div className="text-xs text-muted-foreground">Total</div>
                    </div>
                  </div>
                </div>
              </>
            )}

            {/* Cost breakdown */}
            {(observation.inputCost !== null || observation.outputCost !== null) && (
              <>
                <Separator />
                <div className="space-y-2">
                  <h4 className="text-sm font-medium">Cost Breakdown</h4>
                  <div className="grid grid-cols-2 gap-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Input:</span>
                      <span>${observation.inputCost?.toFixed(6) ?? "-"}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Output:</span>
                      <span>${observation.outputCost?.toFixed(6) ?? "-"}</span>
                    </div>
                  </div>
                </div>
              </>
            )}

            {/* Timestamps */}
            <Separator />
            <div className="space-y-2">
              <h4 className="text-sm font-medium">Timestamps</h4>
              <div className="text-xs space-y-1">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Start:</span>
                  <span>{format(new Date(observation.startTime), "PPpp")}</span>
                </div>
                {observation.endTime && (
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">End:</span>
                    <span>{format(new Date(observation.endTime), "PPpp")}</span>
                  </div>
                )}
              </div>
            </div>

            {/* Metadata */}
            {observation.metadata && Object.keys(observation.metadata).length > 0 && (
              <>
                <Separator />
                <div className="space-y-2">
                  <h4 className="text-sm font-medium">Metadata</h4>
                  <pre className="text-xs bg-muted p-2 rounded overflow-auto max-h-32">
                    {JSON.stringify(observation.metadata, null, 2)}
                  </pre>
                </div>
              </>
            )}
          </TabsContent>

          <TabsContent value="input" className="mt-4">
            <pre className="whitespace-pre-wrap text-sm font-mono bg-muted p-4 rounded-lg overflow-auto max-h-[400px]">
              {observation.input ? formatJson(observation.input) : "No input"}
            </pre>
          </TabsContent>

          <TabsContent value="output" className="mt-4">
            <pre className="whitespace-pre-wrap text-sm font-mono bg-muted p-4 rounded-lg overflow-auto max-h-[400px]">
              {observation.output ? formatJson(observation.output) : "No output"}
            </pre>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}

function MetricItem({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="flex items-center gap-2 p-2 bg-muted rounded">
      <div className="text-muted-foreground">{icon}</div>
      <div>
        <div className="text-xs text-muted-foreground">{label}</div>
        <div className="font-medium">{value}</div>
      </div>
    </div>
  );
}

function TypeBadge({ type, level }: { type: string; level: string }) {
  return (
    <Badge
      variant="outline"
      className={cn(
        level === "ERROR" && "border-destructive text-destructive",
        level !== "ERROR" && type === "GENERATION" && "border-blue-500 text-blue-500",
        level !== "ERROR" && type === "EVENT" && "border-yellow-500 text-yellow-500",
        level !== "ERROR" && type === "SPAN" && "border-green-500 text-green-500"
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

function formatJson(value: string): string {
  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch {
    return value;
  }
}
