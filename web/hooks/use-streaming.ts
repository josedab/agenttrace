'use client';

import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
} from 'react';
import { API_URL, getApiAccessToken, getApiProjectId } from '@/lib/api';

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
  const abortRef = useRef<AbortController | null>(null);
  const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const connect = useCallback(() => {
    if (!traceId) return;

    if (retryRef.current) {
      clearTimeout(retryRef.current);
      retryRef.current = null;
    }

    const controller = new AbortController();
    abortRef.current?.abort();
    abortRef.current = controller;

    void (async () => {
      const headers = new Headers({ Accept: 'text/event-stream' });
      const token = getApiAccessToken();
      if (token) {
        headers.set('Authorization', ['Bearer', token].join(' '));
      }
      const projectId = getApiProjectId();
      if (projectId) {
        headers.set('X-Project-ID', projectId);
      }

      const response = await fetch(`${API_URL}/api/public/traces/${traceId}/stream?follow=true`, {
        headers,
        signal: controller.signal,
      });
      if (!response.ok || !response.body) {
        throw new Error(`stream request failed with status ${response.status}`);
      }

      setConnected(true);
      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      while (!controller.signal.aborted) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        buffer = buffer.replace(/\r\n/g, '\n');
        let boundary = buffer.indexOf('\n\n');
        while (boundary >= 0) {
          const block = buffer.slice(0, boundary);
          buffer = buffer.slice(boundary + 2);
          handleSSEBlock(block, setMetrics, setActivities);
          boundary = buffer.indexOf('\n\n');
        }
      }

      if (!controller.signal.aborted) {
        setConnected(false);
        retryRef.current = setTimeout(connect, 3000);
      }
    })().catch(() => {
      if (controller.signal.aborted) return;
      setConnected(false);
      retryRef.current = setTimeout(connect, 3000);
    });
  }, [traceId]);

  useEffect(() => {
    connect();
    return () => {
      abortRef.current?.abort();
      if (retryRef.current) clearTimeout(retryRef.current);
      setConnected(false);
    };
  }, [connect]);

  const disconnect = useCallback(() => {
    abortRef.current?.abort();
    abortRef.current = null;
    if (retryRef.current) {
      clearTimeout(retryRef.current);
      retryRef.current = null;
    }
    setConnected(false);
  }, []);

  return { metrics, activities, connected, disconnect };
}

function handleSSEBlock(
  block: string,
  setMetrics: (metrics: LiveMetrics) => void,
  setActivities: Dispatch<SetStateAction<StreamActivity[]>>
) {
  let eventType = 'message';
  const dataLines: string[] = [];

  for (const line of block.split('\n')) {
    if (line.startsWith('event:')) {
      eventType = line.slice('event:'.length).trim();
    } else if (line.startsWith('data:')) {
      dataLines.push(line.slice('data:'.length).trimStart());
    }
  }

  if (dataLines.length === 0) return;

  try {
    const data: unknown = JSON.parse(dataLines.join('\n'));
    if (eventType === 'metrics') {
      setMetrics(data as LiveMetrics);
    } else if (eventType === 'activity') {
      setActivities((previous) => [...previous.slice(-499), data as StreamActivity]);
    }
  } catch {
    // Ignore malformed stream events.
  }
}
