"use client";

import { useQuery, useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface CopilotResponse {
  answer: string;
  sources: { type: string; id: string; relevance: number }[];
  suggestions: string[];
}

export interface CopilotSuggestion {
  id: string;
  type: string;
  title: string;
  description: string;
  impact: string;
  priority: number;
  createdAt: string;
}

export interface CopilotInsight {
  id: string;
  category: string;
  title: string;
  description: string;
  affectedTraces: number;
  severity: string;
  createdAt: string;
}

export function useAskCopilot() {
  return useMutation({
    mutationFn: (data: { question: string; context?: Record<string, unknown> }) =>
      api.copilot.ask(data) as Promise<CopilotResponse>,
  });
}

export function useCopilotSuggestions() {
  return useQuery({
    queryKey: ["copilot-suggestions"],
    queryFn: () => api.copilot.getSuggestions() as Promise<CopilotSuggestion[]>,
    refetchInterval: 60000,
  });
}

export function useCopilotInsights() {
  return useQuery({
    queryKey: ["copilot-insights"],
    queryFn: () => api.copilot.getInsights() as Promise<CopilotInsight[]>,
    refetchInterval: 60000,
  });
}
