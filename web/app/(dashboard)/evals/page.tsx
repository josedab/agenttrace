"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api";
import { PageHeader } from "@/components/layout/page-header";
import { EvaluatorList } from "@/components/evals/evaluator-list";
import { EvaluatorListSkeleton } from "@/components/evals/evaluator-list-skeleton";
import { CreateEvaluatorDialog } from "@/components/evals/create-evaluator-dialog";

export default function EvaluatorsPage() {
  const { data: evaluators, isLoading, error } = useQuery({
    queryKey: ["evaluators"],
    queryFn: () => api.evaluators.list(),
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Evaluators"
        description="Configure automated and human evaluation for your traces."
        actions={<CreateEvaluatorDialog />}
      />

      {isLoading ? (
        <EvaluatorListSkeleton />
      ) : error ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="text-destructive">Failed to load evaluators</p>
          <p className="text-sm text-muted-foreground mt-1">
            Please try again later
          </p>
        </div>
      ) : evaluators && evaluators.length > 0 ? (
        <EvaluatorList evaluators={evaluators} />
      ) : (
        <div className="flex flex-col items-center justify-center py-12 text-center border rounded-lg bg-muted/20">
          <h3 className="text-lg font-semibold">No evaluators yet</h3>
          <p className="text-sm text-muted-foreground mt-1 mb-4">
            Create an evaluator to automatically score your traces.
          </p>
          <CreateEvaluatorDialog />
        </div>
      )}
    </div>
  );
}
