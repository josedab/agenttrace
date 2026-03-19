"use client";

import * as React from "react";
import {
  Globe,
  Shield,
  TrendingUp,
  TrendingDown,
  Minus,
  Activity,
  Lock,
  BarChart3,
  Lightbulb,
  Search,
  Info,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

// Types
interface MeshStatus {
  totalInstances: number;
  activeInstances: number;
  totalMetrics: number;
  avgPrivacyLevel: string;
  meshHealth: string;
  lastSync?: string;
}

interface PrivacyBudget {
  instanceId: string;
  totalEpsilon: number;
  usedEpsilon: number;
  remainingEpsilon: number;
  queriesCount: number;
  resetAt: string;
}

interface CrossOrgComparison {
  metricName: string;
  yourValue: number;
  industryMedian: number;
  industryP25: number;
  industryP75: number;
  industryP90: number;
  percentile: number;
  trend: string;
  participantCount: number;
}

interface FederatedInsight {
  id: string;
  category: string;
  title: string;
  description: string;
  impact: string;
  recommendation: string;
  benchmark: number;
  yourValue: number;
  percentile: number;
  createdAt: string;
}

interface IndustryBaseline {
  metricType: string;
  p10: number;
  p25: number;
  p50: number;
  p75: number;
  p90: number;
  mean: number;
  stdDev: number;
  participants: number;
  lastUpdated: string;
}

interface FederatedAnalyticsDashboard {
  meshStatus: MeshStatus;
  privacyBudget: PrivacyBudget;
  comparisons: CrossOrgComparison[];
  insights: FederatedInsight[];
  baselines: IndustryBaseline[];
}

// Trend icon helper
function TrendIcon({ trend }: { trend: string }) {
  if (trend === "improving") return <TrendingUp className="h-4 w-4 text-green-500" />;
  if (trend === "declining") return <TrendingDown className="h-4 w-4 text-red-500" />;
  return <Minus className="h-4 w-4 text-muted-foreground" />;
}

// Impact badge helper
function ImpactBadge({ impact }: { impact: string }) {
  const variants: Record<string, string> = {
    high: "bg-red-500 text-white",
    medium: "bg-yellow-500 text-white",
    low: "bg-green-500 text-white",
  };

  return (
    <Badge className={cn("text-xs", variants[impact] || "bg-gray-500 text-white")}>
      {impact}
    </Badge>
  );
}

// Mesh Status Overview
function MeshStatusOverview({ status }: { status: MeshStatus }) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">Active Instances</CardTitle>
          <Globe className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{status.activeInstances}/{status.totalInstances}</div>
          <p className="text-xs text-muted-foreground mt-1">Federated mesh participants</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">Total Metrics</CardTitle>
          <BarChart3 className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{status.totalMetrics.toLocaleString()}</div>
          <p className="text-xs text-muted-foreground mt-1">Aggregated across mesh</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">Mesh Health</CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold capitalize">{status.meshHealth}</div>
          <Badge variant={status.meshHealth === "healthy" ? "default" : "destructive"} className="text-xs mt-1">
            {status.avgPrivacyLevel.replace("_", " ")}
          </Badge>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">Last Sync</CardTitle>
          <Shield className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {status.lastSync ? new Date(status.lastSync).toLocaleTimeString() : "—"}
          </div>
          <p className="text-xs text-muted-foreground mt-1">Privacy-preserving sync</p>
        </CardContent>
      </Card>
    </div>
  );
}

