"use client";

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useDatasets() {
  return useQuery({
    queryKey: ["datasets"],
    queryFn: () => api.datasets.list(),
  });
}

export function useDataset(datasetId: string) {
  return useQuery({
    queryKey: ["dataset", datasetId],
    queryFn: () => api.datasets.get(datasetId),
    enabled: !!datasetId,
  });
}

export function useDatasetItems(datasetId: string) {
  return useInfiniteQuery({
    queryKey: ["dataset-items", datasetId],
    queryFn: async ({ pageParam }) => {
      const response = await api.datasets.listItems(datasetId, {
        cursor: pageParam,
        limit: 50,
      });
      return response;
    },
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor,
    enabled: !!datasetId,
  });
}

export function useDatasetRuns(datasetId: string) {
  return useQuery({
    queryKey: ["dataset-runs", datasetId],
    queryFn: () => api.datasets.listRuns(datasetId),
    enabled: !!datasetId,
  });
}

export function useDatasetRun(datasetId: string, runId: string) {
  return useQuery({
    queryKey: ["dataset-run", datasetId, runId],
    queryFn: () => api.datasets.getRun(datasetId, runId),
    enabled: !!datasetId && !!runId,
    refetchInterval: (query) => {
      // Keep polling while running
      if (query.state.data?.status === "RUNNING") {
        return 3000;
      }
      return false;
    },
  });
}

export function useCreateDataset() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      name: string;
      description?: string;
      metadata?: Record<string, any>;
    }) => api.datasets.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["datasets"] });
    },
  });
}

export function useUpdateDataset() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      datasetId,
      data,
    }: {
      datasetId: string;
      data: {
        name?: string;
        description?: string;
        metadata?: Record<string, any>;
      };
    }) => api.datasets.update(datasetId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["dataset", variables.datasetId] });
      queryClient.invalidateQueries({ queryKey: ["datasets"] });
    },
  });
}

export function useDeleteDataset() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (datasetId: string) => api.datasets.delete(datasetId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["datasets"] });
    },
  });
}

export function useAddDatasetItem() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      datasetId,
      data,
    }: {
      datasetId: string;
      data: {
        input: any;
        expectedOutput?: any;
        metadata?: Record<string, any>;
      };
    }) => api.datasets.addItem(datasetId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["dataset-items", variables.datasetId] });
      queryClient.invalidateQueries({ queryKey: ["dataset", variables.datasetId] });
    },
  });
}

export function useUpdateDatasetItem() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      datasetId,
      itemId,
      data,
    }: {
      datasetId: string;
      itemId: string;
      data: {
        input?: any;
        expectedOutput?: any;
        metadata?: Record<string, any>;
      };
    }) => api.datasets.updateItem(datasetId, itemId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["dataset-items", variables.datasetId] });
    },
  });
}

export function useDeleteDatasetItem() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      datasetId,
      itemId,
    }: {
      datasetId: string;
      itemId: string;
    }) => api.datasets.deleteItem(datasetId, itemId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["dataset-items", variables.datasetId] });
      queryClient.invalidateQueries({ queryKey: ["dataset", variables.datasetId] });
    },
  });
}

export function useCreateDatasetRun() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      datasetId,
      data,
    }: {
      datasetId: string;
      data: {
        name: string;
        evaluatorId?: string;
        metadata?: Record<string, any>;
      };
    }) => api.datasets.createRun(datasetId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["dataset-runs", variables.datasetId] });
    },
  });
}
