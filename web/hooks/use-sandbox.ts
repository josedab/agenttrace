"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface SandboxReview {
  id: string;
  status: "pending" | "approved" | "rejected";
  riskLevel: string;
  riskScore: number;
  proposedActions: {
    id: string;
    type: string;
    target: string;
    description: string;
    riskLevel: string;
  }[];
  createdAt: string;
}

export interface SandboxPolicy {
  id: string;
  name: string;
  rules: { action: string; condition: string; effect: "allow" | "deny" }[];
}

export function usePendingReviews() {
  return useQuery({
    queryKey: ["sandbox", "pending"],
    queryFn: () => api.sandbox.listPending(),
  });
}

export function useSubmitReview() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { traceId: string; actions: SandboxReview["proposedActions"] }) =>
      api.sandbox.submitReview(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["sandbox"] }),
  });
}

export function useSandboxReview(id: string) {
  return useQuery({
    queryKey: ["sandbox", "review", id],
    queryFn: () => api.sandbox.getReview(id),
    enabled: !!id,
  });
}

export function useDecideReview() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, decision }: { id: string; decision: { action: "approve" | "reject"; reason?: string } }) =>
      api.sandbox.decide(id, decision),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["sandbox"] }),
  });
}

export function useSandboxPolicies() {
  return useQuery({
    queryKey: ["sandbox", "policies"],
    queryFn: () => api.sandbox.listPolicies(),
  });
}

export function useCreatePolicy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; rules: SandboxPolicy["rules"] }) =>
      api.sandbox.createPolicy(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["sandbox", "policies"] }),
  });
}

export function useSandboxStats() {
  return useQuery({
    queryKey: ["sandbox", "stats"],
    queryFn: () => api.sandbox.getStats(),
  });
}