// Privacy Budget Gauge
function PrivacyBudgetGauge({ budget }: { budget: PrivacyBudget }) {
  const usedPct = (budget.usedEpsilon / budget.totalEpsilon) * 100;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Lock className="h-5 w-5" />
          Privacy Budget (ε)
        </CardTitle>
        <CardDescription>
          Differential privacy budget tracks cumulative information disclosure
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span>Used: {budget.usedEpsilon.toFixed(1)}ε</span>
              <span>Remaining: {budget.remainingEpsilon.toFixed(1)}ε</span>
            </div>
            <div className="h-4 bg-muted rounded-full overflow-hidden">
              <div
                className={cn(
                  "h-full rounded-full transition-all",
                  usedPct < 50 ? "bg-green-500" : usedPct < 80 ? "bg-yellow-500" : "bg-red-500"
                )}
                style={{ width: `${usedPct}%` }}
              />
            </div>
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <span>Total budget: {budget.totalEpsilon.toFixed(1)}ε</span>
              <span>{budget.queriesCount} queries executed</span>
            </div>
          </div>
          <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted p-2 rounded">
            <Info className="h-3 w-3 flex-shrink-0" />
            <span>Budget resets {new Date(budget.resetAt).toLocaleDateString()}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

// Cross-Org Comparison Table
function ComparisonTable({ comparisons }: { comparisons: CrossOrgComparison[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Cross-Organization Comparison</CardTitle>
        <CardDescription>Anonymous benchmarking against industry participants</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Metric</TableHead>
                <TableHead>Your Value</TableHead>
                <TableHead>Industry Median</TableHead>
                <TableHead>Percentile</TableHead>
                <TableHead>Trend</TableHead>
                <TableHead>Participants</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {comparisons.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                    No comparison data available
                  </TableCell>
                </TableRow>
              ) : (
                comparisons.map((c) => (
                  <TableRow key={c.metricName}>
                    <TableCell>
                      <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{c.metricName}</code>
                    </TableCell>
                    <TableCell className="font-medium">{c.yourValue.toFixed(3)}</TableCell>
                    <TableCell>{c.industryMedian.toFixed(3)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div className="w-24 h-2 bg-muted rounded-full overflow-hidden">
                          <div
                            className={cn(
                              "h-full rounded-full",
                              c.percentile < 30 ? "bg-red-500" : c.percentile < 70 ? "bg-yellow-500" : "bg-green-500"
                            )}
                            style={{ width: `${c.percentile}%` }}
                          />
                        </div>
                        <span className="text-sm">{c.percentile}th</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <TrendIcon trend={c.trend} />
                        <span className="text-sm capitalize">{c.trend}</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-sm">{c.participantCount}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}

// Industry Baseline Cards
function BaselineCards({ baselines }: { baselines: IndustryBaseline[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Industry Baselines</CardTitle>
        <CardDescription>Aggregated performance baselines across {baselines[0]?.participants || 0} participants</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid gap-4 md:grid-cols-2">
          {baselines.map((b) => (
            <div key={b.metricType} className="border rounded-lg p-4 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium capitalize">{b.metricType.replace("_", " ")}</span>
                <Badge variant="outline" className="text-xs">{b.participants} orgs</Badge>
              </div>
              <div className="grid grid-cols-5 gap-1 text-xs text-center">
                <div>
                  <div className="text-muted-foreground">P10</div>
                  <div className="font-medium">{b.p10.toFixed(2)}</div>
                </div>
                <div>
                  <div className="text-muted-foreground">P25</div>
                  <div className="font-medium">{b.p25.toFixed(2)}</div>
                </div>
                <div className="bg-muted rounded px-1">
                  <div className="text-muted-foreground">P50</div>
                  <div className="font-bold">{b.p50.toFixed(2)}</div>
                </div>
                <div>
                  <div className="text-muted-foreground">P75</div>
                  <div className="font-medium">{b.p75.toFixed(2)}</div>
                </div>
                <div>
                  <div className="text-muted-foreground">P90</div>
                  <div className="font-medium">{b.p90.toFixed(2)}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

// Actionable Insights List
function InsightsList({ insights }: { insights: FederatedInsight[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Lightbulb className="h-5 w-5" />
          Actionable Insights
        </CardTitle>
        <CardDescription>Recommendations based on cross-organization analytics</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {insights.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-4">
              All metrics are within normal range — no insights to report
            </p>
          ) : (
            insights.map((insight) => (
              <div key={insight.id} className="border rounded-lg p-4 space-y-2">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-medium">{insight.title}</h4>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="text-xs capitalize">{insight.category}</Badge>
                    <ImpactBadge impact={insight.impact} />
                  </div>
                </div>
                <p className="text-xs text-muted-foreground">{insight.description}</p>
                <div className="text-xs bg-muted p-2 rounded">
                  <span className="font-medium">Recommendation:</span> {insight.recommendation}
                </div>
                <div className="flex items-center gap-4 text-xs text-muted-foreground">
                  <span>Your value: {insight.yourValue.toFixed(3)}</span>
                  <span>Benchmark: {insight.benchmark.toFixed(3)}</span>
                  <span>Percentile: {insight.percentile}th</span>
                </div>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}

// Query Builder
function QueryBuilder() {
  const [metrics, setMetrics] = React.useState<string[]>([]);
  const [results, setResults] = React.useState<CrossOrgComparison[] | null>(null);
  const [loading, setLoading] = React.useState(false);

  const availableMetrics = ["avg_latency_ms", "avg_cost_per_trace", "error_rate", "avg_quality_score"];

  const toggleMetric = (metric: string) => {
    setMetrics((prev) =>
      prev.includes(metric) ? prev.filter((m) => m !== metric) : [...prev, metric]
    );
  };

  const runQuery = async () => {
    if (metrics.length === 0) return;
    setLoading(true);
    try {
      const res = await fetch("/api/public/federated/query", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ metrics }),
      });
      if (res.ok) {
        const data = await res.json();
        setResults(data.comparisons || []);
      }
    } catch {
      // Silently handle fetch errors
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Search className="h-5 w-5" />
          Privacy-Preserving Query
        </CardTitle>
        <CardDescription>Run queries against the federated mesh with differential privacy guarantees</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="flex flex-wrap gap-2">
            {availableMetrics.map((metric) => (
              <Badge
                key={metric}
                variant={metrics.includes(metric) ? "default" : "outline"}
                className="cursor-pointer"
                onClick={() => toggleMetric(metric)}
              >
                {metric}
              </Badge>
            ))}
          </div>
          <Button onClick={runQuery} disabled={metrics.length === 0 || loading} size="sm">
            {loading ? "Querying..." : "Run Query"}
          </Button>

          {results && results.length > 0 && (
            <div className="border rounded-lg mt-4">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Metric</TableHead>
                    <TableHead>Your Value</TableHead>
                    <TableHead>Industry Median</TableHead>
                    <TableHead>Percentile</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {results.map((r) => (
                    <TableRow key={r.metricName}>
                      <TableCell><code className="text-xs bg-muted px-1.5 py-0.5 rounded">{r.metricName}</code></TableCell>
                      <TableCell>{r.yourValue.toFixed(3)}</TableCell>
                      <TableCell>{r.industryMedian.toFixed(3)}</TableCell>
                      <TableCell>{r.percentile}th</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

// Differential Privacy Explainer
function PrivacyExplainer() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          How Differential Privacy Works
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3 text-sm text-muted-foreground">
          <div className="flex items-start gap-3">
            <div className="h-6 w-6 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 mt-0.5">
              <span className="text-xs font-bold text-primary">1</span>
            </div>
            <p>Your raw trace data never leaves your organization. Only aggregated, noise-injected metrics are shared.</p>
          </div>
          <div className="flex items-start gap-3">
            <div className="h-6 w-6 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 mt-0.5">
              <span className="text-xs font-bold text-primary">2</span>
            </div>
            <p>Laplacian noise is added to each metric before submission, calibrated by your privacy budget (ε).</p>
          </div>
          <div className="flex items-start gap-3">
            <div className="h-6 w-6 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 mt-0.5">
              <span className="text-xs font-bold text-primary">3</span>
            </div>
            <p>A lower epsilon means stronger privacy but less accurate benchmarks. The default ε=1.0 provides strong privacy guarantees.</p>
          </div>
          <div className="flex items-start gap-3">
            <div className="h-6 w-6 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 mt-0.5">
              <span className="text-xs font-bold text-primary">4</span>
            </div>
            <p>Your privacy budget resets monthly. Each query consumes a small amount of budget to limit cumulative disclosure.</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

// Main Dashboard
export function FederatedAnalyticsDashboardComponent() {
  const [dashboard, setDashboard] = React.useState<FederatedAnalyticsDashboard | null>(null);

  React.useEffect(() => {
    fetch("/api/public/federated/dashboard")
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (data) setDashboard(data);
      })
      .catch(() => {
        // Silently handle fetch errors in development
      });
  }, []);

  const defaultMeshStatus: MeshStatus = {
    totalInstances: 0,
    activeInstances: 0,
    totalMetrics: 0,
    avgPrivacyLevel: "differential_privacy",
    meshHealth: "healthy",
  };

  const defaultBudget: PrivacyBudget = {
    instanceId: "",
    totalEpsilon: 10,
    usedEpsilon: 0,
    remainingEpsilon: 10,
    queriesCount: 0,
    resetAt: new Date().toISOString(),
  };

  return (
    <div className="space-y-6">
      <MeshStatusOverview status={dashboard?.meshStatus || defaultMeshStatus} />

      <div className="grid gap-6 lg:grid-cols-2">
        <PrivacyBudgetGauge budget={dashboard?.privacyBudget || defaultBudget} />
        <PrivacyExplainer />
      </div>

      <ComparisonTable comparisons={dashboard?.comparisons || []} />

      <div className="grid gap-6 lg:grid-cols-2">
        <InsightsList insights={dashboard?.insights || []} />
        <BaselineCards baselines={dashboard?.baselines || []} />
      </div>

      <QueryBuilder />
    </div>
  );
}
