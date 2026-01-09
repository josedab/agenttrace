"use client";

import * as React from "react";
import { useState, useEffect, useCallback, useRef } from "react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Slider } from "@/components/ui/slider";
import { Separator } from "@/components/ui/separator";
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  RotateCcw,
  Download,
  Terminal,
  FileCode,
  GitBranch,
  MessageSquare,
  Cpu,
  Zap,
  AlertCircle,
  Bookmark,
  ChevronRight,
  ChevronDown,
} from "lucide-react";

// Types matching the backend domain models
interface ReplayEvent {
  id: string;
  type: "llm_call" | "tool_call" | "file_operation" | "terminal_command" | "checkpoint" | "git_operation" | "user_input" | "agent_thought" | "error";
  timestamp: string;
  durationMs?: number;
  title: string;
  description?: string;
  status: "success" | "error" | "pending" | "running";
  data: ReplayEventData;
  children?: ReplayEvent[];
}

interface ReplayEventData {
  model?: string;
  input?: any;
  output?: any;
  tokensInput?: number;
  tokensOutput?: number;
  cost?: number;
  toolName?: string;
  arguments?: any;
  result?: any;
  filePath?: string;
  operation?: string;
  diff?: string;
  command?: string;
  exitCode?: number;
  stdout?: string;
  stderr?: string;
  checkpointId?: string;
  gitCommit?: string;
  gitBranch?: string;
  error?: string;
}

interface ReplaySummary {
  totalEvents: number;
  llmCalls: number;
  toolCalls: number;
  fileOperations: number;
  terminalCommands: number;
  checkpoints: number;
  errors: number;
  totalTokens: number;
  totalCost: number;
  averageLatencyMs: number;
}

interface ReplayTimeline {
  traceId: string;
  traceName: string;
  startTime: string;
  endTime?: string;
  durationMs: number;
  events: ReplayEvent[];
  summary: ReplaySummary;
}

interface TraceReplayPlayerProps {
  timeline: ReplayTimeline;
  onExport?: () => void;
}

