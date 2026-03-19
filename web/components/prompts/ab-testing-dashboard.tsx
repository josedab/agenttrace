"use client";

import { useState } from "react";
import {
  FlaskConical,
  Play,
  Pause,
  Square,
  Trophy,
  TrendingUp,
  BarChart3,
  Plus,
  ArrowUpRight,
  ArrowDownRight,
  Minus,
  Percent,
  Target,
  Zap,
  CheckCircle2,
  Clock,
  AlertCircle,
} from "lucide-react";
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface ABTest {
  id: string;
  name: string;
  description?: string;
  promptId: string;
  status: "draft" | "running" | "paused" | "completed" | "cancelled";
  variants: ABTestVariant[];
  targetMetric: string;
  minSampleSize: number;
  confidenceLevel: number;
  winnerId?: string;
  gradualRollout?: GradualRollout;
  startedAt?: string;
  endedAt?: string;
  createdAt: string;
}

interface ABTestVariant {
  id: string;
  name: string;
  promptVersionId: string;
  trafficPercent: number;
  isControl: boolean;
  sampleCount: number;
  metrics: {
    avgScore: number;
    stdDeviation: number;
    avgLatencyMs: number;
    avgCostUsd: number;
    errorRate: number;
    p95LatencyMs: number;
    totalTokens: number;
  };
}

interface GradualRollout {
  enabled: boolean;
  initialPercent: number;
  incrementPercent: number;
  incrementIntervalHours: number;
  currentPercent: number;
  autoComplete: boolean;
}

interface VariantStat {
  variantId: string;
  variantName: string;
  mean: number;
  stdDev: number;
  confidenceLower: number;
  confidenceUpper: number;
  sampleSize: number;
  isWinner: boolean;
  improvement: number;
}

interface ABTestStatistics {
  testId: string;
  isSignificant: boolean;
  pValue: number;
  confidenceLevel: number;
  effect: number;
  powerAnalysis: number;
  requiredSamples: number;
  currentSamples: number;
  variantStats: VariantStat[];
  recommendation: string;
}

const statusConfig: Record<string, { label: string; variant: "default" | "secondary" | "destructive" | "outline" }> = {
  draft: { label: "Draft", variant: "secondary" },
  running: { label: "Running", variant: "default" },
  paused: { label: "Paused", variant: "outline" },
  completed: { label: "Completed", variant: "default" },
  cancelled: { label: "Cancelled", variant: "destructive" },
};

