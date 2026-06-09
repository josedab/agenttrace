'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';

export interface FileImpact {
  path: string;
  operation: 'created' | 'modified' | 'deleted';
  linesAdded: number;
  linesRemoved: number;
  language: string;
  complexity: 'low' | 'medium' | 'high';
}

export interface ImpactSummary {
  totalFiles: number;
  totalLinesAdded: number;
  totalLinesRemoved: number;
  languages: string[];
}

export interface CodeImpact {
  files: FileImpact[];
  summary: ImpactSummary;
}

export function useCodeImpact(traceId: string | null) {
  return useQuery({
    queryKey: ['code-impact', traceId],
    queryFn: () => api.get<CodeImpact>(`/api/public/traces/${traceId}/code-impact`),
    enabled: !!traceId,
  });
}

export function useCodeImpactSummary() {
  return useQuery({
    queryKey: ['code-impact-summary'],
    queryFn: () => api.get<ImpactSummary>('/api/public/code-impact/summary'),
  });
}

export function useCodeImpactFileTree(traceId: string | null) {
  return useQuery({
    queryKey: ['code-impact-file-tree', traceId],
    queryFn: () => api.get<FileImpact[]>(`/api/public/code-impact/file-tree?traceId=${traceId}`),
    enabled: !!traceId,
  });
}