export function TraceReplayPlayer({ timeline, onExport }: TraceReplayPlayerProps) {
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentEventIndex, setCurrentEventIndex] = useState(-1);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);
  const [elapsedTime, setElapsedTime] = useState(0);
  const [expandedEvents, setExpandedEvents] = useState<Set<string>>(new Set());
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  const totalDuration = timeline.durationMs;
  const events = timeline.events;

  // Calculate event times relative to start
  const eventTimes = events.map((event) => {
    const eventTime = new Date(event.timestamp).getTime();
    const startTime = new Date(timeline.startTime).getTime();
    return eventTime - startTime;
  });

  // Update current event based on elapsed time
  useEffect(() => {
    const currentEvent = eventTimes.findIndex(
      (time, index) =>
        time <= elapsedTime &&
        (index === eventTimes.length - 1 || eventTimes[index + 1] > elapsedTime)
    );
    setCurrentEventIndex(currentEvent);
  }, [elapsedTime, eventTimes]);

  // Playback logic
  useEffect(() => {
    if (isPlaying) {
      intervalRef.current = setInterval(() => {
        setElapsedTime((prev) => {
          const next = prev + 100 * playbackSpeed;
          if (next >= totalDuration) {
            setIsPlaying(false);
            return totalDuration;
          }
          return next;
        });
      }, 100);
    } else {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [isPlaying, playbackSpeed, totalDuration]);

  const handlePlayPause = useCallback(() => {
    if (elapsedTime >= totalDuration) {
      setElapsedTime(0);
    }
    setIsPlaying((prev) => !prev);
  }, [elapsedTime, totalDuration]);

  const handleReset = useCallback(() => {
    setIsPlaying(false);
    setElapsedTime(0);
    setCurrentEventIndex(-1);
  }, []);

  const handleSkipBack = useCallback(() => {
    const prevEventTime = eventTimes
      .slice(0, currentEventIndex)
      .findLast((time) => time < elapsedTime - 100);
    if (prevEventTime !== undefined) {
      setElapsedTime(prevEventTime);
    } else {
      setElapsedTime(0);
    }
  }, [eventTimes, currentEventIndex, elapsedTime]);

  const handleSkipForward = useCallback(() => {
    const nextEventTime = eventTimes.find((time) => time > elapsedTime);
    if (nextEventTime !== undefined) {
      setElapsedTime(nextEventTime);
    }
  }, [eventTimes, elapsedTime]);

  const handleSeek = useCallback((value: number[]) => {
    setElapsedTime(value[0]);
  }, []);

  const toggleEventExpanded = useCallback((eventId: string) => {
    setExpandedEvents((prev) => {
      const next = new Set(prev);
      if (next.has(eventId)) {
        next.delete(eventId);
      } else {
        next.add(eventId);
      }
      return next;
    });
  }, []);

  const formatTime = (ms: number) => {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    const remainingMs = Math.floor((ms % 1000) / 10);
    return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}.${remainingMs.toString().padStart(2, "0")}`;
  };

  return (
    <div className="space-y-4">
      {/* Playback Controls */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center gap-4">
            {/* Control Buttons */}
            <div className="flex items-center gap-2">
              <Button variant="outline" size="icon" onClick={handleReset}>
                <RotateCcw className="h-4 w-4" />
              </Button>
              <Button variant="outline" size="icon" onClick={handleSkipBack}>
                <SkipBack className="h-4 w-4" />
              </Button>
              <Button size="icon" onClick={handlePlayPause}>
                {isPlaying ? (
                  <Pause className="h-4 w-4" />
                ) : (
                  <Play className="h-4 w-4" />
                )}
              </Button>
              <Button variant="outline" size="icon" onClick={handleSkipForward}>
                <SkipForward className="h-4 w-4" />
              </Button>
            </div>

            {/* Timeline Slider */}
            <div className="flex-1">
              <Slider
                value={[elapsedTime]}
                max={totalDuration}
                step={100}
                onValueChange={handleSeek}
                className="w-full"
              />
            </div>

            {/* Time Display */}
            <div className="flex items-center gap-2 text-sm font-mono">
              <span>{formatTime(elapsedTime)}</span>
              <span className="text-muted-foreground">/</span>
              <span className="text-muted-foreground">
                {formatTime(totalDuration)}
              </span>
            </div>

            {/* Speed Control */}
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">Speed:</span>
              <select
                value={playbackSpeed}
                onChange={(e) => setPlaybackSpeed(Number(e.target.value))}
                className="text-sm border rounded px-2 py-1"
              >
                <option value={0.5}>0.5x</option>
                <option value={1}>1x</option>
                <option value={2}>2x</option>
                <option value={4}>4x</option>
                <option value={10}>10x</option>
              </select>
            </div>

            {/* Export Button */}
            {onExport && (
              <Button variant="outline" size="sm" onClick={onExport}>
                <Download className="h-4 w-4 mr-2" />
                Export
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Summary Stats */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid grid-cols-7 gap-4 text-center">
            <div>
              <div className="text-2xl font-bold">{timeline.summary.totalEvents}</div>
              <div className="text-xs text-muted-foreground">Events</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-blue-500">{timeline.summary.llmCalls}</div>
              <div className="text-xs text-muted-foreground">LLM Calls</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-green-500">{timeline.summary.toolCalls}</div>
              <div className="text-xs text-muted-foreground">Tool Calls</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-purple-500">{timeline.summary.fileOperations}</div>
              <div className="text-xs text-muted-foreground">File Ops</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-orange-500">{timeline.summary.terminalCommands}</div>
              <div className="text-xs text-muted-foreground">Commands</div>
            </div>
            <div>
              <div className="text-2xl font-bold">{timeline.summary.totalTokens.toLocaleString()}</div>
              <div className="text-xs text-muted-foreground">Tokens</div>
            </div>
            <div>
              <div className="text-2xl font-bold">${timeline.summary.totalCost.toFixed(4)}</div>
              <div className="text-xs text-muted-foreground">Cost</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Event Timeline */}
      <Card>
        <CardHeader>
          <CardTitle>Event Timeline</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {events.map((event, index) => (
              <ReplayEventItem
                key={event.id}
                event={event}
                isActive={index === currentEventIndex}
                isPast={index < currentEventIndex}
                isExpanded={expandedEvents.has(event.id)}
                onToggleExpand={() => toggleEventExpanded(event.id)}
              />
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

interface ReplayEventItemProps {
  event: ReplayEvent;
  isActive: boolean;
  isPast: boolean;
  isExpanded: boolean;
  onToggleExpand: () => void;
}

function ReplayEventItem({
  event,
  isActive,
  isPast,
  isExpanded,
  onToggleExpand,
}: ReplayEventItemProps) {
  const hasDetails = event.data.input || event.data.output || event.data.command;

  return (
    <div
      className={cn(
        "border rounded-lg p-3 transition-all",
        isActive && "ring-2 ring-primary bg-primary/5",
        isPast && !isActive && "opacity-60",
        !isPast && !isActive && "opacity-40"
      )}
    >
      <div
        className="flex items-center gap-3 cursor-pointer"
        onClick={hasDetails ? onToggleExpand : undefined}
      >
        {/* Expand/Collapse Icon */}
        {hasDetails && (
          <div className="flex-shrink-0">
            {isExpanded ? (
              <ChevronDown className="h-4 w-4 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            )}
          </div>
        )}

        {/* Event Icon */}
        <div className="flex-shrink-0">
          <EventIcon type={event.type} status={event.status} />
        </div>

        {/* Event Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium truncate">{event.title}</span>
            <EventTypeBadge type={event.type} />
            {event.status === "error" && (
              <Badge variant="destructive" className="text-xs">
                Error
              </Badge>
            )}
          </div>
          {event.description && (
            <p className="text-sm text-muted-foreground truncate">
              {event.description}
            </p>
          )}
        </div>

        {/* Metadata */}
        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          {event.data.model && (
            <Badge variant="outline" className="text-xs">
              {event.data.model}
            </Badge>
          )}
          {event.durationMs && <span>{event.durationMs}ms</span>}
          {event.data.cost && <span>${event.data.cost.toFixed(4)}</span>}
        </div>
      </div>

      {/* Expanded Details */}
      {isExpanded && hasDetails && (
        <div className="mt-3 pt-3 border-t">
          <EventDetails event={event} />
        </div>
      )}
    </div>
  );
}

function EventIcon({ type, status }: { type: ReplayEvent["type"]; status: string }) {
  const className = cn(
    "h-5 w-5",
    status === "error" && "text-red-500"
  );

  switch (type) {
    case "llm_call":
      return <Cpu className={cn(className, status !== "error" && "text-blue-500")} />;
    case "tool_call":
      return <Zap className={cn(className, status !== "error" && "text-green-500")} />;
    case "file_operation":
      return <FileCode className={cn(className, status !== "error" && "text-purple-500")} />;
    case "terminal_command":
      return <Terminal className={cn(className, status !== "error" && "text-orange-500")} />;
    case "checkpoint":
      return <Bookmark className={cn(className, status !== "error" && "text-cyan-500")} />;
    case "git_operation":
      return <GitBranch className={cn(className, status !== "error" && "text-pink-500")} />;
    case "error":
      return <AlertCircle className="h-5 w-5 text-red-500" />;
    default:
      return <MessageSquare className={cn(className, status !== "error" && "text-gray-500")} />;
  }
}

function EventTypeBadge({ type }: { type: ReplayEvent["type"] }) {
  const labels: Record<ReplayEvent["type"], string> = {
    llm_call: "LLM",
    tool_call: "Tool",
    file_operation: "File",
    terminal_command: "Terminal",
    checkpoint: "Checkpoint",
    git_operation: "Git",
    user_input: "User",
    agent_thought: "Agent",
    error: "Error",
  };

  return (
    <Badge variant="outline" className="text-xs">
      {labels[type] || type}
    </Badge>
  );
}

function EventDetails({ event }: { event: ReplayEvent }) {
  const { data } = event;

  return (
    <div className="space-y-3 text-sm">
      {/* LLM Call Details */}
      {event.type === "llm_call" && (
        <>
          {data.input && (
            <div>
              <div className="font-medium mb-1">Input</div>
              <pre className="bg-muted p-2 rounded text-xs overflow-auto max-h-48">
                {JSON.stringify(data.input, null, 2)}
              </pre>
            </div>
          )}
          {data.output && (
            <div>
              <div className="font-medium mb-1">Output</div>
              <pre className="bg-muted p-2 rounded text-xs overflow-auto max-h-48">
                {JSON.stringify(data.output, null, 2)}
              </pre>
            </div>
          )}
          {(data.tokensInput || data.tokensOutput) && (
            <div className="flex gap-4">
              <span>Input tokens: {data.tokensInput || 0}</span>
              <span>Output tokens: {data.tokensOutput || 0}</span>
            </div>
          )}
        </>
      )}

      {/* Tool Call Details */}
      {event.type === "tool_call" && (
        <>
          {data.arguments && (
            <div>
              <div className="font-medium mb-1">Arguments</div>
              <pre className="bg-muted p-2 rounded text-xs overflow-auto max-h-48">
                {JSON.stringify(data.arguments, null, 2)}
              </pre>
            </div>
          )}
          {data.result && (
            <div>
              <div className="font-medium mb-1">Result</div>
              <pre className="bg-muted p-2 rounded text-xs overflow-auto max-h-48">
                {JSON.stringify(data.result, null, 2)}
              </pre>
            </div>
          )}
        </>
      )}

      {/* Terminal Command Details */}
      {event.type === "terminal_command" && (
        <>
          <div>
            <div className="font-medium mb-1">Command</div>
            <pre className="bg-muted p-2 rounded text-xs font-mono">
              $ {data.command}
            </pre>
          </div>
          {data.stdout && (
            <div>
              <div className="font-medium mb-1">Output</div>
              <pre className="bg-muted p-2 rounded text-xs overflow-auto max-h-48 whitespace-pre-wrap">
                {data.stdout}
              </pre>
            </div>
          )}
          {data.stderr && (
            <div>
              <div className="font-medium mb-1 text-red-500">Error Output</div>
              <pre className="bg-red-50 dark:bg-red-900/20 p-2 rounded text-xs overflow-auto max-h-48 whitespace-pre-wrap text-red-600 dark:text-red-400">
                {data.stderr}
              </pre>
            </div>
          )}
          {data.exitCode !== undefined && (
            <div>
              Exit code:{" "}
              <Badge variant={data.exitCode === 0 ? "outline" : "destructive"}>
                {data.exitCode}
              </Badge>
            </div>
          )}
        </>
      )}

      {/* File Operation Details */}
      {event.type === "file_operation" && (
        <>
          <div className="flex gap-4">
            <span>File: {data.filePath}</span>
            <span>Operation: {data.operation}</span>
          </div>
          {data.diff && (
            <div>
              <div className="font-medium mb-1">Diff</div>
              <pre className="bg-muted p-2 rounded text-xs overflow-auto max-h-48 font-mono">
                {data.diff}
              </pre>
            </div>
          )}
        </>
      )}

      {/* Error Details */}
      {data.error && (
        <div>
          <div className="font-medium mb-1 text-red-500">Error</div>
          <pre className="bg-red-50 dark:bg-red-900/20 p-2 rounded text-xs overflow-auto max-h-48 text-red-600 dark:text-red-400">
            {data.error}
          </pre>
        </div>
      )}
    </div>
  );
}