// Demo data
const demoTests: ABTest[] = [
  {
    id: "t1",
    name: "System Prompt Optimization",
    description: "Testing concise vs detailed system prompts for code generation",
    promptId: "p1",
    status: "running",
    targetMetric: "accuracy",
    minSampleSize: 200,
    confidenceLevel: 0.95,
    variants: [
      {
        id: "v1", name: "Control (Detailed)", promptVersionId: "pv1",
        trafficPercent: 50, isControl: true, sampleCount: 156,
        metrics: { avgScore: 0.72, stdDeviation: 0.12, avgLatencyMs: 340, avgCostUsd: 0.0032, errorRate: 0.03, p95LatencyMs: 520, totalTokens: 45200 },
      },
      {
        id: "v2", name: "Concise Prompt", promptVersionId: "pv2",
        trafficPercent: 50, isControl: false, sampleCount: 148,
        metrics: { avgScore: 0.78, stdDeviation: 0.10, avgLatencyMs: 280, avgCostUsd: 0.0024, errorRate: 0.02, p95LatencyMs: 410, totalTokens: 38100 },
      },
    ],
    startedAt: "2024-01-15T10:00:00Z",
    createdAt: "2024-01-14T09:00:00Z",
  },
  {
    id: "t2",
    name: "Few-shot vs Zero-shot",
    description: "Comparing few-shot examples with zero-shot for classification",
    promptId: "p2",
    status: "completed",
    targetMetric: "accuracy",
    minSampleSize: 100,
    confidenceLevel: 0.95,
    winnerId: "v4",
    variants: [
      {
        id: "v3", name: "Zero-shot", promptVersionId: "pv3",
        trafficPercent: 50, isControl: true, sampleCount: 210,
        metrics: { avgScore: 0.65, stdDeviation: 0.15, avgLatencyMs: 200, avgCostUsd: 0.0018, errorRate: 0.05, p95LatencyMs: 380, totalTokens: 32000 },
      },
      {
        id: "v4", name: "3-shot Examples", promptVersionId: "pv4",
        trafficPercent: 50, isControl: false, sampleCount: 205,
        metrics: { avgScore: 0.82, stdDeviation: 0.09, avgLatencyMs: 360, avgCostUsd: 0.0041, errorRate: 0.01, p95LatencyMs: 550, totalTokens: 58000 },
      },
    ],
    startedAt: "2024-01-10T08:00:00Z",
    endedAt: "2024-01-13T18:00:00Z",
    createdAt: "2024-01-09T10:00:00Z",
  },
  {
    id: "t3",
    name: "Temperature Tuning",
    promptId: "p3",
    status: "draft",
    targetMetric: "latency",
    minSampleSize: 150,
    confidenceLevel: 0.90,
    variants: [
      {
        id: "v5", name: "Temp 0.3", promptVersionId: "pv5",
        trafficPercent: 33.33, isControl: true, sampleCount: 0,
        metrics: { avgScore: 0, stdDeviation: 0, avgLatencyMs: 0, avgCostUsd: 0, errorRate: 0, p95LatencyMs: 0, totalTokens: 0 },
      },
      {
        id: "v6", name: "Temp 0.7", promptVersionId: "pv6",
        trafficPercent: 33.33, isControl: false, sampleCount: 0,
        metrics: { avgScore: 0, stdDeviation: 0, avgLatencyMs: 0, avgCostUsd: 0, errorRate: 0, p95LatencyMs: 0, totalTokens: 0 },
      },
      {
        id: "v7", name: "Temp 1.0", promptVersionId: "pv7",
        trafficPercent: 33.34, isControl: false, sampleCount: 0,
        metrics: { avgScore: 0, stdDeviation: 0, avgLatencyMs: 0, avgCostUsd: 0, errorRate: 0, p95LatencyMs: 0, totalTokens: 0 },
      },
    ],
    createdAt: "2024-01-16T14:00:00Z",
  },
];

const demoStats: ABTestStatistics = {
  testId: "t1",
  isSignificant: true,
  pValue: 0.023,
  confidenceLevel: 0.95,
  effect: 0.52,
  powerAnalysis: 0.81,
  requiredSamples: 280,
  currentSamples: 304,
  recommendation: "select_winner",
  variantStats: [
    { variantId: "v1", variantName: "Control (Detailed)", mean: 0.72, stdDev: 0.12, confidenceLower: 0.70, confidenceUpper: 0.74, sampleSize: 156, isWinner: false, improvement: 0 },
    { variantId: "v2", variantName: "Concise Prompt", mean: 0.78, stdDev: 0.10, confidenceLower: 0.76, confidenceUpper: 0.80, sampleSize: 148, isWinner: true, improvement: 8.33 },
  ],
};

