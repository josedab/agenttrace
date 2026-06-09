"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface PlaygroundExecuteInput {
  code: string;
  language: "javascript" | "python";
  traceIds: string[];
  timeout?: number;
}

export interface PlaygroundTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  code: string;
  language: "javascript" | "python";
}

export interface PlaygroundResult {
  traceId: string;
  score: number | null;
  label: string | null;
  reasoning: string | null;
  error: string | null;
  durationMs: number;
}

export function usePlaygroundTemplates() {
  return useQuery({
    queryKey: ["eval-playground-templates"],
    queryFn: () => api.get<{ templates: PlaygroundTemplate[] }>("/api/public/eval-playground/templates"),
  });
}

export function usePlaygroundSession(sessionId: string | null) {
  return useQuery({
    queryKey: ["eval-playground-session", sessionId],
    queryFn: () => api.get(`/api/public/eval-playground/sessions/${sessionId}`),
    enabled: !!sessionId,
  });
}

export function useCreatePlaygroundSession() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: { name?: string; code?: string; language?: string }) =>
      api.post("/api/public/eval-playground/sessions", input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["eval-playground-session"] });
    },
  });
}

export function useExecutePlayground() {
  return useMutation({
    mutationFn: (input: PlaygroundExecuteInput) =>
      api.post<{ results: PlaygroundResult[] }>("/api/public/eval-playground/execute", input),
  });
}

export function useSharePlayground() {
  return useMutation({
    mutationFn: (sessionId: string) =>
      api.post<{ shareUrl: string; shareToken: string }>("/api/public/eval-playground/share", {
        sessionId,
      }),
  });
}

export function useSharedPlayground(shareToken: string | null) {
  return useQuery({
    queryKey: ["eval-playground-shared", shareToken],
    queryFn: () => api.get(`/api/public/eval-playground/shared/${shareToken}`),
    enabled: !!shareToken,
  });
}
