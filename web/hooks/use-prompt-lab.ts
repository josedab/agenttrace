"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface Experiment {
  id: string;
  name: string;
  status: "draft" | "running" | "completed";
  variants: { name: string; prompt: string; metrics?: Record<string, number> }[];
  createdAt: string;
}

export interface PromptSuggestion {
  original: string;
  suggested: string;
  reason: string;
  expectedImprovement: number;
}

export function usePromptExperiments() {
  return useQuery({
    queryKey: ["prompt-lab", "experiments"],
    queryFn: () => api.promptLab.listExperiments(),
  });
}

export function useCreateExperiment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; variants: { name: string; prompt: string }[] }) =>
      api.promptLab.createExperiment(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["prompt-lab", "experiments"] }),
  });
}

export function useStartExperiment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.promptLab.startExperiment(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["prompt-lab", "experiments"] }),
  });
}

export function useCompleteExperiment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.promptLab.completeExperiment(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["prompt-lab", "experiments"] }),
  });
}

export function usePromptSuggestions(promptName?: string) {
  return useQuery({
    queryKey: ["prompt-lab", "suggestions", promptName],
    queryFn: () => api.promptLab.getSuggestions(promptName),
    enabled: !!promptName,
  });
}
