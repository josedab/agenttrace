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
  indexHealth: "healthy" | "stale" | "rebuilding";
  embeddingModel: string;
  lastIndexedAt: string;
}

export interface RAGSearchQuery {
  query: string;
  limit?: number;
  filters?: Record<string, unknown>;
  searchMode?: "semantic" | "hybrid" | "keyword";
  minScore?: number;
  dateRange?: { from: string; to: string };
  traceTypes?: string[];
  includeContext?: boolean;
}

export interface RAGSearchResult extends SemanticSearchResult {
  context?: string;
  summary?: string;
  relatedTraces?: { traceId: string; similarity: number }[];
}

export interface EmbeddingIndexStatus {
  totalDocuments: number;
  indexedDocuments: number;
  pendingDocuments: number;
  lastIndexedAt: string;
  embeddingModel: string;
  vectorDimensions: number;
  indexSizeBytes: number;
  status: "ready" | "indexing" | "error";
}

export function useSemanticSearch() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: RAGSearchQuery) =>
      api.semanticSearch.search(data) as Promise<RAGSearchResult[]>,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["search-dashboard"] });
    },
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

export function useRebuildIndex() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () =>
      api.semanticSearch.search({ query: "__rebuild_index__", limit: 0 }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["search-dashboard"] });
    },
  });
}

export function useSimilarTraces(traceId: string) {
  return useQuery({
    queryKey: ["similar-traces", traceId],
    queryFn: () =>
      api.semanticSearch.search({ query: `similar:${traceId}`, limit: 10 }) as Promise<RAGSearchResult[]>,
    enabled: !!traceId,
  });
}
