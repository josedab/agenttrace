"use client";

import * as React from "react";
import { useQueryClient } from "@tanstack/react-query";

interface RealtimeEvent {
  type: "trace:created" | "trace:updated" | "observation:created" | "score:created";
  data: any;
}

interface UseRealtimeOptions {
  enabled?: boolean;
  onEvent?: (event: RealtimeEvent) => void;
}

export function useRealtime(options: UseRealtimeOptions = {}) {
  const { enabled = true, onEvent } = options;
  const queryClient = useQueryClient();
  const [connected, setConnected] = React.useState(false);
  const [error, setError] = React.useState<Error | null>(null);
  const eventSourceRef = React.useRef<EventSource | null>(null);

  React.useEffect(() => {
    if (!enabled) {
      return;
    }

    const connect = () => {
      try {
        const eventSource = new EventSource("/api/events");
        eventSourceRef.current = eventSource;

        eventSource.onopen = () => {
          setConnected(true);
          setError(null);
        };

        eventSource.onerror = (e) => {
          setConnected(false);
          setError(new Error("Connection lost"));

          // Attempt to reconnect after 5 seconds
          setTimeout(() => {
            if (eventSourceRef.current === eventSource) {
              eventSource.close();
              connect();
            }
          }, 5000);
        };

        eventSource.onmessage = (e) => {
          try {
            const event: RealtimeEvent = JSON.parse(e.data);

            // Handle different event types
            switch (event.type) {
              case "trace:created":
              case "trace:updated":
                queryClient.invalidateQueries({ queryKey: ["traces"] });
                if (event.data.traceId) {
                  queryClient.invalidateQueries({
                    queryKey: ["trace", event.data.traceId],
                  });
                }
                break;

              case "observation:created":
                if (event.data.traceId) {
                  queryClient.invalidateQueries({
                    queryKey: ["trace-observations", event.data.traceId],
                  });
                }
                break;

              case "score:created":
                queryClient.invalidateQueries({ queryKey: ["scores"] });
                if (event.data.traceId) {
                  queryClient.invalidateQueries({
                    queryKey: ["trace-scores", event.data.traceId],
                  });
                }
                break;
            }

            // Call custom event handler
            onEvent?.(event);
          } catch (err) {
            console.error("Failed to parse SSE event:", err);
          }
        };

        return eventSource;
      } catch (err) {
        setError(err as Error);
        return null;
      }
    };

    const eventSource = connect();

    return () => {
      if (eventSource) {
        eventSource.close();
      }
      eventSourceRef.current = null;
    };
  }, [enabled, queryClient, onEvent]);

  return {
    connected,
    error,
    disconnect: () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
        setConnected(false);
      }
    },
  };
}

// Hook for subscribing to specific trace updates
export function useTraceRealtime(traceId: string, options: UseRealtimeOptions = {}) {
  const { onEvent, ...rest } = options;

  const handleEvent = React.useCallback(
    (event: RealtimeEvent) => {
      // Only handle events for this specific trace
      if (event.data.traceId === traceId) {
        onEvent?.(event);
      }
    },
    [traceId, onEvent]
  );

  return useRealtime({
    ...rest,
    onEvent: handleEvent,
  });
}

// Hook for subscribing to project-wide events
export function useProjectRealtime(projectId: string, options: UseRealtimeOptions = {}) {
  const { onEvent, ...rest } = options;

  const handleEvent = React.useCallback(
    (event: RealtimeEvent) => {
      // Only handle events for this project
      if (event.data.projectId === projectId) {
        onEvent?.(event);
      }
    },
    [projectId, onEvent]
  );

  return useRealtime({
    ...rest,
    onEvent: handleEvent,
  });
}

// Hook for live trace count
export function useLiveTraceCount() {
  const [count, setCount] = React.useState(0);

  useRealtime({
    onEvent: (event) => {
      if (event.type === "trace:created") {
        setCount((prev) => prev + 1);
      }
    },
  });

  return count;
}

// Hook for live cost tracking
export function useLiveCost() {
  const [cost, setCost] = React.useState(0);

  useRealtime({
    onEvent: (event) => {
      if (event.type === "observation:created" && event.data.cost) {
        setCost((prev) => prev + event.data.cost);
      }
    },
  });

  return cost;
}
