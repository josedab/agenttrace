"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useSessionJourney(sessionId: string | null) {
  return useQuery({
    queryKey: ["session-journey", sessionId],
    queryFn: () => api.get(`/api/public/sessions/${sessionId}/journey`),
    enabled: !!sessionId,
  });
}

export function useSessionPhases(sessionId: string | null) {
  return useQuery({
    queryKey: ["session-phases", sessionId],
    queryFn: () => api.get(`/api/public/sessions/${sessionId}/phases`),
    enabled: !!sessionId,
  });
}

export function useRecentJourneys(limit: number = 10) {
  return useQuery({
    queryKey: ["recent-journeys", limit],
    queryFn: () => api.get(`/api/public/sessions/journeys/recent?limit=${limit}`),
  });
}
