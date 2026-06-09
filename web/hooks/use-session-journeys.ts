'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';

export interface WorkflowPhase {
  id: string;
  name: string;
  type: 'planning' | 'implementation' | 'testing' | 'debugging' | 'review';
  durationSeconds: number;
  cost: number;
  tokenCount: number;
  confidenceScore: number;
}

export interface JourneySummary {
  totalDuration: number;
  totalCost: number;
  phaseCount: number;
}

export interface SessionJourney {
  phases: WorkflowPhase[];
  summary: JourneySummary;
}

export function useSessionJourney(sessionId: string | null) {
  return useQuery({
    queryKey: ['session-journey', sessionId],
    queryFn: () => api.get<SessionJourney>(`/api/public/sessions/${sessionId}/journey`),
    enabled: !!sessionId,
  });
}

export function useSessionPhases(sessionId: string | null) {
  return useQuery({
    queryKey: ['session-phases', sessionId],
    queryFn: () => api.get<WorkflowPhase[]>(`/api/public/sessions/${sessionId}/phases`),
    enabled: !!sessionId,
  });
}

export function useRecentJourneys(limit: number = 10) {
  return useQuery({
    queryKey: ['recent-journeys', limit],
    queryFn: () => api.get<SessionJourney[]>(`/api/public/sessions/journeys/recent?limit=${limit}`),
  });
}
