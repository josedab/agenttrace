"use client";

import * as React from "react";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api";
import { PageHeader } from "@/components/layout/page-header";
import { EvaluatorDetail } from "@/components/evals/evaluator-detail";
import { EvaluatorDetailSkeleton } from "@/components/evals/evaluator-detail-skeleton";

export default function EvaluatorDetailPage() {
  const params = useParams();
  const evalId = params.evalId as string;

  const { data: evaluator, isLoading, error } = useQuery({
    queryKey: ["evaluator", evalId],
    queryFn: () => api.evaluators.get(evalId),
  });

  if (isLoading) {
    return <EvaluatorDetailSkeleton />;
  }

  if (error || !evaluator) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-destructive">Failed to load evaluator</p>
        <p className="text-sm text-muted-foreground mt-1">
          The evaluator may have been deleted or you don't have access.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={evaluator.name}
        description={evaluator.description || "No description"}
      />
      <EvaluatorDetail evaluator={evaluator} />
    </div>
  );
}
