"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useEvalMarketplaceDatasets(params?: {
  search?: string;
  category?: string;
  page?: number;
  limit?: number;
}) {
  return useQuery({
    queryKey: ["eval-marketplace-datasets", params],
    queryFn: () => {
      const searchParams = new URLSearchParams();
      if (params?.search) searchParams.set("search", params.search);
      if (params?.category) searchParams.set("category", params.category);
      if (params?.page) searchParams.set("page", String(params.page));
      if (params?.limit) searchParams.set("limit", String(params.limit));
      const qs = searchParams.toString();
      return api.get(`/api/public/eval-marketplace/datasets${qs ? `?${qs}` : ""}`);
    },
  });
}

export function useEvalMarketplaceDataset(id: string | null) {
  return useQuery({
    queryKey: ["eval-marketplace-dataset", id],
    queryFn: () => api.get(`/api/public/eval-marketplace/datasets/${id}`),
    enabled: !!id,
  });
}

export function usePublishEvalDataset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { datasetId: string; description?: string; tags?: string[] }) =>
      api.post("/api/public/eval-marketplace/datasets", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["eval-marketplace-datasets"] });
    },
  });
}

export function useImportEvalDataset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { marketplaceDatasetId: string }) =>
      api.post("/api/public/eval-marketplace/datasets/import", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["eval-marketplace-datasets"] });
    },
  });
}

export function useEvalMarketplaceCategories() {
  return useQuery({
    queryKey: ["eval-marketplace-categories"],
    queryFn: () => api.get("/api/public/eval-marketplace/categories"),
  });
}

export function useRateEvalDataset() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { datasetId: string; rating: number; review?: string }) =>
      api.post(`/api/public/eval-marketplace/datasets/${data.datasetId}/rate`, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["eval-marketplace-dataset", variables.datasetId],
      });
    },
  });
}
