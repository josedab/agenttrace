'use client';

import * as React from 'react';
import { useQuery } from '@tanstack/react-query';
import { cn } from '@/lib/utils';
import { api } from '@/lib/api';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  ChevronRight,
  Terminal,
  FileCode,
  GitCommit,
  Brain,
  Wrench,
  AlertCircle,
  Clock,
  DollarSign,
} from 'lucide-react';

interface ReplayTimelineProps {
  traceId: string;
}

interface ReplayEvent {
  id: string;
  type: string;
  timestamp: string;
  durationMs?: number;
  title: string;
  description?: string;
  status: string;
  data: Record<string, unknown>;
  children?: ReplayEvent[];
}

interface ReplayTimeline {
  traceId: string;
  traceName: string;
  startTime: string;
  endTime?: string;
  durationMs: number;
  events: ReplayEvent[];
  summary: {
    totalEvents: number;
    llmCalls: number;
    toolCalls: number;
    fileOperations: number;
    terminalCommands: number;
    checkpoints: number;
    errors: number;
    totalTokens: number;
    totalCost: number;
  };
}

const eventIcons: Record<string, React.ElementType> = {
  llm_call: Brain,
  tool_call: Wrench,
  file_operation: FileCode,
  terminal_command: Terminal,
  checkpoint: GitCommit,
  git_operation: GitCommit,
  error: AlertCircle,
};

