'use client';

import { useQuery } from '@tanstack/react-query';
import { useRouter, useSearchParams } from 'next/navigation';

import { EvalHubDatasets, EvalHubEvaluators } from '@/components/eval-hub/eval-hub-assets';
import { EvalHubLibrary } from '@/components/eval-hub/eval-hub-library';
import { EvalHubOverview } from '@/components/eval-hub/eval-hub-overview';
import { EvalHubRuns } from '@/components/eval-hub/eval-hub-runs';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useEvalHubPackages, useEvalHubRuns } from '@/hooks/use-eval-hub';
import { api } from '@/lib/api';

type EvalHubView = 'overview' | 'evaluators' | 'datasets' | 'library' | 'runs';

const validViews = new Set<EvalHubView>(['overview', 'evaluators', 'datasets', 'library', 'runs']);

export function EvalHubDashboard() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const requestedView = searchParams.get('view') as EvalHubView | null;
  const view = requestedView && validViews.has(requestedView) ? requestedView : 'overview';

  const evaluatorsQuery = useQuery({
    queryKey: ['evaluators'],
    queryFn: () => api.evaluators.list(),
  });
  const datasetsQuery = useQuery({
    queryKey: ['datasets'],
    queryFn: () => api.datasets.list(),
  });
  const packagesQuery = useEvalHubPackages();
  const runsQuery = useEvalHubRuns();

  const setView = (next: string) => {
    const value = next as EvalHubView;
    router.replace(value === 'overview' ? '/evals' : `/evals?view=${value}`);
  };

  return (
    <Tabs value={view} onValueChange={setView} className="space-y-6">
      <TabsList className="h-auto flex-wrap justify-start">
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="evaluators">Evaluators</TabsTrigger>
        <TabsTrigger value="datasets">Datasets</TabsTrigger>
        <TabsTrigger value="library">Community & private</TabsTrigger>
        <TabsTrigger value="runs">Runs</TabsTrigger>
      </TabsList>

      <TabsContent value="overview" className="space-y-6">
        <EvalHubOverview
          evaluatorCount={evaluatorsQuery.data?.length}
          datasetCount={datasetsQuery.data?.length}
          packageCount={packagesQuery.data?.totalCount}
          runCount={runsQuery.data?.totalCount}
          loading={
            evaluatorsQuery.isLoading ||
            datasetsQuery.isLoading ||
            packagesQuery.isLoading ||
            runsQuery.isLoading
          }
          onNavigate={setView}
        />
      </TabsContent>

      <TabsContent value="evaluators">
        <EvalHubEvaluators query={evaluatorsQuery} />
      </TabsContent>

      <TabsContent value="datasets">
        <EvalHubDatasets query={datasetsQuery} />
      </TabsContent>

      <TabsContent value="library">
        <EvalHubLibrary />
      </TabsContent>

      <TabsContent value="runs">
        <EvalHubRuns query={runsQuery} />
      </TabsContent>
    </Tabs>
  );
}
