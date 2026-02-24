"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface ReviewQueue {
  id: string;
  name: string;
  description: string;
  assignmentCount: number;
  pendingCount: number;
  createdAt: string;
}

export interface ReviewAssignment {
  id: string;
  queueId: string;
  traceId: string;
  assigneeId: string;
  assigneeName: string;
  status: "pending" | "in_progress" | "completed" | "rejected";
  feedback?: string;
  score?: number;
  assignedAt: string;
  completedAt?: string;
}

export interface QualityStandard {
  id: string;
  name: string;
  description: string;
  rules: QualityRule[];
  enabled: boolean;
  createdAt: string;
}

export interface QualityRule {
  id: string;
  metric: string;
  operator: "gt" | "gte" | "lt" | "lte" | "eq";
  threshold: number;
  severity: "info" | "warning" | "error";
}

export interface ActivityFeedItem {
  id: string;
  type: "review_completed" | "review_assigned" | "standard_created" | "violation_detected";
  actorName: string;
  description: string;
  metadata: Record<string, unknown>;
  createdAt: string;
}

export function useCreateReviewQueue() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; description: string }) =>
      api.collabHub.createQueue(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["review-queues"] }),
  });
}

export function useReviewQueues() {
  return useQuery({
    queryKey: ["review-queues"],
    queryFn: () =>
      api.collabHub.listQueues() as Promise<ReviewQueue[]>,
  });
}

export function useAssignReview() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { queueId: string; traceId: string; assigneeId: string }) =>
      api.collabHub.assignReview(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["review-queues"] }),
  });
}

export function useCompleteReview() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { id: string; feedback: string; score: number }) =>
      api.collabHub.completeReview(data.id, { feedback: data.feedback, score: data.score }),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["review-queues"] }),
  });
}

export function useCreateQualityStandard() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; description: string; rules: Omit<QualityRule, "id">[] }) =>
      api.collabHub.createStandard(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["quality-standards"] }),
  });
}

export function useQualityStandards() {
  return useQuery({
    queryKey: ["quality-standards"],
    queryFn: () =>
      api.collabHub.listStandards() as Promise<QualityStandard[]>,
  });
}

export function useActivityFeed() {
  return useQuery({
    queryKey: ["activity-feed"],
    queryFn: () =>
      api.collabHub.getActivityFeed() as Promise<ActivityFeedItem[]>,
    refetchInterval: 30000,
  });
}