export function ReplayTimeline({ traceId }: ReplayTimelineProps) {
  const [currentStep, setCurrentStep] = React.useState(0);
  const [isPlaying, setIsPlaying] = React.useState(false);
  const [playbackSpeed, setPlaybackSpeed] = React.useState(1);

  const {
    data: timeline,
    isLoading,
    isError,
    refetch,
  } = useQuery<ReplayTimeline>({
    queryKey: ['replay-timeline', traceId],
    queryFn: () => api.traces.getReplay(traceId),
    enabled: !!traceId,
  });

  const { data: stepState } = useQuery({
    queryKey: ['replay-step', traceId, currentStep],
    queryFn: () => api.traces.getReplayStep(traceId, currentStep),
    enabled: !!traceId && currentStep >= 0,
  });

  React.useEffect(() => {
    if (!isPlaying || !timeline) return;

    const interval = Math.max(100, 1000 / playbackSpeed);
    const timer = setInterval(() => {
      setCurrentStep((prev) => {
        if (prev >= timeline.events.length - 1) {
          setIsPlaying(false);
          return prev;
        }
        return prev + 1;
      });
    }, interval);

    return () => clearInterval(timer);
  }, [isPlaying, timeline, playbackSpeed]);

  if (isLoading) {
    return <ReplayTimelineSkeleton />;
  }

  if (isError || !timeline) {
    return (
      <Card className="border-destructive/30">
        <CardContent className="flex min-h-48 flex-col items-center justify-center gap-3 text-center">
          <AlertCircle className="h-7 w-7 text-destructive" />
          <p className="font-medium">The replay timeline could not be built.</p>
          <Button variant="outline" onClick={() => refetch()}>
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  const events = timeline.events;
  if (events.length === 0) {
    return (
      <Card>
        <CardContent className="flex min-h-48 flex-col items-center justify-center text-center">
          <Clock className="h-7 w-7 text-muted-foreground" />
          <p className="mt-3 font-medium">No replay events were recorded</p>
          <p className="mt-1 text-sm text-muted-foreground">
            The trace exists, but it has no observations, file operations, commands, or checkpoints.
          </p>
        </CardContent>
      </Card>
    );
  }

  const currentEvent = events[currentStep];
  const progress = events.length > 0 ? ((currentStep + 1) / events.length) * 100 : 0;

  return (
    <div className="flex flex-col gap-4">
      {/* Summary Bar */}
      <div className="flex items-center gap-4 rounded-lg bg-muted/50 px-4 py-2 text-sm">
        <span className="font-medium">{timeline.traceName}</span>
        <Badge variant="outline">
          <Clock className="mr-1 h-3 w-3" />
          {formatDuration(timeline.durationMs)}
        </Badge>
        <Badge variant="outline">
          <Brain className="mr-1 h-3 w-3" />
          {timeline.summary.llmCalls} LLM calls
        </Badge>
        <Badge variant="outline">
          <DollarSign className="mr-1 h-3 w-3" />${timeline.summary.totalCost.toFixed(4)}
        </Badge>
        {timeline.summary.errors > 0 && (
          <Badge variant="destructive">{timeline.summary.errors} errors</Badge>
        )}
      </div>

      {/* Progress Bar */}
      <div className="px-4">
        <div className="h-1.5 w-full rounded-full bg-muted">
          <div
            className="h-1.5 rounded-full bg-primary transition-all duration-200"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      {/* Playback Controls */}
      <div className="flex items-center gap-2 px-4">
        <Button
          variant="outline"
          size="sm"
          onClick={() => setCurrentStep(0)}
          disabled={currentStep === 0}
        >
          <SkipBack className="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setCurrentStep(Math.max(0, currentStep - 1))}
          disabled={currentStep === 0}
        >
          <ChevronRight className="h-4 w-4 rotate-180" />
        </Button>
        <Button
          variant={isPlaying ? 'secondary' : 'default'}
          size="sm"
          onClick={() => setIsPlaying(!isPlaying)}
        >
          {isPlaying ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setCurrentStep(Math.min(events.length - 1, currentStep + 1))}
          disabled={currentStep >= events.length - 1}
        >
          <ChevronRight className="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setCurrentStep(events.length - 1)}
          disabled={currentStep >= events.length - 1}
        >
          <SkipForward className="h-4 w-4" />
        </Button>

        {/* Speed Control */}
        <div className="ml-2 flex items-center gap-1 rounded-md border px-2 py-1">
          {[0.5, 1, 2, 4, 8].map((speed) => (
            <button
              key={speed}
              onClick={() => setPlaybackSpeed(speed)}
              className={cn(
                'rounded px-1.5 py-0.5 text-xs',
                playbackSpeed === speed ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'
              )}
            >
              {speed}x
            </button>
          ))}
        </div>

        <span className="ml-2 text-sm text-muted-foreground">
          Step {currentStep + 1} / {events.length}
        </span>
      </div>

      <div className="grid grid-cols-3 gap-4">
        {/* Event List */}
        <div className="col-span-1 max-h-[600px] overflow-auto rounded-lg border">
          {events.map((event, idx) => {
            const Icon = eventIcons[event.type] || ChevronRight;
            return (
              <button
                key={event.id}
                onClick={() => setCurrentStep(idx)}
                className={cn(
                  'flex w-full items-start gap-3 border-b px-3 py-2 text-left text-sm transition-colors hover:bg-muted/50',
                  idx === currentStep && 'border-l-2 border-l-primary bg-primary/10',
                  event.status === 'error' && 'text-destructive'
                )}
              >
                <Icon className="mt-0.5 h-4 w-4 shrink-0" />
                <div className="min-w-0">
                  <div className="truncate font-medium">{event.title}</div>
                  {event.durationMs !== undefined && (
                    <div className="text-xs text-muted-foreground">
                      {formatDuration(event.durationMs)}
                    </div>
                  )}
                </div>
              </button>
            );
          })}
        </div>

        {/* Event Detail */}
        <div className="col-span-2 space-y-4">
          {currentEvent && (
            <Card>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">{currentEvent.title}</CardTitle>
                  <Badge variant={currentEvent.status === 'error' ? 'destructive' : 'outline'}>
                    {currentEvent.status}
                  </Badge>
                </div>
                {currentEvent.description && (
                  <p className="text-sm text-muted-foreground">{currentEvent.description}</p>
                )}
              </CardHeader>
              <CardContent>
                <EventDataView data={currentEvent.data} type={currentEvent.type} />
              </CardContent>
            </Card>
          )}

          {/* Cumulative Stats */}
          {stepState && (
            <div className="grid grid-cols-3 gap-2">
              <Card>
                <CardContent className="pt-4">
                  <div className="text-xs text-muted-foreground">Cost so far</div>
                  <div className="text-lg font-bold">${stepState.costSoFar.toFixed(4)}</div>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="pt-4">
                  <div className="text-xs text-muted-foreground">Tokens used</div>
                  <div className="text-lg font-bold">{stepState.tokensSoFar.toLocaleString()}</div>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="pt-4">
                  <div className="text-xs text-muted-foreground">Elapsed</div>
                  <div className="text-lg font-bold">{formatDuration(stepState.elapsedMs)}</div>
                </CardContent>
              </Card>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function EventDataView({ data, type }: { data: Record<string, unknown>; type: string }) {
  if (type === 'llm_call') {
    return (
      <div className="space-y-2">
        {typeof data.model === 'string' && (
          <div className="text-sm">
            <span className="text-muted-foreground">Model: </span>
            <span className="font-mono">{data.model}</span>
          </div>
        )}
        {data.input !== undefined && data.input !== null && (
          <div>
            <div className="mb-1 text-xs text-muted-foreground">Input</div>
            <pre className="max-h-40 overflow-auto rounded bg-muted p-2 text-xs">
              {typeof data.input === 'string' ? data.input : JSON.stringify(data.input, null, 2)}
            </pre>
          </div>
        )}
        {data.output !== undefined && data.output !== null && (
          <div>
            <div className="mb-1 text-xs text-muted-foreground">Output</div>
            <pre className="max-h-40 overflow-auto rounded bg-muted p-2 text-xs">
              {typeof data.output === 'string' ? data.output : JSON.stringify(data.output, null, 2)}
            </pre>
          </div>
        )}
      </div>
    );
  }

  if (type === 'file_operation') {
    return (
      <div className="space-y-2">
        <div className="font-mono text-sm">{String(data.filePath ?? '')}</div>
        {typeof data.diff === 'string' && (
          <pre className="max-h-60 overflow-auto rounded bg-muted p-2 font-mono text-xs">
            {data.diff}
          </pre>
        )}
      </div>
    );
  }

  if (type === 'terminal_command') {
    return (
      <div className="space-y-2">
        <div className="rounded bg-black p-2 font-mono text-sm text-green-400">
          $ {String(data.command ?? '')}
        </div>
        {typeof data.stdout === 'string' && (
          <pre className="max-h-40 overflow-auto rounded bg-muted p-2 text-xs">{data.stdout}</pre>
        )}
        {typeof data.stderr === 'string' && (
          <pre className="max-h-40 overflow-auto rounded bg-red-50 p-2 text-xs text-red-700 dark:bg-red-950 dark:text-red-300">
            {data.stderr}
          </pre>
        )}
      </div>
    );
  }

  // Generic JSON view for other types
  return (
    <pre className="max-h-60 overflow-auto rounded bg-muted p-2 text-xs">
      {JSON.stringify(data, null, 2)}
    </pre>
  );
}

function ReplayTimelineSkeleton() {
  return (
    <div className="space-y-4">
      <div className="h-10 animate-pulse rounded-lg bg-muted" />
      <div className="h-8 w-64 animate-pulse rounded bg-muted" />
      <div className="grid grid-cols-3 gap-4">
        <div className="h-96 animate-pulse rounded-lg bg-muted" />
        <div className="col-span-2 h-96 animate-pulse rounded-lg bg-muted" />
      </div>
    </div>
  );
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`;
}
