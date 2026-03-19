"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface Adapter {
  id: string;
  name: string;
  type: string;
  status: "active" | "inactive" | "error";
  config: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface AdapterTemplate {
  id: string;
  name: string;
  type: string;
  description: string;
  defaultConfig: Record<string, unknown>;
}

export function useAdapters() {
  return useQuery({
    queryKey: ["adapters"],
    queryFn: () => api.adapters.list() as Promise<Adapter[]>,
  });
}

export function useAdapter(id: string) {
  return useQuery({
    queryKey: ["adapters", id],
    queryFn: () => api.adapters.get(id) as Promise<Adapter>,
    enabled: !!id,
  });
}

export function useRegisterAdapter() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; type: string; config: Record<string, unknown> }) =>
      api.adapters.register(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["adapters"] }),
  });
}

export function useUpdateAdapter() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Adapter> }) =>
      api.adapters.update(id, data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["adapters"] }),
  });
}

export function useDeleteAdapter() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.adapters.delete(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["adapters"] }),
  });
}

export function useTestAdapter() {
  return useMutation({
    mutationFn: (id: string) => api.adapters.test(id),
  });
}

export function useAdapterTemplates() {
  return useQuery({
    queryKey: ["adapter-templates"],
    queryFn: () => api.adapters.templates() as Promise<AdapterTemplate[]>,
  });
}

export function useIngestAdapterEvent() {
  return useMutation({
    mutationFn: (data: { adapterId: string; event: Record<string, unknown> }) =>
      api.adapters.ingestEvent(data),
  });
}