export function ABTestingDashboard() {
  const [tests] = useState<ABTest[]>(demoTests);
  const [selectedTest, setSelectedTest] = useState<ABTest | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);

  const runningTests = tests.filter((t) => t.status === "running").length;
  const completedTests = tests.filter((t) => t.status === "completed").length;
  const totalSamples = tests.reduce(
    (sum, t) => sum + t.variants.reduce((vs, v) => vs + v.sampleCount, 0),
    0
  );

  return (
    <div className="space-y-6">
      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Tests</CardTitle>
            <FlaskConical className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{runningTests}</div>
            <p className="text-xs text-muted-foreground">{tests.length} total tests</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Completed</CardTitle>
            <CheckCircle2 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{completedTests}</div>
            <p className="text-xs text-muted-foreground">with winners selected</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Samples</CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalSamples.toLocaleString()}</div>
            <p className="text-xs text-muted-foreground">across all variants</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg Improvement</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">+8.3%</div>
            <p className="text-xs text-muted-foreground">winning variants vs control</p>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="tests" className="space-y-4">
        <TabsList>
          <TabsTrigger value="tests">All Tests</TabsTrigger>
          <TabsTrigger value="details">Test Details</TabsTrigger>
        </TabsList>

        {/* Tests List */}
        <TabsContent value="tests" className="space-y-4">
          <div className="flex justify-between items-center">
            <h3 className="text-lg font-semibold">A/B Tests</h3>
            <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
              <DialogTrigger asChild>
                <Button size="sm">
                  <Plus className="h-4 w-4 mr-2" /> New Test
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-lg">
                <DialogHeader>
                  <DialogTitle>Create A/B Test</DialogTitle>
                  <DialogDescription>Set up a new prompt A/B test experiment</DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div className="space-y-2">
                    <Label>Test Name</Label>
                    <Input placeholder="e.g., System Prompt Optimization" />
                  </div>
                  <div className="space-y-2">
                    <Label>Target Metric</Label>
                    <Select>
                      <SelectTrigger>
                        <SelectValue placeholder="Select metric" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="accuracy">Accuracy</SelectItem>
                        <SelectItem value="latency">Latency</SelectItem>
                        <SelectItem value="cost">Cost</SelectItem>
                        <SelectItem value="custom">Custom</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>Min Sample Size</Label>
                      <Input type="number" defaultValue={100} />
                    </div>
                    <div className="space-y-2">
                      <Label>Confidence Level</Label>
                      <Select defaultValue="0.95">
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="0.90">90%</SelectItem>
                          <SelectItem value="0.95">95%</SelectItem>
                          <SelectItem value="0.99">99%</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                  <Button className="w-full" onClick={() => setShowCreateDialog(false)}>
                    Create Test
                  </Button>
                </div>
              </DialogContent>
            </Dialog>
          </div>

          <Card>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Test Name</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Variants</TableHead>
                  <TableHead>Samples</TableHead>
                  <TableHead>Target</TableHead>
                  <TableHead>Winner</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tests.map((test) => {
                  const totalSamples = test.variants.reduce((s, v) => s + v.sampleCount, 0);
                  const config = statusConfig[test.status];
                  return (
                    <TableRow key={test.id} className="cursor-pointer" onClick={() => setSelectedTest(test)}>
                      <TableCell>
                        <div>
                          <div className="font-medium">{test.name}</div>
                          {test.description && (
                            <div className="text-xs text-muted-foreground truncate max-w-[300px]">
                              {test.description}
                            </div>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={config.variant}>{config.label}</Badge>
                      </TableCell>
                      <TableCell>{test.variants.length}</TableCell>
                      <TableCell>
                        {totalSamples} / {test.minSampleSize}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{test.targetMetric}</Badge>
                      </TableCell>
                      <TableCell>
                        {test.winnerId ? (
                          <div className="flex items-center gap-1 text-green-600">
                            <Trophy className="h-3 w-3" />
                            <span className="text-xs">
                              {test.variants.find((v) => v.id === test.winnerId)?.name}
                            </span>
                          </div>
                        ) : (
                          <span className="text-xs text-muted-foreground">—</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          {test.status === "draft" && (
                            <Button size="sm" variant="ghost" title="Start">
                              <Play className="h-3 w-3" />
                            </Button>
                          )}
                          {test.status === "running" && (
                            <>
                              <Button size="sm" variant="ghost" title="Pause">
                                <Pause className="h-3 w-3" />
                              </Button>
                              <Button size="sm" variant="ghost" title="Stop">
                                <Square className="h-3 w-3" />
                              </Button>
                            </>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </Card>
        </TabsContent>

        {/* Test Details */}
        <TabsContent value="details" className="space-y-4">
          {selectedTest ? (
            <>
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-semibold">{selectedTest.name}</h3>
                  <p className="text-sm text-muted-foreground">{selectedTest.description}</p>
                </div>
                <Badge variant={statusConfig[selectedTest.status].variant}>
                  {statusConfig[selectedTest.status].label}
                </Badge>
              </div>

              {/* Variant Performance Comparison */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Variant Performance</CardTitle>
                  <CardDescription>Real-time comparison across all variants</CardDescription>
                </CardHeader>
                <CardContent>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Variant</TableHead>
                        <TableHead>Traffic</TableHead>
                        <TableHead>Samples</TableHead>
                        <TableHead>Avg Score</TableHead>
                        <TableHead>Avg Latency</TableHead>
                        <TableHead>Cost/Call</TableHead>
                        <TableHead>Error Rate</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {selectedTest.variants.map((v) => (
                        <TableRow key={v.id}>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <span className="font-medium">{v.name}</span>
                              {v.isControl && <Badge variant="outline" className="text-xs">Control</Badge>}
                              {selectedTest.winnerId === v.id && (
                                <Trophy className="h-3 w-3 text-yellow-500" />
                              )}
                            </div>
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-1">
                              <Percent className="h-3 w-3" />
                              {v.trafficPercent}%
                            </div>
                          </TableCell>
                          <TableCell>{v.sampleCount}</TableCell>
                          <TableCell>
                            <span className="font-mono">{v.metrics.avgScore.toFixed(3)}</span>
                          </TableCell>
                          <TableCell>
                            <span className="font-mono">{v.metrics.avgLatencyMs.toFixed(0)}ms</span>
                          </TableCell>
                          <TableCell>
                            <span className="font-mono">${v.metrics.avgCostUsd.toFixed(4)}</span>
                          </TableCell>
                          <TableCell>
                            <span className="font-mono">{(v.metrics.errorRate * 100).toFixed(1)}%</span>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>

              {/* Statistical Significance */}
              {selectedTest.status === "running" && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Statistical Analysis</CardTitle>
                    <CardDescription>
                      Confidence intervals and significance testing
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-6">
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                      <div className="text-center p-3 rounded-lg bg-muted/50">
                        <div className="text-xs text-muted-foreground mb-1">P-Value</div>
                        <div className="text-lg font-bold font-mono">
                          {demoStats.pValue.toFixed(4)}
                        </div>
                        <Badge variant={demoStats.isSignificant ? "default" : "secondary"} className="mt-1">
                          {demoStats.isSignificant ? "Significant" : "Not Significant"}
                        </Badge>
                      </div>
                      <div className="text-center p-3 rounded-lg bg-muted/50">
                        <div className="text-xs text-muted-foreground mb-1">Effect Size</div>
                        <div className="text-lg font-bold font-mono">{demoStats.effect.toFixed(3)}</div>
                        <div className="text-xs text-muted-foreground mt-1">Cohen&apos;s d</div>
                      </div>
                      <div className="text-center p-3 rounded-lg bg-muted/50">
                        <div className="text-xs text-muted-foreground mb-1">Power</div>
                        <div className="text-lg font-bold font-mono">
                          {(demoStats.powerAnalysis * 100).toFixed(1)}%
                        </div>
                        <div className="text-xs text-muted-foreground mt-1">
                          {demoStats.powerAnalysis >= 0.8 ? "Adequate" : "Low"}
                        </div>
                      </div>
                      <div className="text-center p-3 rounded-lg bg-muted/50">
                        <div className="text-xs text-muted-foreground mb-1">Samples</div>
                        <div className="text-lg font-bold font-mono">
                          {demoStats.currentSamples}/{demoStats.requiredSamples}
                        </div>
                        <Progress
                          value={(demoStats.currentSamples / demoStats.requiredSamples) * 100}
                          className="mt-2 h-1"
                        />
                      </div>
                    </div>

                    {/* Confidence Intervals Visualization */}
                    <div className="space-y-3">
                      <h4 className="text-sm font-medium">Confidence Intervals ({(demoStats.confidenceLevel * 100)}%)</h4>
                      {demoStats.variantStats.map((vs) => {
                        const range = vs.confidenceUpper - vs.confidenceLower;
                        const minVal = Math.min(...demoStats.variantStats.map((s) => s.confidenceLower)) - 0.02;
                        const maxVal = Math.max(...demoStats.variantStats.map((s) => s.confidenceUpper)) + 0.02;
                        const totalRange = maxVal - minVal;
                        const leftPct = ((vs.confidenceLower - minVal) / totalRange) * 100;
                        const widthPct = (range / totalRange) * 100;
                        const meanPct = ((vs.mean - minVal) / totalRange) * 100;

                        return (
                          <div key={vs.variantId} className="space-y-1">
                            <div className="flex justify-between text-xs">
                              <span className="flex items-center gap-1">
                                {vs.variantName}
                                {vs.isWinner && <Trophy className="h-3 w-3 text-yellow-500" />}
                              </span>
                              <span className="font-mono text-muted-foreground">
                                {vs.mean.toFixed(3)} [{vs.confidenceLower.toFixed(3)}, {vs.confidenceUpper.toFixed(3)}]
                              </span>
                            </div>
                            <div className="relative h-6 bg-muted rounded">
                              <div
                                className={`absolute h-full rounded ${vs.isWinner ? "bg-green-200 dark:bg-green-900" : "bg-blue-200 dark:bg-blue-900"}`}
                                style={{ left: `${leftPct}%`, width: `${widthPct}%` }}
                              />
                              <div
                                className={`absolute top-0 h-full w-0.5 ${vs.isWinner ? "bg-green-600" : "bg-blue-600"}`}
                                style={{ left: `${meanPct}%` }}
                              />
                            </div>
                          </div>
                        );
                      })}
                    </div>

                    {/* Recommendation */}
                    <div className="flex items-center gap-3 p-4 rounded-lg border">
                      {demoStats.recommendation === "select_winner" ? (
                        <>
                          <CheckCircle2 className="h-5 w-5 text-green-500" />
                          <div>
                            <div className="font-medium text-sm">Ready to Select Winner</div>
                            <div className="text-xs text-muted-foreground">
                              Statistical significance reached. &quot;Concise Prompt&quot; shows +8.3% improvement.
                            </div>
                          </div>
                          <Button size="sm" className="ml-auto">
                            <Trophy className="h-3 w-3 mr-1" /> Select Winner
                          </Button>
                        </>
                      ) : demoStats.recommendation === "continue" ? (
                        <>
                          <Clock className="h-5 w-5 text-yellow-500" />
                          <div>
                            <div className="font-medium text-sm">Continue Collecting Data</div>
                            <div className="text-xs text-muted-foreground">
                              More samples needed for statistical significance.
                            </div>
                          </div>
                        </>
                      ) : (
                        <>
                          <AlertCircle className="h-5 w-5 text-orange-500" />
                          <div>
                            <div className="font-medium text-sm">Insufficient Data</div>
                            <div className="text-xs text-muted-foreground">
                              Minimum sample size not yet reached.
                            </div>
                          </div>
                        </>
                      )}
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Traffic Split Visualization */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Traffic Distribution</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="flex h-8 rounded-lg overflow-hidden">
                    {selectedTest.variants.map((v, i) => {
                      const colors = [
                        "bg-blue-500",
                        "bg-green-500",
                        "bg-purple-500",
                        "bg-orange-500",
                      ];
                      return (
                        <div
                          key={v.id}
                          className={`${colors[i % colors.length]} flex items-center justify-center text-white text-xs font-medium`}
                          style={{ width: `${v.trafficPercent}%` }}
                        >
                          {v.name} ({v.trafficPercent}%)
                        </div>
                      );
                    })}
                  </div>
                </CardContent>
              </Card>

              {/* Gradual Rollout */}
              {selectedTest.gradualRollout && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Gradual Rollout</CardTitle>
                    <CardDescription>Progressive traffic shift to winning variant</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Rollout Progress</span>
                      <span className="text-sm font-medium">
                        {selectedTest.gradualRollout.currentPercent}%
                      </span>
                    </div>
                    <Progress value={selectedTest.gradualRollout.currentPercent} />
                    <div className="grid grid-cols-3 gap-4 text-center text-xs text-muted-foreground">
                      <div>
                        <Zap className="h-3 w-3 mx-auto mb-1" />
                        Initial: {selectedTest.gradualRollout.initialPercent}%
                      </div>
                      <div>
                        <ArrowUpRight className="h-3 w-3 mx-auto mb-1" />
                        +{selectedTest.gradualRollout.incrementPercent}% every{" "}
                        {selectedTest.gradualRollout.incrementIntervalHours}h
                      </div>
                      <div>
                        <Target className="h-3 w-3 mx-auto mb-1" />
                        Auto-complete: {selectedTest.gradualRollout.autoComplete ? "Yes" : "No"}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Winner Selection Controls */}
              {selectedTest.status === "running" && !selectedTest.winnerId && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Winner Selection</CardTitle>
                    <CardDescription>Select the winning variant to deploy</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="flex gap-2">
                      {selectedTest.variants.map((v) => (
                        <Button key={v.id} variant="outline" size="sm">
                          <Trophy className="h-3 w-3 mr-1" />
                          Select &quot;{v.name}&quot;
                        </Button>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              )}
            </>
          ) : (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <FlaskConical className="h-8 w-8 mb-2" />
                <p>Select a test from the list to view details</p>
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
