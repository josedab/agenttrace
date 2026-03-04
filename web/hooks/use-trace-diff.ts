"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface TraceDiffInput {
  leftTraceId: string;
  rightTraceId: string;
}

export interface BisectStartInput {
  goodTraceId: string;
  badTraceId: string;
  traceHistory: string[];
  metricName: string;
  threshold?: number;
}

export function useTraceDiff(input: TraceDiffInput | null) {
  return useQuery({
    queryKey: ["trace-diff", input?.leftTraceId, input?.rightTraceId],
    queryFn: () => api.post("/api/public/trace-diff", input),
    enabled: !!input?.leftTraceId && !!input?.rightTraceId,
  });
}

export function useBisectSessions() {
  return useQuery({
    queryKey: ["bisect-sessions"],
    queryFn: () => api.get<{ sessions: any[] }>("/api/public/bisect/sessions"),
  });
}

export function useBisectSession(sessionId: string | null) {
  return useQuery({
    queryKey: ["bisect-session", sessionId],
    queryFn: () => api.get(`/api/public/bisect/sessions/${sessionId}`),
    enabled: !!sessionId,
    refetchInterval: 5000,
  });
}

export function useStartBisect() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: BisectStartInput) =>
      api.post("/api/public/bisect/sessions", input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bisect-sessions"] });
    },
  });
}

export function useSubmitBisectVerdict(sessionId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (verdict: "good" | "bad" | "skip") =>
      api.post(`/api/public/bisect/sessions/${sessionId}/verdict`, { verdict }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bisect-session", sessionId] });
    },
  });
}

export function useBisectResult(sessionId: string | null) {
  return useQuery({
    queryKey: ["bisect-result", sessionId],
    queryFn: () => api.get(`/api/public/bisect/sessions/${sessionId}/result`),
    enabled: !!sessionId,
  });
}
