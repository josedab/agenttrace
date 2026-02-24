"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useCacheAnalysis() {
  return useQuery({
    queryKey: ["prompt-cache", "analysis"],
    queryFn: () => api.promptCache.analyze(),
  });
}

export function useCacheConfig() {
  return useQuery({
    queryKey: ["prompt-cache", "config"],
    queryFn: () => api.promptCache.getConfig(),
  });
}

export function useCacheStats() {
  return useQuery({
    queryKey: ["prompt-cache", "stats"],
    queryFn: () => api.promptCache.getStats(),
  });
}
