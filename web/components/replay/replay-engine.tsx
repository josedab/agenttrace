"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { cn } from "@/lib/utils";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  ChevronRight,
  Terminal,
  FileCode,
  Brain,
  AlertCircle,
  Clock,
  DollarSign,
  MessageSquare,
  GitBranch,
  Filter,
  Zap,
  Send,
} from "lucide-react";

// --- Types ---

interface UnifiedTimelineEvent {
  id: string;
  sessionId: string;
  eventType: string;
  category: string;
  title: string;
  description?: string;
  startTime: string;
  endTime?: string;
  durationMs: number;
  parentId?: string;
  depth: number;
  metadata?: Record<string, unknown>;
  fileDelta?: {
    filePath: string;
    changeType: string;
    linesAdded: number;
    linesRemoved: number;
    diffPreview?: string;
  };
  costUsd: number;
  tokensUsed: number;
  model?: string;
  status: string;
  annotations?: ReplayAnnotation[];
}

interface ReplayAnnotation {
  id: string;
  userId: string;
  userName: string;
  content: string;
  timestamp: string;
  eventId: string;
}

interface ReplaySnapshot {
  eventIndex: number;
  timestamp: string;
  fileStates: Record<string, string>;
  activeModel: string;
  totalCost: number;
  totalTokens: number;
  elapsedMs: number;
  eventCounts: Record<string, number>;
}

// --- Constants ---

const SPEED_OPTIONS = [0.5, 1, 2, 4, 8, 16];

const EVENT_TYPE_CONFIG: Record<string, { icon: React.ElementType; label: string; color: string }> = {
  llm_call: { icon: Brain, label: "LLM Call", color: "text-purple-500" },
  file_edit: { icon: FileCode, label: "File Edit", color: "text-blue-500" },
  terminal_cmd: { icon: Terminal, label: "Terminal", color: "text-green-500" },
  decision_point: { icon: GitBranch, label: "Decision", color: "text-amber-500" },
  checkpoint: { icon: Zap, label: "Checkpoint", color: "text-cyan-500" },
  error: { icon: AlertCircle, label: "Error", color: "text-red-500" },
  tool_call: { icon: Zap, label: "Tool Call", color: "text-indigo-500" },
};

// --- Component ---

interface ReplayEngineProps {
  sessionId: string;
}

