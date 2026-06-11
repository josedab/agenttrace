'use client';

import * as React from 'react';
import { toast } from 'sonner';
import { GitFork, PackageOpen, Play, Search } from 'lucide-react';

import {
  EvalHubEmptyState,
  EvalHubErrorState,
  EvalHubGridSkeleton,
  EvalHubVisibilityBadge,
} from '@/components/eval-hub/eval-hub-states';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  useEvalHubPackages,
  useForkEvalHubPackage,
  useRunEvalHubPackage,
} from '@/hooks/use-eval-hub';
import { getApiProjectId } from '@/lib/api';
import { createIdempotencyKey } from '@/lib/idempotency';
import type { EvalHubAssetKind, EvalHubPackage } from '@/lib/eval-hub';

const SEARCH_INPUT_ID = 'eval-hub-package-search';
const KIND_SELECT_ID = 'eval-hub-package-kind';

export function EvalHubLibrary() {
  const activeProjectId = getApiProjectId();
  const [query, setQuery] = React.useState('');
  const [kind, setKind] = React.useState<EvalHubAssetKind | 'all'>('all');
  const [pendingPackageId, setPendingPackageId] = React.useState<string | null>(null);
  const deferredQuery = React.useDeferredValue(query);
  const packagesQuery = useEvalHubPackages({
    query: deferredQuery || undefined,
    kind: kind === 'all' ? undefined : kind,
  });
  const forkPackage = useForkEvalHubPackage();
  const runPackage = useRunEvalHubPackage();

  // One key per package version keeps a double-click, or a retry after a failed
  // request, from starting a second run on the server.
  const runKeys = React.useRef(new Map<string, string>());
  const runKeyFor = React.useCallback((pkg: EvalHubPackage) => {
    const attempt = `${pkg.id}:v${pkg.latestVersion}`;
    const existing = runKeys.current.get(attempt);
    if (existing) {
      return existing;
    }
    const created = createIdempotencyKey();
    runKeys.current.set(attempt, created);
    return created;
  }, []);

  const fork = React.useCallback(
    (pkg: EvalHubPackage) => {
      setPendingPackageId(pkg.id);
      forkPackage.mutate(
        { packageId: pkg.id, name: `${pkg.name} fork`, visibility: 'private' },
        {
          onSuccess: () => toast.success('Package forked into this project'),
          onError: (error) => toast.error(error.message),
          onSettled: () => setPendingPackageId(null),
        }
      );
    },
    [forkPackage]
  );

  const run = React.useCallback(
    (pkg: EvalHubPackage) => {
      const attempt = `${pkg.id}:v${pkg.latestVersion}`;
      setPendingPackageId(pkg.id);
      runPackage.mutate(
        {
          packageId: pkg.id,
          name: `${pkg.name} run`,
          idempotencyKey: runKeyFor(pkg),
        },
        {
          onSuccess: (result) => {
            // Ready/running results still identify active work. Keep their key
            // so a quick retry or second click returns the durable server run
            // instead of creating another one.
            if (
              result.status === 'completed' ||
              result.status === 'unsupported' ||
              result.status === 'failed'
            ) {
              runKeys.current.delete(attempt);
            }
            if (result.status === 'unsupported') {
              toast.warning(
                result.capabilityMessage || 'This package requires another execution API.'
              );
            } else {
              toast.success(`Eval Hub run is ${result.status}`);
            }
          },
          onError: (error) => toast.error(error.message),
          onSettled: () => setPendingPackageId(null),
        }
      );
    },
    [runKeyFor, runPackage]
  );

  const packages = packagesQuery.data?.packages ?? [];

  return (
    <div className="space-y-5">
      <div className="flex flex-col gap-3 sm:flex-row">
        <div className="flex-1 space-y-1.5">
          <Label htmlFor={SEARCH_INPUT_ID}>Search packages</Label>
          <div className="relative">
            <Search
              aria-hidden="true"
              className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id={SEARCH_INPUT_ID}
              type="search"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search versioned packages"
              className="pl-9"
            />
          </div>
        </div>
        <div className="space-y-1.5 sm:w-48">
          <Label htmlFor={KIND_SELECT_ID}>Asset type</Label>
          <Select value={kind} onValueChange={(value) => setKind(value as typeof kind)}>
            <SelectTrigger id={KIND_SELECT_ID} className="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All asset types</SelectItem>
              <SelectItem value="dataset">Datasets</SelectItem>
              <SelectItem value="evaluator">Evaluators</SelectItem>
              <SelectItem value="prompt">Prompts</SelectItem>
              <SelectItem value="experiment">Experiments</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {packagesQuery.isLoading ? (
        <EvalHubGridSkeleton />
      ) : packagesQuery.isError ? (
        <EvalHubErrorState
          label="Eval Hub packages could not be loaded."
          onRetry={() => packagesQuery.refetch()}
        />
      ) : packages.length > 0 ? (
        <ul className="grid list-none gap-4 p-0 md:grid-cols-2 xl:grid-cols-3">
          {packages.map((pkg) => (
            <li key={pkg.id} className="flex">
              <PackageCard
                pkg={pkg}
                owned={pkg.ownerProjectId === activeProjectId}
                busy={pendingPackageId === pkg.id}
                onFork={fork}
                onRun={run}
              />
            </li>
          ))}
        </ul>
      ) : (
        <EvalHubEmptyState
          icon={PackageOpen}
          title="No packages match this view"
          description="Publish a project dataset or adjust the search filters."
        />
      )}
    </div>
  );
}

const PackageCard = React.memo(function PackageCard({
  pkg,
  owned,
  busy,
  onFork,
  onRun,
}: {
  pkg: EvalHubPackage;
  owned: boolean;
  busy: boolean;
  onFork: (pkg: EvalHubPackage) => void;
  onRun: (pkg: EvalHubPackage) => void;
}) {
  const actionLabel = owned
    ? `Run ${pkg.name} in this project`
    : `Fork ${pkg.name} into this project`;
  return (
    <Card className="flex w-full flex-col">
      <CardHeader className="space-y-3">
        <div className="flex items-start justify-between gap-3">
          <Badge variant="outline">{pkg.kind}</Badge>
          <EvalHubVisibilityBadge visibility={pkg.visibility} />
        </div>
        <div>
          <CardTitle className="text-base">{pkg.name}</CardTitle>
          <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">
            {pkg.description || 'No description'}
          </p>
        </div>
      </CardHeader>
      <CardContent className="mt-auto space-y-4">
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>Version {pkg.latestVersion}</span>
          <span>{pkg.forkedFromPackageId ? 'Fork with provenance' : 'Original package'}</span>
        </div>
        <Button
          className="w-full"
          variant={owned ? 'default' : 'outline'}
          onClick={() => (owned ? onRun(pkg) : onFork(pkg))}
          disabled={busy}
          aria-label={actionLabel}
          aria-busy={busy}
        >
          {owned ? (
            <>
              <Play aria-hidden="true" className="mr-2 h-4 w-4" />
              Run in project
            </>
          ) : (
            <>
              <GitFork aria-hidden="true" className="mr-2 h-4 w-4" />
              Fork into project
            </>
          )}
        </Button>
      </CardContent>
    </Card>
  );
});
