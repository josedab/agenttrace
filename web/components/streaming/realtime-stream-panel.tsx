"use client";

import * as React from "react";
import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import { useWebSocketStream } from "@/hooks/use-websocket-stream";
import {
  Activity,
  Clock,
  DollarSign,
  Zap,
  FileText,
  Terminal,
  AlertTriangle,
  AlertCircle,
  Bot,
  Cpu,
  Wrench,
  GitBranch,
  Plug,
  PlugZap,
  RefreshCw,
} from "lucide-react";

const activityIcons: Record<string, React.ElementType> = {
  "observation.start": Cpu,
  "observation.end": Cpu,
  "file.change": FileText,
  "terminal.output": Terminal,
  "cost.update": Bot,
  "error.occurred": AlertCircle,
  "trace.activity": Wrench,
  "metrics.snapshot": GitBranch,
};

const statusColors: Record<string, string> = {
  running: "bg-blue-500",
  completed: "bg-green-500",
  error: "bg-red-500",
  pending: "bg-yellow-500",
};

function formatDuration(ms: number) {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${(ms / 60000).toFixed(1)}m`;
}

function formatCost(cost: number) {
  return `$${cost.toFixed(4)}`;
}

export function RealtimeStreamPanel() {
  const [traceInput, setTraceInput] = useState("");
  const [activeTraceId, setActiveTraceId] = useState<string | null>(null);
  const [autoReconnect, setAutoReconnect] = useState(true);

  const { metrics, activities, connected, error, disconnect, reconnect } =
    useWebSocketStream(activeTraceId);

  const scrollRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [activities]);

  const handleConnect = () => {
    const id = traceInput.trim();
    if (id) setActiveTraceId(id);
  };

  const handleDisconnect = () => {
    disconnect();
    setActiveTraceId(null);
  };

  return (
    <div className="space-y-6">
      {/* Connection bar */}
      <Card>
        <CardContent className="flex items-center gap-3 py-4">
          <span
            className={cn(
              "h-3 w-3 rounded-full shrink-0",
              connected ? "bg-green-500 animate-pulse" : "bg-red-500",
            )}
          />
          <Badge variant={connected ? "default" : "destructive"}>
            {connected ? "Connected" : "Disconnected"}
          </Badge>

          {!activeTraceId ? (
            <>
              <Input
                placeholder="Enter trace ID…"
                value={traceInput}
                onChange={(e) => setTraceInput(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleConnect()}
                className="max-w-sm"
              />
              <Button onClick={handleConnect} disabled={!traceInput.trim()}>
                <Plug className="h-4 w-4 mr-1.5" />
                Connect
              </Button>
            </>
          ) : (
            <>
              <span className="text-sm text-muted-foreground">
                Trace: <code className="font-mono">{activeTraceId}</code>
              </span>
              <div className="ml-auto flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={reconnect}>
                  <RefreshCw className="h-3.5 w-3.5 mr-1" />
                  Reconnect
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={handleDisconnect}
                >
                  <PlugZap className="h-3.5 w-3.5 mr-1" />
                  Disconnect
                </Button>
                <label className="flex items-center gap-1.5 text-xs text-muted-foreground cursor-pointer ml-2">
                  <input
                    type="checkbox"
                    checked={autoReconnect}
                    onChange={(e) => setAutoReconnect(e.target.checked)}
                    className="rounded"
                  />
                  Auto-reconnect
                </label>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {error && (
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2">
          <AlertCircle className="h-4 w-4" />
          {error}
        </div>
      )}

      {activeTraceId && (
        <>
          {/* Live metrics grid */}
          <div className="space-y-4">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              <Card>
                <CardHeader className="pb-2 pt-4 px-4">
                  <CardTitle className="text-xs font-medium text-muted-foreground flex items-center gap-1">
                    <Clock className="h-3 w-3" /> Elapsed
                  </CardTitle>
                </CardHeader>
                <CardContent className="px-4 pb-4">
                  <div className="text-2xl font-bold">
                    {metrics ? formatDuration(metrics.elapsedMs) : "--"}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-2 pt-4 px-4">
                  <CardTitle className="text-xs font-medium text-muted-foreground flex items-center gap-1">
                    <DollarSign className="h-3 w-3" /> Total Cost
                  </CardTitle>
                </CardHeader>
                <CardContent className="px-4 pb-4">
                  <div className="text-2xl font-bold">
                    {metrics ? formatCost(metrics.totalCost) : "--"}
                  </div>
                  {metrics && metrics.costPerMinute > 0 && (
                    <p className="text-xs text-muted-foreground">
                      {formatCost(metrics.costPerMinute)}/min
                    </p>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-2 pt-4 px-4">
                  <CardTitle className="text-xs font-medium text-muted-foreground flex items-center gap-1">
                    <Zap className="h-3 w-3" /> Tokens
                  </CardTitle>
                </CardHeader>
                <CardContent className="px-4 pb-4">
                  <div className="text-2xl font-bold">
                    {metrics ? metrics.totalTokens.toLocaleString() : "--"}
                  </div>
                  {metrics && metrics.tokensPerSecond > 0 && (
                    <p className="text-xs text-muted-foreground">
                      {metrics.tokensPerSecond.toFixed(0)} tok/s
                    </p>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-2 pt-4 px-4">
                  <CardTitle className="text-xs font-medium text-muted-foreground flex items-center gap-1">
                    <Activity className="h-3 w-3" /> Spans
                  </CardTitle>
                </CardHeader>
                <CardContent className="px-4 pb-4">
                  <div className="text-2xl font-bold">
                    {metrics
                      ? `${metrics.activeSpans}/${metrics.completedSpans}`
                      : "--"}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    active/completed
                  </p>
                </CardContent>
              </Card>
            </div>

            <div className="grid grid-cols-3 gap-3">
              <Card>
                <CardContent className="flex items-center gap-2 py-3 px-4">
                  <FileText className="h-4 w-4 text-blue-500" />
                  <div>
                    <div className="text-lg font-semibold">
                      {metrics?.filesModified ?? 0}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      Files modified
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardContent className="flex items-center gap-2 py-3 px-4">
                  <Terminal className="h-4 w-4 text-green-500" />
                  <div>
                    <div className="text-lg font-semibold">
                      {metrics?.terminalCommands ?? 0}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      Commands
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardContent className="flex items-center gap-2 py-3 px-4">
                  <AlertTriangle className="h-4 w-4 text-red-500" />
                  <div>
                    <div className="text-lg font-semibold">
                      {metrics?.errorCount ?? 0}
                    </div>
                    <div className="text-xs text-muted-foreground">Errors</div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>

          {/* Activity feed */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <Activity className="h-4 w-4" />
                Activity Feed
                <Badge variant="outline" className="ml-auto text-[10px]">
                  {activities.length} events
                </Badge>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-[400px]" ref={scrollRef}>
                <div className="space-y-1">
                  {activities.length === 0 && (
                    <div className="flex items-center justify-center h-32 text-muted-foreground text-sm">
                      Waiting for agent activity…
                    </div>
                  )}
                  {activities.map((activity) => {
                    const Icon = activityIcons[activity.type] || Wrench;
                    return (
                      <div
                        key={activity.id}
                        className={cn(
                          "flex items-start gap-2 p-2 rounded-md text-sm transition-colors",
                          activity.status === "error"
                            ? "bg-destructive/10"
                            : "hover:bg-muted/50",
                        )}
                      >
                        <span
                          className={cn(
                            "mt-0.5 h-2 w-2 rounded-full shrink-0",
                            statusColors[activity.status] || "bg-gray-400",
                          )}
                        />
                        <Icon className="h-4 w-4 shrink-0 mt-0.5 text-muted-foreground" />
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-medium truncate">
                              {activity.title}
                            </span>
                            {activity.durationMs != null && (
                              <Badge
                                variant="outline"
                                className="text-[10px] px-1 py-0"
                              >
                                {formatDuration(activity.durationMs)}
                              </Badge>
                            )}
                          </div>
                          {activity.description && (
                            <p className="text-xs text-muted-foreground mt-0.5 truncate">
                              {activity.description}
                            </p>
                          )}
                        </div>
                        <span className="text-[10px] text-muted-foreground shrink-0">
                          {new Date(activity.timestamp).toLocaleTimeString()}
                        </span>
                      </div>
                    );
                  })}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
