"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface FileTraceMapping {
  filePath: string;
  traceCount: number;
  annotations: LineAnnotation[];
  lastUpdated: string;
}

export interface LineAnnotation {
  line: number;
  traceId: string;
  spanId: string;
  type: "call" | "error" | "performance" | "coverage";
  message: string;
  severity?: "info" | "warning" | "error";
  metadata?: Record<string, unknown>;
}

export interface IDETraceContext {
  traceId: string;
  fileMappings: FileTraceMapping[];
  totalAnnotations: number;
  coveragePercent: number;
}

export function useFileMapping(filePath: string | null) {
  return useQuery({
    queryKey: ["file-mapping", filePath],
    queryFn: () =>
      api.ideTraceView.getFileMapping(filePath!) as Promise<FileTraceMapping>,
    enabled: !!filePath,
  });
}

export function useBatchMappings() {
  return useMutation({
    mutationFn: (data: { filePaths: string[] }) =>
      api.ideTraceView.batchMappings(data) as Promise<FileTraceMapping[]>,
  });
}

export function useIDETraceContext(traceId: string | null) {
  return useQuery({
    queryKey: ["ide-trace-context", traceId],
    queryFn: () =>
      api.ideTraceView.getTraceContext(traceId!) as Promise<IDETraceContext>,
    enabled: !!traceId,
  });
}
