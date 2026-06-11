"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import {
  Shield,
  Activity,
  GitPullRequest,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  BarChart3,
  RefreshCw,
} from "lucide-react";

interface DashboardStats {
  totalBaselines: number;
  totalRuns: number;
  totalGateConfigs: number;
  passRate: number;
  blockedPrs: number;
  recentRegressions: number;
}

interface GateConfig {
  id: string;
  name: string;
  baselineId: string;
  blockOnSeverity: string;
  confidenceLevel: number;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

interface RegressionHistoryEntry {
  id: string;
  branch: string;
  commitSha: string;
  prNumber?: number;
  passed: boolean;
  severity: string;
  metricDeltas: Record<string, number>;
  blockedPr: boolean;
  createdAt: string;
}

function severityBadgeVariant(severity: string) {
  switch (severity) {
    case "critical":
      return "destructive";
    case "major":
      return "destructive";
    case "minor":
      return "secondary";
    default:
      return "outline";
  }
}

export function PromptCIDashboard() {
  const queryClient = useQueryClient();

  const { data: stats, isLoading: statsLoading } = useQuery<DashboardStats>({
    queryKey: ["prompt-ci-stats"],
    queryFn: () => api.promptCI.getDashboardStats() as Promise<DashboardStats>,
    refetchInterval: 30000,
  });

  const { data: gateConfigs, isLoading: configsLoading } = useQuery<
    GateConfig[]
  >({
    queryKey: ["prompt-ci-gates"],
    queryFn: () => api.promptCI.listGateConfigs() as Promise<GateConfig[]>,
  });

  const { data: history, isLoading: historyLoading } = useQuery<
    RegressionHistoryEntry[]
  >({
    queryKey: ["prompt-ci-history"],
    queryFn: () => api.promptCI.getRegressionHistory() as Promise<RegressionHistoryEntry[]>,
  });

  const toggleGateMutation = useMutation({
    mutationFn: (config: GateConfig) =>
      api.promptCI.updateGateConfig(config.id, {
        enabled: !config.enabled,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["prompt-ci-gates"] });
    },
  });

  if (statsLoading) {
    return <DashboardSkeleton />;
  }

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Baselines</CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats?.totalBaselines ?? 0}
            </div>
            <p className="text-xs text-muted-foreground">
              Active performance baselines
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">CI Runs</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.totalRuns ?? 0}</div>
            <p className="text-xs text-muted-foreground">
              Total evaluation runs
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Pass Rate</CardTitle>
            <CheckCircle2 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(stats?.passRate ?? 100).toFixed(1)}%
            </div>
            <p className="text-xs text-muted-foreground">
              Gate evaluation pass rate
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Blocked PRs</CardTitle>
            <GitPullRequest className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.blockedPrs ?? 0}</div>
            <p className="text-xs text-muted-foreground">
              PRs blocked by regressions
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Gate Configs Table */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Shield className="h-5 w-5" />
                Gate Configurations
              </CardTitle>
              <CardDescription>
                Configure thresholds for blocking PRs on prompt regressions
              </CardDescription>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() =>
                queryClient.invalidateQueries({
                  queryKey: ["prompt-ci-gates"],
                })
              }
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Refresh
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {configsLoading ? (
            <div className="h-32 bg-muted animate-pulse rounded-lg" />
          ) : !gateConfigs?.length ? (
            <p className="text-sm text-muted-foreground text-center py-8">
              No gate configurations yet. Create one to start blocking PRs on
              regressions.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Block Severity</TableHead>
                  <TableHead>Confidence</TableHead>
                  <TableHead>Enabled</TableHead>
                  <TableHead>Updated</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {gateConfigs.map((config) => (
                  <TableRow key={config.id}>
                    <TableCell className="font-medium">{config.name}</TableCell>
                    <TableCell>
                      <Badge
                        variant={severityBadgeVariant(config.blockOnSeverity)}
                      >
                        {config.blockOnSeverity}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {(config.confidenceLevel * 100).toFixed(0)}%
                    </TableCell>
                    <TableCell>
                      <Switch
                        checked={config.enabled}
                        onCheckedChange={() =>
                          toggleGateMutation.mutate(config)
                        }
                      />
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(config.updatedAt).toLocaleDateString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Regression History Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5" />
            Regression History
          </CardTitle>
          <CardDescription>
            Recent regression events and gate evaluation results
          </CardDescription>
        </CardHeader>
        <CardContent>
          {historyLoading ? (
            <div className="h-32 bg-muted animate-pulse rounded-lg" />
          ) : !history?.length ? (
            <p className="text-sm text-muted-foreground text-center py-8">
              No regression events recorded yet.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Status</TableHead>
                  <TableHead>Branch</TableHead>
                  <TableHead>Commit</TableHead>
                  <TableHead>PR</TableHead>
                  <TableHead>Severity</TableHead>
                  <TableHead>Blocked</TableHead>
                  <TableHead>Date</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {history.map((entry) => (
                  <TableRow key={entry.id}>
                    <TableCell>
                      {entry.passed ? (
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                      ) : (
                        <XCircle className="h-4 w-4 text-red-500" />
                      )}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {entry.branch}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {entry.commitSha.substring(0, 7)}
                    </TableCell>
                    <TableCell>
                      {entry.prNumber ? `#${entry.prNumber}` : "—"}
                    </TableCell>
                    <TableCell>
                      <Badge variant={severityBadgeVariant(entry.severity)}>
                        {entry.severity}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {entry.blockedPr ? (
                        <Badge variant="destructive">Blocked</Badge>
                      ) : (
                        "—"
                      )}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(entry.createdAt).toLocaleDateString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[...Array(4)].map((_, i) => (
          <div
            key={i}
            className="h-32 rounded-lg border bg-card animate-pulse"
          />
        ))}
      </div>
      <div className="h-64 rounded-lg border bg-card animate-pulse" />
      <div className="h-64 rounded-lg border bg-card animate-pulse" />
    </div>
  );
}
