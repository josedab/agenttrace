"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface DiscoveryDashboard {
  frameworks: DiscoveredFramework[];
  components: DiscoveredComponent[];
  totalDiscovered: number;
  lastScanAt: string;
  config: DiscoveryConfig;
}

export interface DiscoveredFramework {
  id: string;
  name: string;
  version: string;
  type: "llm" | "agent" | "tool" | "vector_db" | "orchestrator";
  detected: boolean;
  instrumented: boolean;
  detectedAt: string;
}

export interface DiscoveredComponent {
  id: string;
  frameworkId: string;
  name: string;
  type: string;
  path: string;
  instrumented: boolean;
}

export interface DiscoveryConfig {
  autoScan: boolean;
  scanInterval: number;
  includePaths: string[];
  excludePaths: string[];
  autoInstrument: boolean;
}

export function useScanProject() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () =>
      api.autoDiscovery.scan(),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["discovery-dashboard"] }),
  });
}

export function useDiscoveryDashboard() {
  return useQuery({
    queryKey: ["discovery-dashboard"],
    queryFn: () =>
      api.autoDiscovery.scan() as Promise<DiscoveryDashboard>,
    refetchInterval: 30000,
  });
}

export function useDiscoveryFramework(id: string | null) {
  return useQuery({
    queryKey: ["discovery-framework", id],
    queryFn: () =>
      api.autoDiscovery.getFramework(id!) as Promise<DiscoveredFramework>,
    enabled: !!id,
  });
}

export function useUpdateDiscoveryConfig() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<DiscoveryConfig>) =>
      api.autoDiscovery.updateConfig(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["discovery-dashboard"] }),
  });
}

export function useToggleInstrumentation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { id: string; enabled: boolean }) =>
      api.autoDiscovery.toggleInstrumentation(data.id, data.enabled),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["discovery-dashboard"] }),
  });
}