export function ReplayEngine({ sessionId }: ReplayEngineProps) {
  const queryClient = useQueryClient();
  const [currentStep, setCurrentStep] = React.useState(0);
  const [isPlaying, setIsPlaying] = React.useState(false);
  const [playbackSpeed, setPlaybackSpeed] = React.useState(1);
  const [activeFilters, setActiveFilters] = React.useState<string[]>([]);
  const [annotationText, setAnnotationText] = React.useState("");
  const [detailTab, setDetailTab] = React.useState("detail");

  // Fetch unified timeline
  const { data: timelineData, isLoading } = useQuery({
    queryKey: ["unified-timeline", sessionId],
    queryFn: () =>
      api.replaySessions.getUnifiedTimeline(sessionId) as Promise<{ events: UnifiedTimelineEvent[] }>,
    enabled: !!sessionId,
  });

  // Fetch snapshot at current step
  const { data: snapshot } = useQuery<ReplaySnapshot>({
    queryKey: ["replay-snapshot", sessionId, currentStep],
    queryFn: () =>
      api.replaySessions.getSnapshot(sessionId, currentStep) as Promise<ReplaySnapshot>,
    enabled: !!sessionId,
  });

  // Add annotation mutation
  const addAnnotation = useMutation({
    mutationFn: (data: { eventId: string; content: string }) =>
      api.replaySessions.addAnnotation(sessionId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["unified-timeline", sessionId] });
      setAnnotationText("");
    },
  });

  const events: UnifiedTimelineEvent[] = React.useMemo(
    () => timelineData?.events ?? [],
    [timelineData]
  );

  // Filter events
  const filteredEvents = React.useMemo(() => {
    if (activeFilters.length === 0) return events;
    return events.filter((e) => activeFilters.includes(e.eventType));
  }, [events, activeFilters]);

  const currentEvent = filteredEvents[currentStep] ?? null;

  // Playback timer
  React.useEffect(() => {
    if (!isPlaying || filteredEvents.length === 0) return;

    const currentDuration = filteredEvents[currentStep]?.durationMs ?? 1000;
    const interval = Math.max(50, currentDuration / playbackSpeed);
    const timer = setTimeout(() => {
      setCurrentStep((prev) => {
        if (prev >= filteredEvents.length - 1) {
          setIsPlaying(false);
          return prev;
        }
        return prev + 1;
      });
    }, interval);

    return () => clearTimeout(timer);
  }, [isPlaying, currentStep, filteredEvents, playbackSpeed]);

  // Running totals
  const runningCost = React.useMemo(() => {
    return filteredEvents.slice(0, currentStep + 1).reduce((sum, e) => sum + (e.costUsd || 0), 0);
  }, [filteredEvents, currentStep]);

  const runningTokens = React.useMemo(() => {
    return filteredEvents.slice(0, currentStep + 1).reduce((sum, e) => sum + (e.tokensUsed || 0), 0);
  }, [filteredEvents, currentStep]);

  const toggleFilter = (type: string) => {
    setActiveFilters((prev) =>
      prev.includes(type) ? prev.filter((t) => t !== type) : [...prev, type]
    );
    setCurrentStep(0);
  };

  if (isLoading) {
    return <ReplayEngineSkeleton />;
  }

  if (events.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p className="text-lg font-medium">No timeline events found</p>
        <p className="text-sm mt-2">This session has no recorded events to replay.</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {/* Top Stats Bar */}
      <div className="flex items-center gap-3 px-4 py-2 bg-muted/50 rounded-lg text-sm flex-wrap">
        <Badge variant="outline">
          <Clock className="h-3 w-3 mr-1" />
          {filteredEvents.length} events
        </Badge>
        <Badge variant="outline">
          <DollarSign className="h-3 w-3 mr-1" />
          ${runningCost.toFixed(4)}
        </Badge>
        <Badge variant="outline">
          <Brain className="h-3 w-3 mr-1" />
          {runningTokens.toLocaleString()} tokens
        </Badge>
        {snapshot && (
          <Badge variant="outline">
            <Clock className="h-3 w-3 mr-1" />
            {formatDuration(snapshot.elapsedMs)} elapsed
          </Badge>
        )}
      </div>

      {/* Progress / Scrubber */}
      <div className="px-4">
        <div className="relative w-full">
          <input
            type="range"
            min={0}
            max={Math.max(0, filteredEvents.length - 1)}
            value={currentStep}
            onChange={(e) => {
              setCurrentStep(Number(e.target.value));
              setIsPlaying(false);
            }}
            className="w-full h-2 bg-muted rounded-full appearance-none cursor-pointer accent-primary"
          />
          {/* Event markers */}
          <div className="absolute top-3 left-0 right-0 flex justify-between pointer-events-none">
            {filteredEvents.map((event, idx) => {
              const config = EVENT_TYPE_CONFIG[event.eventType];
              const left = filteredEvents.length > 1 ? (idx / (filteredEvents.length - 1)) * 100 : 50;
              return (
                <div
                  key={event.id}
                  className={cn(
                    "absolute w-1.5 h-1.5 rounded-full -translate-x-1/2",
                    event.status === "error" ? "bg-red-500" : config?.color ? "bg-current" : "bg-muted-foreground",
                    idx === currentStep && "w-2.5 h-2.5 ring-2 ring-primary"
                  )}
                  style={{ left: `${left}%` }}
                />
              );
            })}
          </div>
        </div>
      </div>

      {/* Playback Controls */}
      <div className="flex items-center gap-2 px-4 flex-wrap">
        <Button variant="outline" size="sm" onClick={() => { setCurrentStep(0); setIsPlaying(false); }} disabled={currentStep === 0}>
          <SkipBack className="h-4 w-4" />
        </Button>
        <Button variant="outline" size="sm" onClick={() => { setCurrentStep(Math.max(0, currentStep - 1)); setIsPlaying(false); }} disabled={currentStep === 0}>
          <ChevronRight className="h-4 w-4 rotate-180" />
        </Button>
        <Button variant={isPlaying ? "secondary" : "default"} size="sm" onClick={() => setIsPlaying(!isPlaying)}>
          {isPlaying ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
        </Button>
        <Button variant="outline" size="sm" onClick={() => { setCurrentStep(Math.min(filteredEvents.length - 1, currentStep + 1)); setIsPlaying(false); }} disabled={currentStep >= filteredEvents.length - 1}>
          <ChevronRight className="h-4 w-4" />
        </Button>
        <Button variant="outline" size="sm" onClick={() => { setCurrentStep(filteredEvents.length - 1); setIsPlaying(false); }} disabled={currentStep >= filteredEvents.length - 1}>
          <SkipForward className="h-4 w-4" />
        </Button>

        {/* Speed Selector */}
        <Select value={String(playbackSpeed)} onValueChange={(v) => setPlaybackSpeed(Number(v))}>
          <SelectTrigger className="w-24 h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {SPEED_OPTIONS.map((speed) => (
              <SelectItem key={speed} value={String(speed)}>
                {speed}x
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <span className="text-sm text-muted-foreground ml-2">
          Step {currentStep + 1} / {filteredEvents.length}
        </span>

        {/* Filters */}
        <div className="flex items-center gap-1 ml-auto">
          <Filter className="h-4 w-4 text-muted-foreground" />
          {Object.entries(EVENT_TYPE_CONFIG).map(([type, config]) => (
            <button
              key={type}
              onClick={() => toggleFilter(type)}
              className={cn(
                "px-2 py-0.5 text-xs rounded-full border transition-colors",
                activeFilters.length === 0 || activeFilters.includes(type)
                  ? "bg-primary/10 border-primary/30 text-foreground"
                  : "bg-muted/50 border-transparent text-muted-foreground"
              )}
            >
              {config.label}
            </button>
          ))}
        </div>
      </div>

      {/* Split Pane: Timeline (left) + Detail (right) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Left Panel: Event List */}
        <ScrollArea className="border rounded-lg h-[600px]">
          <div className="divide-y">
            {filteredEvents.map((event, idx) => {
              const config = EVENT_TYPE_CONFIG[event.eventType] ?? { icon: ChevronRight, label: event.eventType, color: "text-muted-foreground" };
              const Icon = config.icon;
              return (
                <button
                  key={event.id}
                  onClick={() => { setCurrentStep(idx); setIsPlaying(false); }}
                  className={cn(
                    "w-full flex items-start gap-3 px-3 py-2.5 text-left text-sm hover:bg-muted/50 transition-colors",
                    idx === currentStep && "bg-primary/10 border-l-2 border-l-primary",
                    event.status === "error" && "text-destructive"
                  )}
                  style={{ paddingLeft: `${12 + event.depth * 16}px` }}
                >
                  <Icon className={cn("h-4 w-4 mt-0.5 shrink-0", config.color)} />
                  <div className="min-w-0 flex-1">
                    <div className="font-medium truncate">{event.title}</div>
                    <div className="flex items-center gap-2 mt-0.5">
                      {event.durationMs > 0 && (
                        <span className="text-xs text-muted-foreground">{formatDuration(event.durationMs)}</span>
                      )}
                      {event.costUsd > 0 && (
                        <span className="text-xs text-muted-foreground">${event.costUsd.toFixed(4)}</span>
                      )}
                      {event.annotations && event.annotations.length > 0 && (
                        <MessageSquare className="h-3 w-3 text-muted-foreground" />
                      )}
                    </div>
                  </div>
                  <Badge variant={event.status === "error" ? "destructive" : "outline"} className="text-[10px] shrink-0">
                    {event.status}
                  </Badge>
                </button>
              );
            })}
          </div>
        </ScrollArea>

        {/* Right Panel: Event Detail */}
        <div className="lg:col-span-2 space-y-4">
          {currentEvent ? (
            <Card>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">{currentEvent.title}</CardTitle>
                  <div className="flex items-center gap-2">
                    {currentEvent.model && (
                      <Badge variant="outline" className="font-mono text-xs">{currentEvent.model}</Badge>
                    )}
                    <Badge variant={currentEvent.status === "error" ? "destructive" : "outline"}>
                      {currentEvent.status}
                    </Badge>
                  </div>
                </div>
                {currentEvent.description && (
                  <p className="text-sm text-muted-foreground">{currentEvent.description}</p>
                )}
              </CardHeader>
              <CardContent>
                <Tabs value={detailTab} onValueChange={setDetailTab}>
                  <TabsList>
                    <TabsTrigger value="detail">Detail</TabsTrigger>
                    <TabsTrigger value="annotations">
                      Annotations {currentEvent.annotations?.length ? `(${currentEvent.annotations.length})` : ""}
                    </TabsTrigger>
                  </TabsList>

                  <TabsContent value="detail" className="mt-3">
                    <EventDetailView event={currentEvent} />
                  </TabsContent>

                  <TabsContent value="annotations" className="mt-3">
                    <div className="space-y-3">
                      {currentEvent.annotations?.map((ann) => (
                        <div key={ann.id} className="border rounded-lg p-3">
                          <div className="flex items-center justify-between text-xs text-muted-foreground mb-1">
                            <span>{ann.userName || "User"}</span>
                            <span>{new Date(ann.timestamp).toLocaleString()}</span>
                          </div>
                          <p className="text-sm">{ann.content}</p>
                        </div>
                      ))}

                      {/* Add annotation form */}
                      <div className="flex gap-2">
                        <input
                          type="text"
                          value={annotationText}
                          onChange={(e) => setAnnotationText(e.target.value)}
                          placeholder="Add an annotation..."
                          className="flex-1 px-3 py-2 text-sm border rounded-md bg-background"
                          onKeyDown={(e) => {
                            if (e.key === "Enter" && annotationText.trim()) {
                              addAnnotation.mutate({ eventId: currentEvent.id, content: annotationText.trim() });
                            }
                          }}
                        />
                        <Button
                          size="sm"
                          disabled={!annotationText.trim() || addAnnotation.isPending}
                          onClick={() => addAnnotation.mutate({ eventId: currentEvent.id, content: annotationText.trim() })}
                        >
                          <Send className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  </TabsContent>
                </Tabs>
              </CardContent>
            </Card>
          ) : (
            <Card>
              <CardContent className="py-8 text-center text-muted-foreground">
                Select an event to view details
              </CardContent>
            </Card>
          )}

          {/* Running Counters */}
          <div className="grid grid-cols-4 gap-2">
            <Card>
              <CardContent className="pt-4 pb-3">
                <div className="text-xs text-muted-foreground">Cost so far</div>
                <div className="text-lg font-bold">${runningCost.toFixed(4)}</div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-4 pb-3">
                <div className="text-xs text-muted-foreground">Tokens used</div>
                <div className="text-lg font-bold">{runningTokens.toLocaleString()}</div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-4 pb-3">
                <div className="text-xs text-muted-foreground">Active model</div>
                <div className="text-lg font-bold font-mono">{snapshot?.activeModel || "—"}</div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-4 pb-3">
                <div className="text-xs text-muted-foreground">Elapsed</div>
                <div className="text-lg font-bold">{formatDuration(snapshot?.elapsedMs ?? 0)}</div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}

// --- Detail view per event type ---

function EventDetailView({ event }: { event: UnifiedTimelineEvent }) {
  if (event.eventType === "llm_call" || event.eventType === "decision_point") {
    return (
      <div className="space-y-3">
        {event.model && (
          <div className="text-sm">
            <span className="text-muted-foreground">Model: </span>
            <span className="font-mono">{event.model}</span>
            {event.tokensUsed > 0 && (
              <span className="text-muted-foreground ml-3">{event.tokensUsed.toLocaleString()} tokens</span>
            )}
            {event.costUsd > 0 && (
              <span className="text-muted-foreground ml-3">${event.costUsd.toFixed(4)}</span>
            )}
          </div>
        )}
        {event.description && (
          <pre className="text-xs bg-muted p-3 rounded max-h-60 overflow-auto whitespace-pre-wrap">
            {event.description}
          </pre>
        )}
        {event.metadata && Object.keys(event.metadata).length > 0 && (
          <div>
            <div className="text-xs text-muted-foreground mb-1">Metadata</div>
            <pre className="text-xs bg-muted p-2 rounded max-h-40 overflow-auto">
              {JSON.stringify(event.metadata, null, 2)}
            </pre>
          </div>
        )}
      </div>
    );
  }

  if (event.eventType === "file_edit") {
    return (
      <div className="space-y-2">
        {event.fileDelta && (
          <>
            <div className="flex items-center gap-2">
              <span className="font-mono text-sm">{event.fileDelta.filePath}</span>
              <Badge variant="outline">{event.fileDelta.changeType}</Badge>
            </div>
            <div className="flex gap-3 text-sm">
              <span className="text-green-600">+{event.fileDelta.linesAdded}</span>
              <span className="text-red-600">-{event.fileDelta.linesRemoved}</span>
            </div>
            {event.fileDelta.diffPreview && (
              <pre className="text-xs bg-muted p-3 rounded max-h-60 overflow-auto font-mono">
                {event.fileDelta.diffPreview}
              </pre>
            )}
          </>
        )}
      </div>
    );
  }

  if (event.eventType === "terminal_cmd") {
    return (
      <div className="space-y-2">
        {event.description && (
          <div className="font-mono text-sm bg-black text-green-400 p-3 rounded">
            $ {event.description}
          </div>
        )}
        <div className="text-sm text-muted-foreground">
          Duration: {formatDuration(event.durationMs)}
        </div>
      </div>
    );
  }

  // Generic view
  return (
    <div className="space-y-2">
      {event.description && <p className="text-sm">{event.description}</p>}
      {event.metadata && Object.keys(event.metadata).length > 0 && (
        <pre className="text-xs bg-muted p-2 rounded max-h-60 overflow-auto">
          {JSON.stringify(event.metadata, null, 2)}
        </pre>
      )}
    </div>
  );
}

function ReplayEngineSkeleton() {
  return (
    <div className="space-y-4">
      <div className="h-10 bg-muted animate-pulse rounded-lg" />
      <div className="h-2 bg-muted animate-pulse rounded w-full" />
      <div className="h-8 bg-muted animate-pulse rounded w-80" />
      <div className="grid grid-cols-3 gap-4">
        <div className="h-[600px] bg-muted animate-pulse rounded-lg" />
        <div className="col-span-2 h-[600px] bg-muted animate-pulse rounded-lg" />
      </div>
    </div>
  );
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`;
}
