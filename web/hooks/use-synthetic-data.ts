"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface GenerateSyntheticDataInput {
  name: string;
  sourceDatasetId?: string;
  count: number;
  variationStrategy?: string;
  parameters?: Record<string, unknown>;
}

export function useSyntheticDatasets() {
  return useQuery({
    queryKey: ["synthetic-data", "datasets"],
    queryFn: () => api.syntheticData.list(),
  });
}

export function useGenerateSynthetic() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: GenerateSyntheticDataInput) => api.syntheticData.generate(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["synthetic-data"] }),
  });
}

export function useSyntheticStats() {
  return useQuery({
    queryKey: ["synthetic-data", "stats"],
    queryFn: () => api.syntheticData.getStats(),
  });
}
