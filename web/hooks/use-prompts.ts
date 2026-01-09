"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface PromptFilters {
  search?: string;
  label?: string;
  tags?: string[];
}

export function usePrompts(filters: PromptFilters = {}) {
  return useQuery({
    queryKey: ["prompts", filters],
    queryFn: () => api.prompts.list(filters),
  });
}

export function usePrompt(promptName: string) {
  return useQuery({
    queryKey: ["prompt", promptName],
    queryFn: () => api.prompts.get(promptName),
    enabled: !!promptName,
  });
}

export function usePromptVersions(promptName: string) {
  return useQuery({
    queryKey: ["prompt-versions", promptName],
    queryFn: () => api.prompts.listVersions(promptName),
    enabled: !!promptName,
  });
}

export function usePromptVersion(promptName: string, version: number) {
  return useQuery({
    queryKey: ["prompt-version", promptName, version],
    queryFn: () => api.prompts.getVersion(promptName, version),
    enabled: !!promptName && version !== undefined,
  });
}

export function useCreatePrompt() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      name: string;
      prompt: string;
      config?: Record<string, any>;
      labels?: string[];
    }) => api.prompts.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["prompts"] });
    },
  });
}

export function useUpdatePrompt() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      promptName,
      data,
    }: {
      promptName: string;
      data: {
        prompt: string;
        config?: Record<string, any>;
      };
    }) => api.prompts.createVersion(promptName, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["prompt", variables.promptName] });
      queryClient.invalidateQueries({ queryKey: ["prompt-versions", variables.promptName] });
    },
  });
}

export function useSetPromptLabel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      promptName,
      version,
      label,
    }: {
      promptName: string;
      version: number;
      label: string;
    }) => api.prompts.setLabel(promptName, version, label),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["prompt", variables.promptName] });
      queryClient.invalidateQueries({ queryKey: ["prompt-versions", variables.promptName] });
    },
  });
}

export function useRemovePromptLabel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      promptName,
      label,
    }: {
      promptName: string;
      label: string;
    }) => api.prompts.removeLabel(promptName, label),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["prompt", variables.promptName] });
      queryClient.invalidateQueries({ queryKey: ["prompt-versions", variables.promptName] });
    },
  });
}

export function useDeletePrompt() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (promptName: string) => api.prompts.delete(promptName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["prompts"] });
    },
  });
}

export function useCompilePrompt() {
  return useMutation({
    mutationFn: ({
      promptName,
      version,
      variables,
    }: {
      promptName: string;
      version?: number;
      label?: string;
      variables: Record<string, any>;
    }) => api.prompts.compile(promptName, version, variables),
  });
}
