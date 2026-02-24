"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useCustomMetrics() {
  return useQuery({
    queryKey: ["custom-metrics"],
    queryFn: () => api.customMetrics.list(),
  });
}

export function useCreateMetric() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => api.customMetrics.create(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["custom-metrics"] }),
  });
}

export function useMetricValues(id: string) {
  return useQuery({
    queryKey: ["custom-metrics", id, "values"],
    queryFn: () => api.customMetrics.getValues(id),
    enabled: !!id,
  });
}

export function useMetricDashboards() {
  return useQuery({
    queryKey: ["custom-metrics", "dashboards"],
    queryFn: () => api.customMetrics.listDashboards(),
  });
}
