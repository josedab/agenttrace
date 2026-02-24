"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface IntentDeclaration {
  id: string;
  agentId: string;
  description: string;
  expectedOutcome: string;
  status: string;
  verifiedAt?: string;
  alignmentScore?: number;
  createdAt: string;
}

export interface IntentStats {
  totalDeclared: number;
  totalVerified: number;
  averageAlignment: number;
  byStatus: Record<string, number>;
}

export function useDeclareIntent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { agentId: string; description: string; expectedOutcome: string }) =>
      api.intents.declare(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["intent-stats"] }),
  });
}

export function useVerifyIntent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: { actualOutcome: string; evidence?: Record<string, unknown> } }) =>
      api.intents.verify(id, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["intent-stats"] }),
  });
}

export function useIntentStats() {
  return useQuery({
    queryKey: ["intent-stats"],
    queryFn: () => api.intents.getStats() as Promise<IntentStats>,
    refetchInterval: 30000,
  });
}
