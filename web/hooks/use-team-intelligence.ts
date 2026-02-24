"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface TeamDashboard {
  totalMembers: number;
  activeAgents: number;
  totalCost: number;
  totalTraces: number;
  memberStats: { name: string; traces: number; cost: number }[];
}

export interface ROICalculation {
  timeSavedHours: number;
  costSavings: number;
  totalAICost: number;
  netROI: number;
  roiPercentage: number;
}

export function useTeamDashboard() {
  return useQuery({
    queryKey: ["team", "dashboard"],
    queryFn: () => api.team.getDashboard(),
  });
}

export function useTeamROI(hourlyRate?: number) {
  return useQuery({
    queryKey: ["team", "roi", hourlyRate],
    queryFn: () => api.team.calculateROI(hourlyRate),
    enabled: hourlyRate !== undefined && hourlyRate > 0,
  });
}
