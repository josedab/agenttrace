"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api";
import { PageHeader } from "@/components/layout/page-header";
import { ScoreList } from "@/components/scores/score-list";
import { ScoreListSkeleton } from "@/components/scores/score-list-skeleton";
import { ScoreFilters } from "@/components/scores/score-filters";

export default function ScoresPage() {
  const [filters, setFilters] = React.useState({
    scoreName: "",
    source: "",
    minScore: undefined as number | undefined,
    maxScore: undefined as number | undefined,
  });

  const { data: scores, isLoading, error } = useQuery({
    queryKey: ["scores", filters],
    queryFn: () => api.scores.list(filters),
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Scores"
        description="View and analyze scores across all traces."
      />

      <ScoreFilters filters={filters} onFiltersChange={setFilters} />

      {isLoading ? (
        <ScoreListSkeleton />
      ) : error ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="text-destructive">Failed to load scores</p>
          <p className="text-sm text-muted-foreground mt-1">
            Please try again later
          </p>
        </div>
      ) : scores && scores.length > 0 ? (
        <ScoreList scores={scores} />
      ) : (
        <div className="flex flex-col items-center justify-center py-12 text-center border rounded-lg bg-muted/20">
          <h3 className="text-lg font-semibold">No scores found</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Scores will appear here once traces are evaluated.
          </p>
        </div>
      )}
    </div>
  );
}
