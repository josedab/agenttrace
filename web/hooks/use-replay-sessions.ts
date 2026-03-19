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

export function useRecordReplayEvents(sessionId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (events: { type: string; data: Record<string, unknown>; durationMs?: number; fileDelta?: { path: string; operation: string; before?: string; after?: string } }[]) =>
      api.replaySessions.recordEvents(sessionId, events),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["replay-timeline", sessionId] });
    },
  });
}

export function useControlPlayback(sessionId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (cmd: { action: string; eventIndex?: number; speed?: number }) =>
      api.replaySessions.control(sessionId, cmd),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["replay-playback", sessionId] });
    },
  });
}

export function useReplayFileState(sessionId: string | null, eventIndex: number) {
  return useQuery({
    queryKey: ["replay-file-state", sessionId, eventIndex],
    queryFn: () =>
      api.replaySessions.getFileState(sessionId!, eventIndex) as Promise<{
        sessionId: string;
        eventIndex: number;
        files: Record<string, string>;
        timestamp: string;
      }>,
    enabled: !!sessionId && eventIndex >= 0,
  });
}

export function useCompleteReplaySession(sessionId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => api.replaySessions.complete(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["replay-sessions"] });
    },
  });
}

export interface UnifiedTimelineEvent {
  id: string;
  type: string;
  timestamp: number;
  data: Record<string, unknown>;
  spanId?: string;
  annotations: { id: string; content: string; createdAt: string }[];
}

export interface UnifiedTimeline {
  sessionId: string;
  events: UnifiedTimelineEvent[];
  totalDuration: number;
}

export interface ReplaySnapshot {
  sessionId: string;
  eventIndex: number;
  state: Record<string, unknown>;
  timestamp: string;
}

export function useUnifiedTimeline(sessionId: string) {
  return useQuery({
    queryKey: ["replay-unified-timeline", sessionId],
    queryFn: () =>
      api.replaySessions.getUnifiedTimeline(sessionId) as Promise<UnifiedTimeline>,
    enabled: !!sessionId,
  });
}

export function useReplaySnapshot(sessionId: string, eventIndex: number) {
  return useQuery({
    queryKey: ["replay-snapshot", sessionId, eventIndex],
    queryFn: () =>
      api.replaySessions.getSnapshot(sessionId, eventIndex) as Promise<ReplaySnapshot>,
    enabled: !!sessionId && eventIndex >= 0,
  });
}

export function useAddReplayAnnotation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ sessionId, data }: { sessionId: string; data: { eventId: string; content: string } }) =>
      api.replaySessions.addAnnotation(sessionId, data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ["replay-unified-timeline", variables.sessionId] });
    },
  });
}
