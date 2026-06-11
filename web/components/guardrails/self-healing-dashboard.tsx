"use client";

import * as React from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Badge, type BadgeProps } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import {
  Shield,
  Activity,
  RefreshCw,
  Zap,
  CheckCircle,
  XCircle,
  ArrowRight,
  Play,
} from "lucide-react";

interface DashboardStats {
  totalPolicies: number;
  activePolicies: number;
  totalTriggers: number;
  remediationRate: number;
  circuitBreakers: number;
  openCircuits: number;
  avgRemediationMs: number;
  blockedRequests: number;
}

interface SelfHealingPolicy {
  id: string;
  name: string;
  ruleId: string;
  enabled: boolean;
  remediationAction: {
    type: string;
    maxRetries?: number;
    parameters?: Record<string, string>;
  };
  circuitBreaker?: {
    failureThreshold: number;
    successThreshold: number;
    timeoutSeconds: number;
    halfOpenRequests: number;
    state: string;
    failureCount: number;
    successCount: number;
  };
  retryPolicy?: {
    maxAttempts: number;
    initialDelayMs: number;
    maxDelayMs: number;
    backoffMultiplier: number;
    retryableErrors?: string[];
  };
  fallbackChain?: {
    order: number;
    type: string;
    description: string;
    config: Record<string, string>;
    fallbackModel?: string;
  }[];
  triggerCount: number;
  lastTriggered?: string;
  createdAt: string;
}

interface PipelineResult {
  traceId: string;
  passed: boolean;
  evaluations: {
    ruleId: string;
    ruleName: string;
    ruleType: string;
    passed: boolean;
    violationMessage?: string;
    remediated: boolean;
    remediationAction?: string;
    latencyMs: number;
  }[];
  totalLatencyMs: number;
  remediated: boolean;
  blockedReason?: string;
}

interface EvaluatePipelineInput {
  traceId: string;
  output: string;
  costUsd: number;
  latencyMs: number;
}

const actionTypeColors: Record<string, NonNullable<BadgeProps["variant"]>> = {
  retry: "default",
  fallback: "secondary",
  circuit_break: "destructive",
  alert: "default",
  block: "destructive",
  transform: "outline",
};

const circuitStateColors: Record<string, string> = {
  closed: "text-green-600 bg-green-50",
  open: "text-red-600 bg-red-50",
  half_open: "text-yellow-600 bg-yellow-50",
};

