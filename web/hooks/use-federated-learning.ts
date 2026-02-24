"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface FederationRing {
  id: string;
  name: string;
  participants: number;
  status: string;
  privacyLevel: string;
  createdAt: string;
}

export interface FederatedInsight {
  ringId: string;
  metric: string;
  aggregatedValue: number;
  participantCount: number;
  confidence: number;
  generatedAt: string;
}

export interface FederationConfig {
  enabled: boolean;
  privacyBudget: number;
  noiseMultiplier: number;
  minParticipants: number;
}

export function useFederationRings() {
  return useQuery({
    queryKey: ["federation-rings"],
    queryFn: () => api.federatedLearning.listRings() as Promise<FederationRing[]>,
    refetchInterval: 30000,
  });
}

export function useJoinRing() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { ringId: string; contribution?: Record<string, unknown> }) =>
      api.federatedLearning.joinRing(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["federation-rings"] }),
  });
}

export function useFederatedInsights(ringId: string | null) {
  return useQuery({
    queryKey: ["federated-insights", ringId],
    queryFn: () => api.federatedLearning.getInsights(ringId!) as Promise<FederatedInsight[]>,
    enabled: !!ringId,
  });
}
