'use client';

import * as React from 'react';
import Link from 'next/link';
import type { UseQueryResult } from '@tanstack/react-query';
import { toast } from 'sonner';
import { ArrowRight, Boxes, Database, FlaskConical, Loader2 } from 'lucide-react';

import { CreateDatasetDialog } from '@/components/datasets/create-dataset-dialog';
import {
  EvalHubEmptyState,
  EvalHubErrorState,
  EvalHubGridSkeleton,
} from '@/components/eval-hub/eval-hub-states';
import { CreateEvaluatorDialog } from '@/components/evals/create-evaluator-dialog';
import { EvaluatorList } from '@/components/evals/evaluator-list';
import { EvaluatorListSkeleton } from '@/components/evals/evaluator-list-skeleton';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { usePublishEvalHubPackage } from '@/hooks/use-eval-hub';
import type { Dataset, Evaluator } from '@/lib/api';
import type { EvalHubVisibility } from '@/lib/eval-hub';

const PUBLISH_DATASET_SELECT_ID = 'eval-hub-publish-dataset';
const PUBLISH_VISIBILITY_SELECT_ID = 'eval-hub-publish-visibility';

export function EvalHubEvaluators({ query }: { query: UseQueryResult<Evaluator[], Error> }) {
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="font-semibold">Evaluators</h2>
          <p className="text-sm text-muted-foreground">
            Automated and human scoring remains project scoped.
          </p>
        </div>
        <CreateEvaluatorDialog />
      </div>
      {query.isLoading ? (
        <EvaluatorListSkeleton />
      ) : query.isError ? (
        <EvalHubErrorState
          label="Evaluators could not be loaded."
          onRetry={() => query.refetch()}
        />
      ) : query.data && query.data.length > 0 ? (
        <EvaluatorList evaluators={query.data} />
      ) : (
        <EvalHubEmptyState
          icon={FlaskConical}
          title="No evaluators yet"
          description="Create an evaluator before publishing or running evaluation workflows."
          action={<CreateEvaluatorDialog />}
        />
      )}
    </div>
  );
}

export function EvalHubDatasets({ query }: { query: UseQueryResult<Dataset[], Error> }) {
  const datasets = query.data ?? [];
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="font-semibold">Datasets</h2>
          <p className="text-sm text-muted-foreground">
            Curate test cases locally, then publish an immutable package version.
          </p>
        </div>
        <div className="flex gap-2">
          <PublishDatasetDialog datasets={datasets} />
          <CreateDatasetDialog />
        </div>
      </div>
      {query.isLoading ? (
        <EvalHubGridSkeleton />
      ) : query.isError ? (
        <EvalHubErrorState label="Datasets could not be loaded." onRetry={() => query.refetch()} />
      ) : datasets.length > 0 ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {datasets.map((dataset) => (
            <Card key={dataset.id}>
              <CardHeader>
                <CardTitle className="text-base">{dataset.name}</CardTitle>
                <p className="line-clamp-2 text-sm text-muted-foreground">
                  {dataset.description || 'No description'}
                </p>
              </CardHeader>
              <CardContent className="flex items-center justify-between gap-3">
                <div className="text-xs text-muted-foreground">
                  {dataset.itemCount} items · {dataset.runCount} runs
                </div>
                <Button variant="outline" size="sm" asChild>
                  <Link href={`/datasets/${dataset.id}`}>
                    Open
                    <ArrowRight className="ml-2 h-3.5 w-3.5" />
                  </Link>
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <EvalHubEmptyState
          icon={Database}
          title="No datasets yet"
          description="Create a dataset to build repeatable evaluation runs."
          action={<CreateDatasetDialog />}
        />
      )}
    </div>
  );
}

function PublishDatasetDialog({ datasets }: { datasets: Dataset[] }) {
  const [open, setOpen] = React.useState(false);
  const [datasetId, setDatasetId] = React.useState('');
  const [visibility, setVisibility] = React.useState<EvalHubVisibility>('private');
  const publish = usePublishEvalHubPackage();

  const submit = () => {
    if (!datasetId) return;
    publish.mutate(
      {
        kind: 'dataset',
        sourceResourceId: datasetId,
        visibility,
        versionNote: 'Published from the Eval Hub workspace',
      },
      {
        onSuccess: () => {
          toast.success('Dataset package published');
          setOpen(false);
          setDatasetId('');
        },
        onError: (error) => toast.error(error.message),
      }
    );
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" disabled={datasets.length === 0}>
          <Boxes className="mr-2 h-4 w-4" />
          Publish
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Publish dataset package</DialogTitle>
          <DialogDescription>
            Publishes an immutable, source-link-free snapshot of up to 5,000 items.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor={PUBLISH_DATASET_SELECT_ID}>Dataset</Label>
            <Select value={datasetId} onValueChange={setDatasetId}>
              <SelectTrigger id={PUBLISH_DATASET_SELECT_ID}>
                <SelectValue placeholder="Choose a dataset" />
              </SelectTrigger>
              <SelectContent>
                {datasets.map((dataset) => (
                  <SelectItem key={dataset.id} value={dataset.id}>
                    {dataset.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor={PUBLISH_VISIBILITY_SELECT_ID}>Visibility</Label>
            <Select
              value={visibility}
              onValueChange={(value) => setVisibility(value as EvalHubVisibility)}
            >
              <SelectTrigger id={PUBLISH_VISIBILITY_SELECT_ID}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="private">Private project</SelectItem>
                <SelectItem value="organization">Organization</SelectItem>
                <SelectItem value="public">Public</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button onClick={submit} disabled={!datasetId || publish.isPending}>
            {publish.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
            Publish version
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
