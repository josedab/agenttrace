"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useDistributedTrace(traceId: string) {
  return useQuery({
    queryKey: ["distributed-traces", traceId],
    queryFn: () => api.distributedTraces.get(traceId),
    enabled: !!traceId,
  });
}

export function useServiceMap() {
  return useQuery({
    queryKey: ["distributed-traces", "service-map"],
    queryFn: () => api.distributedTraces.getServiceMap(),
  });
}
