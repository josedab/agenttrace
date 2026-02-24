"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface SemanticSearchResult {
  traceId: string;
  traceName: string;
  score: number;
  highlights: string[];
  matchedSpans: { spanId: string; name: string; relevance: number }[];
  timestamp: string;
}

export interface TraceCluster {
  id: string;
  name: string;
  description: string;
  traceCount: number;
  centroid: number[];
  representativeTraceId: string;
  tags: string[];
}

export interface AnomalyPattern {
  id: string;
  name: string;
  description: string;
  frequency: number;
  severity: "low" | "medium" | "high" | "critical";
  affectedTraces: number;
  detectedAt: string;
  pattern: Record<string, unknown>;
}

export interface SearchDashboard {
  totalTraces: number;
  indexedTraces: number;
  clusterCount: number;
  anomalyPatternCount: number;
  recentSearches: { query: string; resultCount: number; timestamp: string }[];
}

export function useSemanticSearch() {
  return useMutation({
    mutationFn: (data: { query: string; limit?: number; filters?: Record<string, unknown> }) =>
      api.semanticSearch.search(data) as Promise<SemanticSearchResult[]>,
  });
}

export function useTraceClusters() {
  return useQuery({
    queryKey: ["trace-clusters"],
    queryFn: () =>
      api.semanticSearch.getClusters() as Promise<TraceCluster[]>,
  });
}

export function useAnomalyPatterns() {
  return useQuery({
    queryKey: ["anomaly-patterns"],
    queryFn: () =>
      api.semanticSearch.getAnomalyPatterns() as Promise<AnomalyPattern[]>,
  });
}

export function useSearchDashboard() {
  return useQuery({
    queryKey: ["search-dashboard"],
    queryFn: () =>
      api.semanticSearch.getDashboard() as Promise<SearchDashboard>,
  });
}
