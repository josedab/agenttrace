"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface CloudOnboarding {
  id: string;
  steps: OnboardingStepStatus[];
  completedSteps: number;
  totalSteps: number;
  status: "in_progress" | "completed";
  startedAt: string;
  completedAt?: string;
}

export interface OnboardingStepStatus {
  id: string;
  name: string;
  description: string;
  status: "pending" | "in_progress" | "completed" | "skipped";
  order: number;
  completedAt?: string;
}

export interface QuickstartConfig {
  framework: string;
  language: string;
  packageManager: string;
  installCommand: string;
  configSnippet: string;
  exampleCode: string;
}

export interface UsageMeter {
  plan: string;
  tracesUsed: number;
  tracesLimit: number;
  spansUsed: number;
  spansLimit: number;
  storageUsedMb: number;
  storageLimitMb: number;
  billingPeriodStart: string;
  billingPeriodEnd: string;
}

export function useOnboarding() {
  return useQuery({
    queryKey: ["cloud-onboarding"],
    queryFn: () =>
      api.cloudOnboarding.get() as Promise<CloudOnboarding>,
  });
}

export function useCompleteStep() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { stepId: string }) =>
      api.cloudOnboarding.completeStep(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["cloud-onboarding"] }),
  });
}

export function useGenerateQuickstart() {
  return useMutation({
    mutationFn: (data: { framework: string; language: string }) =>
      api.cloudOnboarding.generateQuickstart(data) as Promise<QuickstartConfig>,
  });
}

export function useUsage() {
  return useQuery({
    queryKey: ["cloud-usage"],
    queryFn: () =>
      api.cloudOnboarding.getUsage() as Promise<UsageMeter>,
    refetchInterval: 30000,
  });
}

export function useCheckQuota() {
  return useMutation({
    mutationFn: (data: { resource: string; amount: number }) =>
      api.cloudOnboarding.checkQuota(data),
  });
}
