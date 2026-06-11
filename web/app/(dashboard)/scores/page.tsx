'use client';

import * as React from 'react';
import { useQuery } from '@tanstack/react-query';

import { api } from '@/lib/api';
import { PageHeader } from '@/components/layout/page-header';
import { ScoreList } from '@/components/scores/score-list';
import { ScoreListSkeleton } from '@/components/scores/score-list-skeleton';
import { ScoreFilters } from '@/components/scores/score-filters';

export default function ScoresPage() {
  const [filters, setFilters] = React.useState<{
    scoreName: string;
    source: '' | 'API' | 'ANNOTATION' | 'EVAL';
    minScore?: number;
    maxScore?: number;
  }>({
    scoreName: '',
    source: '',
    minScore: undefined,
    maxScore: undefined,
  });

  const {
    data: scores,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['scores', filters],
    queryFn: () => api.scores.list(filters),
  });

  return (
    <div className="space-y-6">
      <PageHeader title="Scores" description="View and analyze scores across all traces." />

      <ScoreFilters filters={filters} onFiltersChange={setFilters} />

      {isLoading ? (
        <ScoreListSkeleton />
      ) : error ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="text-destructive">Failed to load scores</p>
          <p className="mt-1 text-sm text-muted-foreground">Please try again later</p>
        </div>
      ) : scores && scores.scores.length > 0 ? (
        <ScoreList scores={scores.scores} />
      ) : (
        <div className="flex flex-col items-center justify-center rounded-lg border bg-muted/20 py-12 text-center">
          <h3 className="text-lg font-semibold">No scores found</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Scores will appear here once traces are evaluated.
          </p>
        </div>
      )}
    </div>
  );
}
