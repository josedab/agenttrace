"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface KnowledgeGraphView {
  nodes: KGNode[];
  edges: KGEdge[];
  stats: KGStats;
}

export interface KGNode {
  id: string;
  type: "agent" | "tool" | "api" | "file" | "concept" | "pattern";
  label: string;
  properties: Record<string, unknown>;
  weight: number;
}

export interface KGEdge {
  id: string;
  source: string;
  target: string;
  type: string;
  label: string;
  weight: number;
}

export interface KGStats {
  totalNodes: number;
  totalEdges: number;
  nodesByType: Record<string, number>;
  density: number;
  lastUpdated: string;
}

export interface KGEvolution {
  snapshots: { timestamp: string; nodeCount: number; edgeCount: number }[];
  newNodes: KGNode[];
  removedNodes: string[];
  changedEdges: KGEdge[];
}

export function useKnowledgeGraph(focusNode?: string) {
  return useQuery({
    queryKey: ["knowledge-graph", focusNode],
    queryFn: () =>
      api.agentKnowledgeGraph.query({ focusNode }) as Promise<KnowledgeGraphView>,
  });
}

export function useKGEvolution() {
  return useQuery({
    queryKey: ["kg-evolution"],
    queryFn: () =>
      api.agentKnowledgeGraph.getEvolution() as Promise<KGEvolution>,
  });
}

export function useKGStats() {
  return useQuery({
    queryKey: ["kg-stats"],
    queryFn: () =>
      api.agentKnowledgeGraph.getStats() as Promise<KGStats>,
  });
}
