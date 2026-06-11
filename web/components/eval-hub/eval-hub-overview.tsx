'use client';

import { Boxes, Database, FlaskConical, Play } from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';

export function EvalHubOverview({
  evaluatorCount,
  datasetCount,
  packageCount,
  runCount,
  loading,
  onNavigate,
}: {
  evaluatorCount?: number;
  datasetCount?: number;
  packageCount?: number;
  runCount?: number;
  loading: boolean;
  onNavigate: (view: string) => void;
}) {
  const cards = [
    { label: 'Evaluators', value: evaluatorCount, view: 'evaluators', icon: FlaskConical },
    { label: 'Datasets', value: datasetCount, view: 'datasets', icon: Database },
    { label: 'Packages', value: packageCount, view: 'library', icon: Boxes },
    { label: 'Runs', value: runCount, view: 'runs', icon: Play },
  ];

  return (
    <>
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {cards.map((card) => (
          <button
            key={card.label}
            type="button"
            onClick={() => onNavigate(card.view)}
            className="rounded-xl border bg-card p-4 text-left transition-colors hover:border-foreground/30"
          >
            <div className="flex items-center justify-between">
              <p className="text-xs font-semibold uppercase tracking-[0.15em] text-muted-foreground">
                {card.label}
              </p>
              <card.icon className="h-4 w-4 text-muted-foreground" />
            </div>
            {loading ? (
              <Skeleton className="mt-4 h-8 w-16" />
            ) : (
              <p className="mt-4 text-3xl font-semibold tabular-nums">
                {(card.value ?? 0).toLocaleString()}
              </p>
            )}
          </button>
        ))}
      </div>

      <Card className="overflow-hidden">
        <div className="grid md:grid-cols-[1.2fr_0.8fr]">
          <CardContent className="p-6">
            <Badge variant="outline">One Eval Hub</Badge>
            <h2 className="mt-4 max-w-xl text-2xl font-semibold tracking-tight">
              Build locally. Publish deliberately. Fork with provenance.
            </h2>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-muted-foreground">
              Evaluators, datasets, prompts, and experiments share one versioned package model.
              Public assets must be forked before project execution.
            </p>
          </CardContent>
          <div className="border-t bg-muted/30 p-6 md:border-l md:border-t-0">
            <ol className="space-y-4 text-sm">
              <li className="flex gap-3">
                <span className="font-mono text-muted-foreground">01</span>
                Create a project-owned asset.
              </li>
              <li className="flex gap-3">
                <span className="font-mono text-muted-foreground">02</span>
                Publish an immutable version with visibility.
              </li>
              <li className="flex gap-3">
                <span className="font-mono text-muted-foreground">03</span>
                Fork before running in another project.
              </li>
            </ol>
          </div>
        </div>
      </Card>
    </>
  );
}
