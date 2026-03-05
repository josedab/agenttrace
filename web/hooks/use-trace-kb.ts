"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useKBEntries(limit: number = 20, offset: number = 0) {
  return useQuery({
    queryKey: ["kb-entries", limit, offset],
    queryFn: () =>
      api.get(`/api/public/trace-kb/entries?limit=${limit}&offset=${offset}`),
  });
}

export function useCreateKBEntry() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { title: string; content: string; tags?: string[]; traceId?: string }) =>
      api.post("/api/public/trace-kb/entries", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["kb-entries"] });
    },
  });
}

export function useKBEntry(entryId: string | null) {
  return useQuery({
    queryKey: ["kb-entry", entryId],
    queryFn: () => api.get(`/api/public/trace-kb/entries/${entryId}`),
    enabled: !!entryId,
  });
}

export function useKBSearch(query: string) {
  return useQuery({
    queryKey: ["kb-search", query],
    queryFn: () => api.get(`/api/public/trace-kb/search?query=${encodeURIComponent(query)}`),
    enabled: !!query,
  });
}

export function useKBSuggestions(traceId: string | null) {
  return useQuery({
    queryKey: ["kb-suggestions", traceId],
    queryFn: () => api.get(`/api/public/trace-kb/suggestions?traceId=${traceId}`),
    enabled: !!traceId,
  });
}
