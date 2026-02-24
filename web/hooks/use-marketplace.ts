"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface MarketplacePackage {
  id: string;
  name: string;
  description: string;
  author: string;
  version: string;
  downloads: number;
  rating: number;
  tags: string[];
}

export function useMarketplaceSearch(query?: string) {
  return useQuery({
    queryKey: ["marketplace", "search", query],
    queryFn: () => api.marketplace.search(query ? { query } : undefined),
  });
}

export function useFeaturedPackages() {
  return useQuery({
    queryKey: ["marketplace", "featured"],
    queryFn: () => api.marketplace.featured(),
  });
}

export function useMarketplacePackage(id: string) {
  return useQuery({
    queryKey: ["marketplace", "package", id],
    queryFn: () => api.marketplace.get(id),
    enabled: !!id,
  });
}

export function usePublishPackage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; description: string; config: any; tags?: string[] }) =>
      api.marketplace.publish(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["marketplace"] }),
  });
}

export function useInstallPackage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.marketplace.install(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["marketplace"] }),
  });
}

export function useRatePackage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, score }: { id: string; score: number }) =>
      api.marketplace.rate(id, { score }),
    onSuccess: (_, { id }) => queryClient.invalidateQueries({ queryKey: ["marketplace", "package", id] }),
  });
}
