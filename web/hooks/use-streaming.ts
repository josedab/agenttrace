"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { API_URL } from "@/lib/api";

export interface LiveMetrics {
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

export interface StreamActivity {
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

export function useTraceStream(traceId: string | null) {
  const [metrics, setMetrics] = useState<LiveMetrics | null>(null);
  const [activities, setActivities] = useState<StreamActivity[]>([]);
  const [connected, setConnected] = useState(false);
  const eventSourceRef = useRef<EventSource | null>(null);

  const connect = useCallback(() => {
    if (!traceId) return;

    const url = `${API_URL}/api/public/traces/${traceId}/stream?follow=true`;
    const es = new EventSource(url);
    eventSourceRef.current = es;

    es.onopen = () => setConnected(true);

    es.addEventListener("metrics", (event) => {
      try {
        const data = JSON.parse(event.data);
        setMetrics(data);
      } catch {}
    });

    es.addEventListener("activity", (event) => {
      try {
        const data = JSON.parse(event.data) as StreamActivity;
        setActivities((prev) => [...prev.slice(-499), data]);
      } catch {}
    });

    es.onerror = () => {
      setConnected(false);
      es.close();
      // Reconnect after 3 seconds
      setTimeout(() => connect(), 3000);
    };
  }, [traceId]);

  useEffect(() => {
    connect();
    return () => {
      eventSourceRef.current?.close();
      setConnected(false);
    };
  }, [connect]);

  const disconnect = useCallback(() => {
    eventSourceRef.current?.close();
    setConnected(false);
  }, []);

  return { metrics, activities, connected, disconnect };
}
