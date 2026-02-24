"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useCarbonFootprint() {
  return useQuery({
    queryKey: ["carbon", "footprint"],
    queryFn: () => api.carbon.getFootprint(),
  });
}

export function useCarbonConfig() {
  return useQuery({
    queryKey: ["carbon", "config"],
    queryFn: () => api.carbon.getConfig(),
  });
}

export function useCarbonSuggestions() {
  return useQuery({
    queryKey: ["carbon", "suggestions"],
    queryFn: () => api.carbon.getSuggestions(),
  });
}
