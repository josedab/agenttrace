"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface TraceReview {
  id: string;
  traceId: string;
  status: "pending" | "approved" | "rejected";
  reviewer: string;
  comments: ReviewComment[];
  createdAt: string;
  updatedAt: string;
}

export interface ReviewComment {
  id: string;
  reviewId: string;
  author: string;
  content: string;
  createdAt: string;
}

export interface ReviewQueue {
  id: string;
  name: string;
  filters: Record<string, unknown>;
  assignees: string[];
  pendingCount: number;
}

export interface NotificationIntegration {
  id: string;
  type: "slack" | "email" | "webhook";
  config: Record<string, unknown>;
  enabled: boolean;
}

export function useTraceReviews() {
  return useQuery({
    queryKey: ["trace-reviews"],
    queryFn: () => api.traceReviews.list() as Promise<TraceReview[]>,
  });
}

export function useTraceReview(id: string) {
  return useQuery({
    queryKey: ["trace-reviews", id],
    queryFn: () => api.traceReviews.get(id) as Promise<TraceReview>,
    enabled: !!id,
  });
}

export function useCreateReview() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { traceId: string; reviewer?: string }) =>
      api.traceReviews.create(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["trace-reviews"] }),
  });
}

export function useApproveReview() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.traceReviews.approve(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["trace-reviews"] }),
  });
}

export function useReviewComments(reviewId: string) {
  return useQuery({
    queryKey: ["trace-review-comments", reviewId],
    queryFn: () => api.traceReviews.listComments(reviewId) as Promise<ReviewComment[]>,
    enabled: !!reviewId,
  });
}

export function useAddReviewComment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ reviewId, data }: { reviewId: string; data: { content: string } }) =>
      api.traceReviews.addComment(reviewId, data),
    onSuccess: (_data, variables) =>
      queryClient.invalidateQueries({ queryKey: ["trace-review-comments", variables.reviewId] }),
  });
}

export function useReviewQueues() {
  return useQuery({
    queryKey: ["review-queues"],
    queryFn: () => api.traceReviews.listQueues() as Promise<ReviewQueue[]>,
  });
}

export function useCreateReviewQueue() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; filters: Record<string, unknown>; assignees: string[] }) =>
      api.traceReviews.createQueue(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["review-queues"] }),
  });
}

export function useNotificationIntegrations() {
  return useQuery({
    queryKey: ["notification-integrations"],
    queryFn: () => api.traceReviews.listNotificationIntegrations() as Promise<NotificationIntegration[]>,
  });
}

export function useAddNotificationIntegration() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { type: string; config: Record<string, unknown> }) =>
      api.traceReviews.addNotificationIntegration(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["notification-integrations"] }),
  });
}