export function SelfHealingDashboard() {
  const [activeTab, setActiveTab] = React.useState("overview");
  const [evalInput, setEvalInput] = React.useState({
    traceId: "",
    output: "",
    costUsd: 0,
    latencyMs: 0,
  });
  const [evalResult, setEvalResult] = React.useState<PipelineResult | null>(null);

  const { data: stats } = useQuery<DashboardStats>({
    queryKey: ["guardrails-dashboard-stats"],
    queryFn: () => api.guardrails.getDashboardStats() as Promise<DashboardStats>,
  });

  const { data: policiesData } = useQuery<{ policies: SelfHealingPolicy[] }>({
    queryKey: ["self-healing-policies"],
    queryFn: () => api.guardrails.listPolicies() as Promise<{ policies: SelfHealingPolicy[] }>,
  });

  const evaluateMutation = useMutation({
    mutationFn: (data: EvaluatePipelineInput) =>
      api.guardrails.evaluatePipeline(data) as Promise<PipelineResult>,
    onSuccess: (result: PipelineResult) => {
      setEvalResult(result);
    },
  });

  const policies = policiesData?.policies || [];

  const handleEvaluate = () => {
    if (!evalInput.traceId) return;
    evaluateMutation.mutate(evalInput);
  };

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center gap-2">
              <Shield className="h-4 w-4 text-muted-foreground" />
              <span className="text-xs text-muted-foreground">Total Policies</span>
            </div>
            <div className="text-2xl font-bold mt-1">{stats?.totalPolicies ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-green-600" />
              <span className="text-xs text-muted-foreground">Active</span>
            </div>
            <div className="text-2xl font-bold mt-1">{stats?.activePolicies ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center gap-2">
              <Zap className="h-4 w-4 text-yellow-600" />
              <span className="text-xs text-muted-foreground">Total Triggers</span>
            </div>
            <div className="text-2xl font-bold mt-1">{stats?.totalTriggers ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center gap-2">
              <RefreshCw className="h-4 w-4 text-blue-600" />
              <span className="text-xs text-muted-foreground">Remediation Rate</span>
            </div>
            <div className="text-2xl font-bold mt-1">
              {stats?.remediationRate != null ? `${(stats.remediationRate * 100).toFixed(1)}%` : "0%"}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Secondary Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-4">
            <div className="text-xs text-muted-foreground">Circuit Breakers</div>
            <div className="text-lg font-semibold">{stats?.circuitBreakers ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4">
            <div className="text-xs text-muted-foreground">Open Circuits</div>
            <div className="text-lg font-semibold text-red-600">{stats?.openCircuits ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4">
            <div className="text-xs text-muted-foreground">Avg Remediation</div>
            <div className="text-lg font-semibold">
              {stats?.avgRemediationMs != null ? `${stats.avgRemediationMs.toFixed(0)}ms` : "—"}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4">
            <div className="text-xs text-muted-foreground">Blocked Requests</div>
            <div className="text-lg font-semibold text-destructive">{stats?.blockedRequests ?? 0}</div>
          </CardContent>
        </Card>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="overview">Policies</TabsTrigger>
          <TabsTrigger value="evaluate">Pipeline Tester</TabsTrigger>
        </TabsList>

        {/* Policies Tab */}
        <TabsContent value="overview" className="space-y-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle className="text-base">Self-Healing Policies</CardTitle>
                <CardDescription>Automatic remediation policies linked to guardrail rules</CardDescription>
              </div>
            </CardHeader>
            <CardContent>
              {policies.length === 0 ? (
                <div className="text-center py-12">
                  <Shield className="h-12 w-12 mx-auto text-muted-foreground mb-3" />
                  <p className="text-sm text-muted-foreground mb-4">
                    No self-healing policies configured yet.
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Create policies to automatically retry, fallback, or circuit-break when agents violate guardrails.
                  </p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Policy</TableHead>
                      <TableHead>Action Type</TableHead>
                      <TableHead>Circuit Breaker</TableHead>
                      <TableHead>Triggers</TableHead>
                      <TableHead>Last Triggered</TableHead>
                      <TableHead className="w-[80px]">Enabled</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {policies.map((policy) => (
                      <TableRow key={policy.id}>
                        <TableCell>
                          <div className="font-medium">{policy.name}</div>
                          <div className="text-xs text-muted-foreground font-mono">
                            Rule: {policy.ruleId.slice(0, 8)}...
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant={actionTypeColors[policy.remediationAction.type] || "outline"}>
                            {policy.remediationAction.type}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {policy.circuitBreaker ? (
                            <span className={`text-xs px-2 py-1 rounded-full ${circuitStateColors[policy.circuitBreaker.state] || ""}`}>
                              {policy.circuitBreaker.state}
                            </span>
                          ) : (
                            <span className="text-xs text-muted-foreground">—</span>
                          )}
                        </TableCell>
                        <TableCell>{policy.triggerCount}</TableCell>
                        <TableCell>
                          {policy.lastTriggered
                            ? new Date(policy.lastTriggered).toLocaleString()
                            : "Never"}
                        </TableCell>
                        <TableCell>
                          <Switch checked={policy.enabled} />
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>

          {/* Fallback Chain & Retry Config for existing policies */}
          {policies.filter((p) => p.fallbackChain && p.fallbackChain.length > 0).map((policy) => (
            <Card key={`fallback-${policy.id}`}>
              <CardHeader>
                <CardTitle className="text-sm">Fallback Chain — {policy.name}</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-2 flex-wrap">
                  {policy.fallbackChain!.map((step, idx) => (
                    <React.Fragment key={idx}>
                      <div className="border rounded-md p-3 text-center min-w-[120px]">
                        <div className="text-xs font-medium">{step.type.replace("_", " ")}</div>
                        <div className="text-xs text-muted-foreground mt-1">{step.description}</div>
                        {step.fallbackModel && (
                          <Badge variant="secondary" className="mt-1 text-xs">{step.fallbackModel}</Badge>
                        )}
                      </div>
                      {idx < policy.fallbackChain!.length - 1 && (
                        <ArrowRight className="h-4 w-4 text-muted-foreground" />
                      )}
                    </React.Fragment>
                  ))}
                </div>
              </CardContent>
            </Card>
          ))}

          {policies.filter((p) => p.retryPolicy).map((policy) => (
            <Card key={`retry-${policy.id}`}>
              <CardHeader>
                <CardTitle className="text-sm">Retry Policy — {policy.name}</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-4 gap-4 text-sm">
                  <div>
                    <div className="text-xs text-muted-foreground">Max Attempts</div>
                    <div className="font-medium">{policy.retryPolicy!.maxAttempts}</div>
                  </div>
                  <div>
                    <div className="text-xs text-muted-foreground">Initial Delay</div>
                    <div className="font-medium">{policy.retryPolicy!.initialDelayMs}ms</div>
                  </div>
                  <div>
                    <div className="text-xs text-muted-foreground">Max Delay</div>
                    <div className="font-medium">{policy.retryPolicy!.maxDelayMs}ms</div>
                  </div>
                  <div>
                    <div className="text-xs text-muted-foreground">Backoff</div>
                    <div className="font-medium">{policy.retryPolicy!.backoffMultiplier}x</div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </TabsContent>

        {/* Pipeline Tester Tab */}
        <TabsContent value="evaluate" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Guardrail Pipeline Tester</CardTitle>
              <CardDescription>
                Test inputs against all guardrail rules with self-healing remediation
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Trace ID</Label>
                  <Input
                    placeholder="trace-abc-123"
                    value={evalInput.traceId}
                    onChange={(e) => setEvalInput({ ...evalInput, traceId: e.target.value })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Output</Label>
                  <Input
                    placeholder="Agent output text..."
                    value={evalInput.output}
                    onChange={(e) => setEvalInput({ ...evalInput, output: e.target.value })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Cost (USD)</Label>
                  <Input
                    type="number"
                    step="0.01"
                    value={evalInput.costUsd}
                    onChange={(e) => setEvalInput({ ...evalInput, costUsd: parseFloat(e.target.value) || 0 })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Latency (ms)</Label>
                  <Input
                    type="number"
                    value={evalInput.latencyMs}
                    onChange={(e) => setEvalInput({ ...evalInput, latencyMs: parseInt(e.target.value) || 0 })}
                  />
                </div>
              </div>
              <Button onClick={handleEvaluate} disabled={!evalInput.traceId || evaluateMutation.isPending}>
                <Play className="h-4 w-4 mr-2" />
                {evaluateMutation.isPending ? "Evaluating..." : "Evaluate Pipeline"}
              </Button>

              {evalResult && (
                <div className="mt-4 space-y-3">
                  <div className="flex items-center gap-2">
                    {evalResult.passed ? (
                      <CheckCircle className="h-5 w-5 text-green-600" />
                    ) : (
                      <XCircle className="h-5 w-5 text-red-600" />
                    )}
                    <span className="font-medium">
                      {evalResult.passed ? "Pipeline Passed" : "Pipeline Failed"}
                    </span>
                    {evalResult.remediated && (
                      <Badge variant="secondary">Auto-Remediated</Badge>
                    )}
                    <span className="text-xs text-muted-foreground ml-auto">
                      {evalResult.totalLatencyMs}ms total
                    </span>
                  </div>

                  {evalResult.blockedReason && (
                    <div className="text-sm text-red-600 bg-red-50 p-2 rounded">
                      Blocked: {evalResult.blockedReason}
                    </div>
                  )}

                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Rule</TableHead>
                        <TableHead>Type</TableHead>
                        <TableHead>Result</TableHead>
                        <TableHead>Violation</TableHead>
                        <TableHead>Remediation</TableHead>
                        <TableHead>Latency</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {evalResult.evaluations.map((evaluation, idx) => (
                        <TableRow key={idx}>
                          <TableCell className="font-medium">{evaluation.ruleName}</TableCell>
                          <TableCell>
                            <Badge variant="outline">{evaluation.ruleType}</Badge>
                          </TableCell>
                          <TableCell>
                            {evaluation.passed ? (
                              <CheckCircle className="h-4 w-4 text-green-600" />
                            ) : (
                              <XCircle className="h-4 w-4 text-red-600" />
                            )}
                          </TableCell>
                          <TableCell className="text-xs text-muted-foreground">
                            {evaluation.violationMessage || "—"}
                          </TableCell>
                          <TableCell>
                            {evaluation.remediated ? (
                              <Badge variant="secondary">{evaluation.remediationAction}</Badge>
                            ) : (
                              <span className="text-xs text-muted-foreground">—</span>
                            )}
                          </TableCell>
                          <TableCell className="text-xs">{evaluation.latencyMs}ms</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
