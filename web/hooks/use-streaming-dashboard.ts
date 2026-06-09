'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';

export interface ActiveStream {
  id: string;
  model: string;
  elapsedSeconds: number;
  tokensPerSecond: number;
  costPerMinute: number;
  progress: number;
  status: 'streaming' | 'processing' | 'completing';
}

export interface StreamingModelUsage {
  model: string;
  sessions: number;
  totalTokens: number;
  totalCost: number;
}

export interface StreamingDashboardMetrics {
  activeSessions: number;
  totalCost: number;
  totalTokens: number;
  errorCount: number;
  activeStreams: ActiveStream[];
  topModels: StreamingModelUsage[];
}

interface LiveMetricsResponse {
  traceId: string;
  activeSpans: number;
  completedSpans: number;
  totalTokens: number;
  totalCost: number;
  errorCount: number;
  elapsedMs: number;
  tokensPerSecond: number;
  costPerMinute: number;
}

interface StreamingDashboardResponse {
  activeSessions: number;
  totalCost: number;
  totalTokens: number;
  errorCount: number;
  activeStreams: LiveMetricsResponse[];
  topModels: Array<{
    model: string;
    requestCount: number;
    totalTokens: number;
    totalCost: number;
  }>;
}

export function useStreamingDashboard() {
  return useQuery({
    queryKey: ['streaming-dashboard'],
    queryFn: async () => {
      const response = await api.get<StreamingDashboardResponse>('/api/public/streaming-dashboard');
      return {
        activeSessions: response.activeSessions,
        totalCost: response.totalCost,
        totalTokens: response.totalTokens,
        errorCount: response.errorCount,
        activeStreams: response.activeStreams.map((stream) => {
          const spanCount = stream.activeSpans + stream.completedSpans;
          return {
            id: stream.traceId,
            model: 'Trace stream',
            elapsedSeconds: Math.floor(stream.elapsedMs / 1000),
            tokensPerSecond: stream.tokensPerSecond,
            costPerMinute: stream.costPerMinute,
            progress: spanCount === 0 ? 0 : (stream.completedSpans / spanCount) * 100,
            status: stream.activeSpans > 0 ? 'streaming' : 'processing',
          } satisfies ActiveStream;
        }),
        topModels: response.topModels.map((model) => ({
          model: model.model,
          sessions: model.requestCount,
          totalTokens: model.totalTokens,
          totalCost: model.totalCost,
        })),
      } satisfies StreamingDashboardMetrics;
    },
    refetchInterval: 3000,
  });
}

export function useDashboardConfig() {
  return useQuery({
    queryKey: ['streaming-dashboard-config'],
    queryFn: () => api.get<Record<string, unknown>>('/api/public/streaming-dashboard/config'),
  });
}

export function useUpdateDashboardConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (config: Record<string, unknown>) =>
      api.put('/api/public/streaming-dashboard/config', config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['streaming-dashboard-config'] });
    },
  });
}
