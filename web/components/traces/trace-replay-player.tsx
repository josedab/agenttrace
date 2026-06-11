"use client";

import * as React from "react";
import { useState, useEffect, useCallback, useRef, useMemo } from "react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Slider } from "@/components/ui/slider";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
  Wifi,
  WifiOff,
  Eye,
  EyeOff,
  Search,
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
  input?: unknown;
  output?: unknown;
  tokensInput?: number;
  tokensOutput?: number;
  cost?: number;
  toolName?: string;
  arguments?: unknown;
  result?: unknown;
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

type EventTypeFilter = ReplayEvent["type"];

const ALL_EVENT_TYPES: EventTypeFilter[] = [
  "llm_call", "tool_call", "file_operation", "terminal_command",
  "checkpoint", "git_operation", "user_input", "agent_thought", "error",
];

interface TraceReplayPlayerProps {
  timeline: ReplayTimeline;
  onExport?: () => void;
  websocketUrl?: string;
  enableLiveStream?: boolean;
}

export function TraceReplayPlayer({ timeline, onExport, websocketUrl, enableLiveStream }: TraceReplayPlayerProps) {
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentEventIndex, setCurrentEventIndex] = useState(-1);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);
  const [elapsedTime, setElapsedTime] = useState(0);
  const [expandedEvents, setExpandedEvents] = useState<Set<string>>(new Set());
  const [activeFilters, setActiveFilters] = useState<Set<EventTypeFilter>>(new Set());
  const [searchQuery, setSearchQuery] = useState("");
  const [isLiveConnected, setIsLiveConnected] = useState(false);
  const [liveEvents, setLiveEvents] = useState<ReplayEvent[]>([]);
  const [activeView, setActiveView] = useState<"timeline" | "waterfall">("timeline");
  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const timelineRef = useRef<HTMLDivElement>(null);

  // Merge static events with live-streamed events
  const allEvents = useMemo(() => {
    return [...timeline.events, ...liveEvents];
  }, [timeline.events, liveEvents]);

  // Apply filters and search
  const filteredEvents = useMemo(() => {
    let events = allEvents;
    if (activeFilters.size > 0) {
      events = events.filter((e) => activeFilters.has(e.type));
    }
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      events = events.filter((e) =>
        e.title.toLowerCase().includes(query) ||
        e.description?.toLowerCase().includes(query) ||
        e.data.model?.toLowerCase().includes(query) ||
        e.data.filePath?.toLowerCase().includes(query) ||
        e.data.command?.toLowerCase().includes(query)
      );
    }
    return events;
  }, [allEvents, activeFilters, searchQuery]);

  const totalDuration = timeline.durationMs;
  const events = filteredEvents;

  // Calculate event times relative to start
  const eventTimes = useMemo(() => events.map((event) => {
    const eventTime = new Date(event.timestamp).getTime();
    const startTime = new Date(timeline.startTime).getTime();
    return eventTime - startTime;
  }), [events, timeline.startTime]);

  // Running summary up to current event
  const runningSummary = useMemo(() => {
    const eventsUpToCurrent = events.slice(0, Math.max(0, currentEventIndex + 1));
    return {
      tokens: eventsUpToCurrent.reduce((sum, e) => sum + (e.data.tokensInput || 0) + (e.data.tokensOutput || 0), 0),
      cost: eventsUpToCurrent.reduce((sum, e) => sum + (e.data.cost || 0), 0),
      llmCalls: eventsUpToCurrent.filter((e) => e.type === "llm_call").length,
      toolCalls: eventsUpToCurrent.filter((e) => e.type === "tool_call").length,
      errors: eventsUpToCurrent.filter((e) => e.status === "error").length,
    };
  }, [events, currentEventIndex]);

  // WebSocket for live streaming
  useEffect(() => {
    if (!enableLiveStream || !websocketUrl) return;

    const ws = new WebSocket(websocketUrl);
    wsRef.current = ws;

    ws.onopen = () => setIsLiveConnected(true);
    ws.onclose = () => setIsLiveConnected(false);
    ws.onerror = () => setIsLiveConnected(false);

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === "event" && data.event) {
          setLiveEvents((prev) => [...prev, data.event as ReplayEvent]);
        }
      } catch {
        // Ignore malformed messages
      }
    };

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, [enableLiveStream, websocketUrl]);

  // Update current event based on elapsed time
  useEffect(() => {
    const currentEvent = eventTimes.findIndex(
      (time, index) =>
        time <= elapsedTime &&
        (index === eventTimes.length - 1 || eventTimes[index + 1] > elapsedTime)
    );
    setCurrentEventIndex(currentEvent);
  }, [elapsedTime, eventTimes]);

  // Auto-scroll to current event
  useEffect(() => {
    if (currentEventIndex >= 0 && timelineRef.current) {
      const eventEl = timelineRef.current.children[currentEventIndex] as HTMLElement;
      if (eventEl) {
        eventEl.scrollIntoView({ behavior: "smooth", block: "nearest" });
      }
    }
  }, [currentEventIndex]);

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

  // Keep the latest playback control callbacks available to the keydown
  // listener below without re-subscribing on every state change (these
  // callbacks change identity frequently while playing back).
  const playbackControlsRef = useRef({
    handlePlayPause,
    handleSkipBack,
    handleSkipForward,
    handleReset,
  });
  useEffect(() => {
    playbackControlsRef.current = {
      handlePlayPause,
      handleSkipBack,
      handleSkipForward,
      handleReset,
    };
  }, [handlePlayPause, handleSkipBack, handleSkipForward, handleReset]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;
      switch (e.key) {
        case " ":
          e.preventDefault();
          playbackControlsRef.current.handlePlayPause();
          break;
        case "ArrowLeft":
          e.preventDefault();
          playbackControlsRef.current.handleSkipBack();
          break;
        case "ArrowRight":
          e.preventDefault();
          playbackControlsRef.current.handleSkipForward();
          break;
        case "Home":
          e.preventDefault();
          playbackControlsRef.current.handleReset();
          break;
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

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

  const toggleFilter = useCallback((type: EventTypeFilter) => {
    setActiveFilters((prev) => {
      const next = new Set(prev);
      if (next.has(type)) {
        next.delete(type);
      } else {
        next.add(type);
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
      {/* Live Connection Status */}
      {enableLiveStream && (
        <div className={cn(
          "flex items-center gap-2 px-3 py-1.5 rounded-md text-sm",
          isLiveConnected ? "bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-400" : "bg-muted text-muted-foreground"
        )}>
          {isLiveConnected ? <Wifi className="h-4 w-4" /> : <WifiOff className="h-4 w-4" />}
          {isLiveConnected ? "Live stream connected" : "Disconnected"}
          {liveEvents.length > 0 && (
            <Badge variant="secondary" className="ml-2">{liveEvents.length} new events</Badge>
          )}
        </div>
      )}

      {/* Playback Controls */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center gap-4">
            {/* Control Buttons */}
            <div className="flex items-center gap-2">
              <Button variant="outline" size="icon" aria-label="Reset replay" onClick={handleReset}>
                <RotateCcw className="h-4 w-4" />
              </Button>
              <Button variant="outline" size="icon" aria-label="Skip back" onClick={handleSkipBack}>
                <SkipBack className="h-4 w-4" />
              </Button>
              <Button size="icon" aria-label={isPlaying ? "Pause" : "Play"} onClick={handlePlayPause}>
                {isPlaying ? (
                  <Pause className="h-4 w-4" />
                ) : (
                  <Play className="h-4 w-4" />
                )}
              </Button>
              <Button variant="outline" size="icon" aria-label="Skip forward" onClick={handleSkipForward}>
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
                <option value={0.25}>0.25x</option>
                <option value={0.5}>0.5x</option>
                <option value={1}>1x</option>
                <option value={2}>2x</option>
                <option value={4}>4x</option>
                <option value={8}>8x</option>
                <option value={16}>16x</option>
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

          {/* Keyboard shortcuts hint */}
          <div className="mt-2 text-xs text-muted-foreground">
            Space: play/pause · ←/→: step · Home: reset
          </div>
        </CardContent>
      </Card>

      {/* Filter Bar and Search */}
      <Card>
        <CardContent className="pt-4 pb-3">
          <div className="flex items-center gap-3 flex-wrap">
            {/* Search */}
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search events..."
                className="w-full pl-9 pr-3 py-2 text-sm border rounded-md bg-background"
              />
            </div>

            <Separator orientation="vertical" className="h-6" />

            {/* Event Type Filters */}
            {ALL_EVENT_TYPES.map((type) => (
              <button
                key={type}
                onClick={() => toggleFilter(type)}
                className={cn(
                  "flex items-center gap-1 px-2 py-1 text-xs rounded-full border transition-colors",
                  activeFilters.size === 0 || activeFilters.has(type)
                    ? "bg-primary/10 border-primary/30"
                    : "bg-muted/50 border-transparent text-muted-foreground"
                )}
              >
                {activeFilters.has(type) ? <Eye className="h-3 w-3" /> : <EyeOff className="h-3 w-3" />}
                {EVENT_TYPE_LABELS[type]}
              </button>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Running Summary Stats */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid grid-cols-7 gap-4 text-center">
            <div>
              <div className="text-2xl font-bold">{events.length}</div>
              <div className="text-xs text-muted-foreground">Events{activeFilters.size > 0 ? " (filtered)" : ""}</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-blue-500">{runningSummary.llmCalls}</div>
              <div className="text-xs text-muted-foreground">LLM Calls</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-green-500">{runningSummary.toolCalls}</div>
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
              <div className="text-2xl font-bold">{runningSummary.tokens.toLocaleString()}</div>
              <div className="text-xs text-muted-foreground">Tokens</div>
            </div>
            <div>
              <div className="text-2xl font-bold">${runningSummary.cost.toFixed(4)}</div>
              <div className="text-xs text-muted-foreground">Cost</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Event Timeline with View Toggle */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Event Timeline</CardTitle>
            <Tabs value={activeView} onValueChange={(v) => setActiveView(v as "timeline" | "waterfall")}>
              <TabsList>
                <TabsTrigger value="timeline">Timeline</TabsTrigger>
                <TabsTrigger value="waterfall">Waterfall</TabsTrigger>
              </TabsList>
            </Tabs>
          </div>
        </CardHeader>
        <CardContent>
          {activeView === "timeline" ? (
            <div className="space-y-2" ref={timelineRef}>
              {events.map((event, index) => (
                <ReplayEventItem
                  key={event.id}
                  event={event}
                  isActive={index === currentEventIndex}
                  isPast={index < currentEventIndex}
                  isExpanded={expandedEvents.has(event.id)}
                  onToggleExpand={() => toggleEventExpanded(event.id)}
                  onJumpTo={() => {
                    const time = eventTimes[index];
                    if (time !== undefined) setElapsedTime(time);
                  }}
                />
              ))}
            </div>
          ) : (
            <WaterfallView
              events={events}
              eventTimes={eventTimes}
              totalDuration={totalDuration}
              currentEventIndex={currentEventIndex}
              onEventClick={(index) => {
                const time = eventTimes[index];
                if (time !== undefined) setElapsedTime(time);
              }}
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}

const EVENT_TYPE_LABELS: Record<string, string> = {
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

// Waterfall view showing event durations as horizontal bars
function WaterfallView({
  events,
  eventTimes,
  totalDuration,
  currentEventIndex,
  onEventClick,
}: {
  events: ReplayEvent[];
  eventTimes: number[];
  totalDuration: number;
  currentEventIndex: number;
  onEventClick: (index: number) => void;
}) {
  if (totalDuration === 0) return null;

  return (
    <div className="space-y-1">
      {events.map((event, index) => {
        const start = eventTimes[index] || 0;
        const duration = event.durationMs || 100;
        const leftPercent = (start / totalDuration) * 100;
        const widthPercent = Math.max(0.5, (duration / totalDuration) * 100);
        const colorClass = EVENT_BAR_COLORS[event.type] || "bg-gray-400";

        return (
          <div
            key={event.id}
            className={cn(
              "flex items-center gap-2 h-6 cursor-pointer hover:bg-muted/50 rounded px-1",
              index === currentEventIndex && "ring-1 ring-primary bg-primary/5"
            )}
            onClick={() => onEventClick(index)}
          >
            <span className="text-xs text-muted-foreground w-24 truncate">{event.title}</span>
            <div className="flex-1 relative h-4">
              <div
                className={cn("absolute h-full rounded-sm", colorClass, event.status === "error" && "bg-red-500")}
                style={{
                  left: `${leftPercent}%`,
                  width: `${widthPercent}%`,
                  minWidth: "4px",
                }}
              />
            </div>
            <span className="text-xs text-muted-foreground w-16 text-right">
              {event.durationMs ? `${event.durationMs}ms` : "—"}
            </span>
          </div>
        );
      })}
    </div>
  );
}

const EVENT_BAR_COLORS: Record<string, string> = {
  llm_call: "bg-blue-500",
  tool_call: "bg-green-500",
  file_operation: "bg-purple-500",
  terminal_command: "bg-orange-500",
  checkpoint: "bg-cyan-500",
  git_operation: "bg-pink-500",
  user_input: "bg-gray-500",
  agent_thought: "bg-indigo-500",
  error: "bg-red-500",
};

interface ReplayEventItemProps {
  event: ReplayEvent;
  isActive: boolean;
  isPast: boolean;
  isExpanded: boolean;
  onToggleExpand: () => void;
  onJumpTo: () => void;
}

function ReplayEventItem({
  event,
  isActive,
  isPast,
  isExpanded,
  onToggleExpand,
  onJumpTo,
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
        onClick={hasDetails ? onToggleExpand : onJumpTo}
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
  return (
    <Badge variant="outline" className="text-xs">
      {EVENT_TYPE_LABELS[type] || type}
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
