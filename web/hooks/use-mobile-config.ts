"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useMobileDashboard() {
  return useQuery({
    queryKey: ["mobile-dashboard"],
    queryFn: () => api.mobile.getDashboard(),
  });
}

export function useMobileNotifications() {
  return useQuery({
    queryKey: ["mobile-notifications"],
    queryFn: () => api.mobile.listNotifications(),
  });
}

export function useRegisterDevice() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { deviceToken: string; platform: string; deviceName?: string }) =>
      api.mobile.registerDevice(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["mobile-dashboard"] }),
  });
}
