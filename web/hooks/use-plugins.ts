"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function usePlugins() {
  return useQuery({
    queryKey: ["plugins"],
    queryFn: () => api.plugins.list(),
  });
}

export function usePlugin(id: string) {
  return useQuery({
    queryKey: ["plugins", id],
    queryFn: () => api.plugins.get(id),
    enabled: !!id,
  });
}

export function useInstallPlugin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; source: string; version?: string; config?: Record<string, any> }) =>
      api.plugins.install(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["plugins"] }),
  });
}

export function useActivatePlugin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.plugins.activate(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["plugins"] }),
  });
}

export function useDisablePlugin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.plugins.disable(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["plugins"] }),
  });
}

export function useExecutePlugin() {
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Record<string, any> }) =>
      api.plugins.execute(id, data),
  });
}

export function useUninstallPlugin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.plugins.uninstall(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["plugins"] }),
  });
}
