"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface DebugQuery {
  traceId: string;
  question: string;
  context?: Record<string, unknown>;
}

export interface DebugResponse {
  id: string;
  traceId: string;
  question: string;
  answer: string;
  rootCauses: RootCause[];
  suggestedFixes: SuggestedFix[];
  confidence: number;
  createdAt: string;
}

export interface DebugContext {
  traceId: string;
  spans: { id: string; name: string; status: string; durationMs: number }[];
  errors: { message: string; spanId: string; timestamp: string }[];
  metadata: Record<string, unknown>;
}

export interface RootCause {
  id: string;
  category: string;
  description: string;
  confidence: number;
  evidence: string[];
  spanId?: string;
}

export interface SuggestedFix {
  id: string;
  rootCauseId: string;
  description: string;
  codeSnippet?: string;
  impact: "low" | "medium" | "high";
  effort: "low" | "medium" | "high";
}

export function useDebugTrace() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: DebugQuery) =>
      api.aiDebugger.debug(data) as Promise<DebugResponse>,
    onSuccess: (_, variables) =>
      queryClient.invalidateQueries({ queryKey: ["debug-history", variables.traceId] }),
  });
}

export function useDebugHistory(traceId: string | null) {
  return useQuery({
    queryKey: ["debug-history", traceId],
    queryFn: () =>
      api.aiDebugger.getHistory(traceId!) as Promise<DebugResponse[]>,
    enabled: !!traceId,
  });
}

export function useDebugContext(traceId: string | null) {
  return useQuery({
    queryKey: ["debug-context", traceId],
    queryFn: () =>
      api.aiDebugger.getContext(traceId!) as Promise<DebugContext>,
    enabled: !!traceId,
  });
}
