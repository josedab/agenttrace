"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface Annotation {
  id: string;
  traceId: string;
  spanId?: string;
  userId: string;
  userName: string;
  type: "comment" | "label" | "score" | "flag";
  content: string;
  label?: string;
  score?: number;
  resolved: boolean;
  parentId?: string;
  replies: Annotation[];
  createdAt: string;
  updatedAt: string;
}

export interface AnnotationQueue {
  id: string;
  name: string;
  description: string;
  traceCount: number;
  annotatedCount: number;
  reviewerIds: string[];
  labelSchema: LabelSchema[];
  scoringRubric?: ScoringRubric;
  status: "active" | "completed" | "paused";
  createdAt: string;
}

export interface LabelSchema {
  name: string;
  description: string;
  options: string[];
  required: boolean;
  multiSelect: boolean;
}

export interface ScoringRubric {
  name: string;
  minScore: number;
  maxScore: number;
  criteria: { name: string; description: string; weight: number }[];
}

export interface AnnotationAssignment {
  id: string;
  queueId: string;
  traceId: string;
  reviewerId: string;
  reviewerName: string;
  status: "pending" | "in_progress" | "completed" | "skipped";
  annotations: Annotation[];
  assignedAt: string;
  completedAt?: string;
}

export interface InterAnnotatorAgreement {
  queueId: string;
  metricName: string;
  overallScore: number;
  pairwiseScores: { reviewer1: string; reviewer2: string; score: number }[];
  labelAgreement: Record<string, number>;
  conflictCount: number;
  totalAnnotations: number;
}

export interface AnnotationPresence {
  userId: string;
  userName: string;
  cursor?: { spanId: string; position: number };
  lastActive: string;
}

export function useAnnotations(traceId: string) {
  return useQuery({
    queryKey: ["annotations", traceId],
    queryFn: () => api.annotations.list(traceId) as Promise<Annotation[]>,
    enabled: !!traceId,
  });
}

export function useCreateAnnotation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      traceId: string;
      spanId?: string;
      type: "comment" | "label" | "score" | "flag";
      content: string;
      label?: string;
      score?: number;
      parentId?: string;
    }) => api.annotations.create(data),
    onSuccess: (_data, variables) =>
      queryClient.invalidateQueries({ queryKey: ["annotations", variables.traceId] }),
  });
}

export function useReplyToAnnotation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: { content: string } }) =>
      api.annotations.reply(id, data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["annotations"] }),
  });
}

export function useResolveAnnotation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.annotations.resolve(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["annotations"] }),
  });
}

export function useAnnotationPresence(traceId: string) {
  return useQuery({
    queryKey: ["annotations", "presence", traceId],
    queryFn: () => api.annotations.getPresence(traceId) as Promise<AnnotationPresence[]>,
    enabled: !!traceId,
    refetchInterval: 5000,
  });
}

export function useAnnotationQueues() {
  return useQuery({
    queryKey: ["annotation-queues"],
    queryFn: () => api.collabHub.listQueues() as Promise<AnnotationQueue[]>,
  });
}

export function useCreateAnnotationQueue() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      name: string;
      description: string;
      traceIds: string[];
      reviewerIds: string[];
      labelSchema: LabelSchema[];
      scoringRubric?: ScoringRubric;
    }) => api.collabHub.createQueue(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["annotation-queues"] }),
  });
}

export function useAssignReview() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { queueId: string; traceId: string; reviewerId: string }) =>
      api.collabHub.assignReview(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["annotation-queues"] }),
  });
}

export function useCompleteReview() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: { annotations: Partial<Annotation>[] } }) =>
      api.collabHub.completeReview(id, data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["annotation-queues"] }),
  });
}

export function useInterAnnotatorAgreement(queueId: string | null) {
  return useQuery({
    queryKey: ["iaa-metrics", queueId],
    queryFn: () =>
      api.collabHub.listStandards() as Promise<InterAnnotatorAgreement>,
    enabled: !!queueId,
  });
}
