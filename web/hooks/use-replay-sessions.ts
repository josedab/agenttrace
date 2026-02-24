"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface ReplaySession {
  id: string;
  name: string;
  traceId: string;
  status: "recording" | "ready" | "playing" | "archived";
  duration: number;
  eventCount: number;
  branches: ReplayBranch[];
  createdAt: string;
  updatedAt: string;
}

export interface ReplayTimeline {
  sessionId: string;
  events: ReplayEvent[];
  totalDuration: number;
  markers: { time: number; label: string }[];
}

export interface ReplayEvent {
  id: string;
  type: string;
  timestamp: number;
  data: Record<string, unknown>;
  spanId?: string;
  description: string;
}

export interface PlaybackState {
  sessionId: string;
  currentTime: number;
  speed: number;
  playing: boolean;
  currentEventIndex: number;
}

export interface ReplayBranch {
  id: string;
  name: string;
  parentSessionId: string;
  branchPoint: number;
  createdAt: string;
}

export function useReplaySessions() {
  return useQuery({
    queryKey: ["replay-sessions"],
    queryFn: () => api.replaySessions.list() as Promise<ReplaySession[]>,
  });
}

export function useReplayTimeline(sessionId: string | null) {
  return useQuery({
    queryKey: ["replay-timeline", sessionId],
    queryFn: () =>
      api.replaySessions.getTimeline(sessionId!) as Promise<ReplayTimeline>,
    enabled: !!sessionId,
  });
}

export function useCreateReplaySession() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { traceId: string; name: string }) =>
      api.replaySessions.create(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["replay-sessions"] }),
  });
}

export function useBranchSession(sessionId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; branchPoint: number }) =>
      api.replaySessions.branch(sessionId, data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["replay-sessions"] }),
  });
}

export function usePlaybackState(sessionId: string | null) {
  return useQuery({
    queryKey: ["replay-playback", sessionId],
    queryFn: () =>
      api.replaySessions.getPlayback(sessionId!) as Promise<PlaybackState>,
    enabled: !!sessionId,
    refetchInterval: 1000,
  });
}

export function useShareSession() {
  return useMutation({
    mutationFn: (sessionId: string) =>
      api.replaySessions.share(sessionId),
  });
}
