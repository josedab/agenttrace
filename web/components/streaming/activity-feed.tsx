"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { StreamActivity } from "@/hooks/use-streaming";
import {
  Bot,
  FileText,
  Terminal,
  GitBranch,
  AlertCircle,
  Cpu,
  Wrench,
} from "lucide-react";

interface ActivityFeedProps {
  activities: StreamActivity[];
  followMode?: boolean;
}

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

export function ActivityFeed({
  activities,
  followMode = true,
}: ActivityFeedProps) {
  const scrollRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (followMode && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [activities, followMode]);

  return (
    <ScrollArea className="h-[500px]" ref={scrollRef}>
      <div className="space-y-1 p-2">
        {activities.length === 0 && (
          <div className="flex items-center justify-center h-32 text-muted-foreground text-sm">
            Waiting for agent activity...
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
                  <span className="font-medium truncate">{activity.title}</span>
                  {activity.durationMs && (
                    <Badge
                      variant="outline"
                      className="text-[10px] px-1 py-0"
                    >
                      {activity.durationMs}ms
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
  );
}
