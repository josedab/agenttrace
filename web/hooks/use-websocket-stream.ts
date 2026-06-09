'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { API_URL, getApiAccessToken, getApiProjectId } from '@/lib/api';

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
  data: unknown;
  timestamp: string;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

export function useWebSocketStream(traceId: string | null) {
  const [metrics, setMetrics] = useState<WSLiveMetrics | null>(null);
  const [activities, setActivities] = useState<WSStreamActivity[]>([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const shouldReconnectRef = useRef(false);

  const connect = useCallback(() => {
    if (!traceId) return;

    shouldReconnectRef.current = true;
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }

    const apiUrl = new URL(
      API_URL || (typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8080')
    );
    const wsUrl =
      process.env.NEXT_PUBLIC_WS_URL ??
      `${apiUrl.protocol === 'https:' ? 'wss:' : 'ws:'}//${apiUrl.host}`;

    const token = getApiAccessToken();
    const projectId = getApiProjectId();
    const protocols = ['agenttrace'];
    if (token) protocols.push(token);
    if (projectId) protocols.push(`project.${projectId}`);
    const ws = new WebSocket(`${wsUrl}/ws/streaming/${traceId}`, protocols);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      setError(null);
      ws.send(JSON.stringify({ action: 'subscribe', traceId }));
    };

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data);
        switch (msg.type) {
          case 'metrics':
            setMetrics(msg.data as WSLiveMetrics);
            break;
          case 'activity':
            setActivities((prev) => [...prev.slice(-499), msg.data as WSStreamActivity]);
            break;
          case 'error':
            setError(
              isRecord(msg.data) && typeof msg.data.message === 'string'
                ? msg.data.message
                : 'Unknown error'
            );
            break;
        }
      } catch {
        // ignore malformed messages
      }
    };

    ws.onclose = () => {
      if (wsRef.current !== ws) return;

      wsRef.current = null;
      setConnected(false);
      if (shouldReconnectRef.current) {
        reconnectTimerRef.current = setTimeout(connect, 3000);
      }
    };

    ws.onerror = () => {
      setError('WebSocket connection error');
      ws.close();
    };
  }, [traceId]);

  const disconnect = useCallback(() => {
    shouldReconnectRef.current = false;
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
    const ws = wsRef.current;
    wsRef.current = null;
    ws?.close();
    setConnected(false);
  }, []);

  const reconnect = useCallback(() => {
    disconnect();
    connect();
  }, [disconnect, connect]);

  useEffect(() => {
    connect();
    return () => {
      shouldReconnectRef.current = false;
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
      const ws = wsRef.current;
      wsRef.current = null;
      ws?.close();
      setConnected(false);
    };
  }, [connect]);

  return { metrics, activities, connected, error, disconnect, reconnect };
}
