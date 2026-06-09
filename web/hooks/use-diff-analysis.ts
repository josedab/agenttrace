"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface DiffAnalysisSummary {
  id: string;
  traceId: string;
  status: string;
  overallScore: number;
  findingCount: number;
  filesChanged: number;
  createdAt: string;
}

export interface DiffAnalysis {
  id: string;
  projectId: string;
  traceId: string;
  status: string;
  filesAdded: number;
  filesModified: number;
  filesDeleted: number;
  linesAdded: number;
  linesRemoved: number;
  overallScore: number;
  dimensionScores: Record<string, number>;
  findings: DiffFinding[];
  fileAnalyses: FileAnalysis[];
  createdAt: string;
}

export interface DiffFinding {
  id: string;
  severity: string;
  category: string;
  title: string;
  description: string;
  filePath: string;
  startLine?: number;
  endLine?: number;
  suggestion?: string;
  confidence: number;
}

export interface FileAnalysis {
  filePath: string;
  language: string;
  linesAdded: number;
  linesRemoved: number;
  complexityDelta: number;
  qualityScore: number;
  findings: DiffFinding[];
}

export interface QualityTrend {
  points: { timestamp: string; overallScore: number; traceId: string }[];
  average: number;
  trend: string;
}

export interface FileChangeInput {
  filePath: string;
  changeType: "added" | "modified" | "deleted";
  diff?: string;
  beforeContent?: string;
  afterContent?: string;
}

export function useDiffAnalyses() {
  return useQuery({
    queryKey: ["diff-analyses"],
    queryFn: () =>
      api.diffAnalysis.list() as Promise<{
        analyses: DiffAnalysisSummary[];
        totalCount: number;
      }>,
    refetchInterval: 30000,
  });
}

export function useDiffAnalysis(id: string | null) {
  return useQuery({
    queryKey: ["diff-analysis", id],
    queryFn: () =>
      api.diffAnalysis.get(id!) as Promise<DiffAnalysis>,
    enabled: !!id,
  });
}

export function useQualityTrend(days = 30) {
  return useQuery({
    queryKey: ["quality-trend", days],
    queryFn: () =>
      api.diffAnalysis.trend(days) as Promise<QualityTrend>,
  });
}

export function useAnalyzeDiff() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { traceId: string; fileChanges: FileChangeInput[] }) =>
      api.diffAnalysis.analyze(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["diff-analyses"] });
    },
  });
}
