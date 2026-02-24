"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface TraceAttachment {
  id: string;
  traceId: string;
  type: string;
  mimeType: string;
  url: string;
  size: number;
  metadata: Record<string, unknown>;
  createdAt: string;
}

export interface AttachmentSummary {
  totalAttachments: number;
  byType: Record<string, number>;
  totalSize: number;
}

export function useTraceAttachments(traceId: string | null) {
  return useQuery({
    queryKey: ["trace-attachments", traceId],
    queryFn: () => api.multimodal.getTraceAttachments(traceId!) as Promise<TraceAttachment[]>,
    enabled: !!traceId,
  });
}

export function useRegisterAttachment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { traceId: string; type: string; mimeType: string; url: string; size?: number }) =>
      api.multimodal.register(data),
    onSuccess: (_data, variables) =>
      queryClient.invalidateQueries({ queryKey: ["trace-attachments", variables.traceId] }),
  });
}

export function useAttachmentSummary(traceId: string | null) {
  return useQuery({
    queryKey: ["attachment-summary", traceId],
    queryFn: () => api.multimodal.getSummary(traceId!) as Promise<AttachmentSummary>,
    enabled: !!traceId,
  });
}
