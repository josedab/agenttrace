"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useFleetDashboard() {
  return useQuery({
    queryKey: ["fleet-dashboard"],
    queryFn: () => api.fleet.getDashboard(),
  });
}

export function useFleetAgents() {
  return useQuery({
    queryKey: ["fleet-agents"],
    queryFn: () => api.fleet.listAgents(),
  });
}

export function useFleetPolicies() {
  return useQuery({
    queryKey: ["fleet-policies"],
    queryFn: () => api.fleet.listPolicies(),
  });
}

export function useCreateFleetPolicy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; rules: Record<string, unknown>; scope?: string }) =>
      api.fleet.createPolicy(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["fleet-policies"] }),
  });
}

export function useBulkUpdate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { agentIds: string[]; updates: Record<string, unknown> }) =>
      api.fleet.bulkUpdate(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["fleet-agents"] });
      queryClient.invalidateQueries({ queryKey: ["fleet-dashboard"] });
    },
  });
}

export function useFleetScaling() {
  return useQuery({
    queryKey: ["fleet-scaling"],
    queryFn: () => api.fleet.getScaling(),
  });
}
