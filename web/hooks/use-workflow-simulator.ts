"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export interface WorkflowDefinition {
  id: string;
  name: string;
  description: string;
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  version: number;
  createdAt: string;
  updatedAt: string;
}

export interface WorkflowNode {
  id: string;
  type: "agent" | "tool" | "condition" | "input" | "output";
  label: string;
  config: Record<string, unknown>;
  position: { x: number; y: number };
}

export interface WorkflowEdge {
  id: string;
  source: string;
  target: string;
  condition?: string;
  label?: string;
}

export interface WorkflowSimulation {
  id: string;
  workflowId: string;
  status: "pending" | "running" | "completed" | "failed";
  steps: SimulationStepResult[];
  startedAt: string;
  completedAt?: string;
}

export interface SimulationStepResult {
  nodeId: string;
  nodeName: string;
  status: "success" | "failure" | "skipped";
  input: Record<string, unknown>;
  output: Record<string, unknown>;
  durationMs: number;
  error?: string;
}

export function useWorkflows() {
  return useQuery({
    queryKey: ["workflows"],
    queryFn: () =>
      api.workflowSimulator.list() as Promise<WorkflowDefinition[]>,
  });
}

export function useCreateWorkflow() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Omit<WorkflowDefinition, "id" | "version" | "createdAt" | "updatedAt">) =>
      api.workflowSimulator.create(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["workflows"] }),
  });
}

export function useWorkflow(id: string | null) {
  return useQuery({
    queryKey: ["workflow", id],
    queryFn: () =>
      api.workflowSimulator.get(id!) as Promise<WorkflowDefinition>,
    enabled: !!id,
  });
}

export function useRunSimulation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { workflowId: string; input?: Record<string, unknown> }) =>
      api.workflowSimulator.simulate(data),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["simulations"] }),
  });
}

export function useSimulations(workflowId: string | null) {
  return useQuery({
    queryKey: ["simulations", workflowId],
    queryFn: () =>
      api.workflowSimulator.listSimulations(workflowId!) as Promise<WorkflowSimulation[]>,
    enabled: !!workflowId,
  });
}

export function useValidateWorkflow() {
  return useMutation({
    mutationFn: (data: { nodes: WorkflowNode[]; edges: WorkflowEdge[] }) =>
      api.workflowSimulator.validate(data),
  });
}
