"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface SkillScore {
  score: number;
  confidence: number;
}

export interface SkillProfile {
  agentName: string;
  skills: Record<string, SkillScore>;
  totalTraces: number;
  lastUpdated: string;
}

export function useSkillProfiles() {
  return useQuery({
    queryKey: ["skill-profiles"],
    queryFn: () => api.skillProfiles.list(),
  });
}

export function useSkillProfile(agentName: string) {
  return useQuery({
    queryKey: ["skill-profiles", agentName],
    queryFn: () => api.skillProfiles.get(agentName),
    enabled: !!agentName,
  });
}

export function useCompareAgents(agents: string[]) {
  return useQuery({
    queryKey: ["skill-profiles", "compare", agents],
    queryFn: () => api.skillProfiles.compare(agents),
    enabled: agents.length >= 2,
  });
}
