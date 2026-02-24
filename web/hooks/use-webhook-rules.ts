"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface WebhookRule {
  id: string;
  name: string;
  event: string;
  conditions: Record<string, any>;
  action: { type: string; url: string; headers?: Record<string, string> };
  enabled: boolean;
  createdAt: string;
}

export interface WebhookTemplate {
  id: string;
  name: string;
  description: string;
  event: string;
  conditions: Record<string, any>;
  action: Record<string, any>;
}

export function useWebhookRules() {
  return useQuery({
    queryKey: ["webhook-rules"],
    queryFn: () => api.webhookRules.list(),
  });
}

export function useCreateWebhookRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; event: string; conditions: any; action: any }) =>
      api.webhookRules.create(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["webhook-rules"] }),
  });
}

export function useDeleteWebhookRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.webhookRules.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["webhook-rules"] }),
  });
}

export function useWebhookTemplates() {
  return useQuery({
    queryKey: ["webhook-rules", "templates"],
    queryFn: () => api.webhookRules.getTemplates(),
  });
}

export function useTestWebhookRule() {
  return useMutation({
    mutationFn: (id: string) => api.webhookRules.test(id),
  });
}
