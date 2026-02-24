"use client";

import { useQuery, useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface SearchResult {
  traceId: string;
  score: number;
  highlight: string;
  metadata: Record<string, any>;
}

export interface SearchResponse {
  results: SearchResult[];
  totalCount: number;
  queryTime: number;
}

export function useSemanticSearch(query: string) {
  return useMutation({
    mutationFn: (data: { query: string; filters?: Record<string, any>; limit?: number }) =>
      api.search.query(data),
  });
}

export function useSearchSuggestions(prefix: string) {
  return useQuery({
    queryKey: ["search", "suggestions", prefix],
    queryFn: () => api.search.suggestions(prefix),
    enabled: prefix.length >= 2,
  });
}
