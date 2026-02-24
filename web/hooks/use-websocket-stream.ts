"use client";

import { useEffect, useRef, useState, useCallback } from "react";

export interface WSLiveMetrics {
  traceId: string;
  activeSpans: number;
  completedSpans: number;
  totalTokens: number;
  totalCost: number;
  errorCount: number;
  elapsedMs: number;
  tokensPerSecond: number;
  costPerMinute: number;
  filesModified: number;
  terminalCommands: number;
  lastUpdated: string;
}

export interface WSStreamActivity {
  id: string;
  traceId: string;
  type: string;
  title: string;
  description?: string;
  timestamp: string;
  durationMs?: number;
  metadata?: Record<string, unknown>;
  status: string;
}

interface WSMessage {
  type: string;
  traceId: string;
  data: any;
  timestamp: string;
}

export function useWebSocketStream(traceId: string | null) {
  const [metrics, setMetrics] = useState<WSLiveMetrics | null>(null);
  const [activities, setActivities] = useState<WSStreamActivity[]>([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const connect = useCallback(() => {
    if (!traceId) return;

    const wsUrl =
      process.env.NEXT_PUBLIC_WS_URL ??
      (typeof window !== "undefined"
        ? `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}`
        : "ws://localhost:8080");

    const ws = new WebSocket(`${wsUrl}/ws/streaming/${traceId}`);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      setError(null);
      ws.send(JSON.stringify({ action: "subscribe", traceId }));
    };

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data);
        switch (msg.type) {
          case "metrics":
            setMetrics(msg.data as WSLiveMetrics);
            break;
          case "activity":
            setActivities((prev) => [
              ...prev.slice(-499),
              msg.data as WSStreamActivity,
            ]);
            break;
          case "error":
            setError(msg.data?.message ?? "Unknown error");
            break;
        }
      } catch {
        // ignore malformed messages
      }
    };

    ws.onclose = () => {
      setConnected(false);
      reconnectTimerRef.current = setTimeout(() => connect(), 3000);
    };

    ws.onerror = () => {
      setError("WebSocket connection error");
      ws.close();
    };
  }, [traceId]);

  const disconnect = useCallback(() => {
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
    wsRef.current?.close();
    wsRef.current = null;
    setConnected(false);
  }, []);

  const reconnect = useCallback(() => {
    disconnect();
    connect();
  }, [disconnect, connect]);

  useEffect(() => {
    connect();
    return () => {
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
      }
      wsRef.current?.close();
      setConnected(false);
    };
  }, [connect]);

  return { metrics, activities, connected, error, disconnect, reconnect };
}
