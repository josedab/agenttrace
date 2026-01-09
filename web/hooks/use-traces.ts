"use client";

import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface TraceFilters {
  search?: string;
  level?: string;
  minLatency?: number;
  maxLatency?: number;
  minCost?: number;
  maxCost?: number;
  startDate?: Date;
  endDate?: Date;
  tags?: string[];
  userId?: string;
  sessionId?: string;
}

export function useTraces(filters: TraceFilters = {}) {
  return useInfiniteQuery({
    queryKey: ["traces", filters],
    queryFn: async ({ pageParam }) => {
      const response = await api.traces.list({
        cursor: pageParam,
        limit: 50,
        ...filters,
      });
      return response;
    },
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    refetchInterval: 30000, // Refetch every 30 seconds
  });
}

export function useTraceCount(filters: TraceFilters = {}) {
  return useQuery({
    queryKey: ["trace-count", filters],
    queryFn: () => api.traces.count(filters),
    refetchInterval: 60000, // Refetch every minute
  });
}

export function useTraceSessions(limit = 10) {
  return useQuery({
    queryKey: ["trace-sessions", limit],
    queryFn: () => api.traces.sessions({ limit }),
  });
}

export function useTraceStats(dateRange: string = "7d") {
  return useQuery({
    queryKey: ["trace-stats", dateRange],
    queryFn: () => api.traces.stats({ dateRange }),
  });
}
