"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface KGNode {
  id: string;
  type: string;
  label: string;
  properties: Record<string, unknown>;
}

export interface KGEdge {
  source: string;
  target: string;
  relationship: string;
  weight: number;
}

export interface KnowledgeGraphData {
  nodes: KGNode[];
  edges: KGEdge[];
}

export interface KGStats {
  totalNodes: number;
  totalEdges: number;
  byType: Record<string, number>;
  density: number;
}

export function useKnowledgeGraph() {
  return useQuery({
    queryKey: ["knowledge-graph"],
    queryFn: () => api.knowledgeGraph.build() as Promise<KnowledgeGraphData>,
    refetchInterval: 60000,
  });
}

export function useKGQuery() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { query: string; depth?: number; nodeTypes?: string[] }) =>
      api.knowledgeGraph.query(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["knowledge-graph"] }),
  });
}

export function useKGStats() {
  return useQuery({
    queryKey: ["kg-stats"],
    queryFn: () => api.knowledgeGraph.getStats() as Promise<KGStats>,
  });
}
