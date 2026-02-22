"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
  Shield,
  DollarSign,
  Clock,
  FileX,
  Code,
  AlertTriangle,
  Plus,
} from "lucide-react";

interface GuardRule {
  id: string;
  name: string;
  description: string;
  type: string;
  action: string;
  enabled: boolean;
  config: {
    maxCostPerTrace?: number;
    maxLatencyMs?: number;
    restrictedPaths?: string[];
    blockedPatterns?: string[];
  };
}

interface ViolationStats {
  totalViolations: number;
  violationsByRule: Record<string, number>;
  violationsByAction: Record<string, number>;
  recentViolations: number;
}

const ruleTypeIcons: Record<string, React.ElementType> = {
  cost_limit: DollarSign,
  latency_limit: Clock,
  file_restriction: FileX,
  pattern_block: Code,
};

const actionColors: Record<string, string> = {
  block: "destructive",
  alert: "default",
  log: "secondary",
};

export function GuardrailsPanel() {
  const queryClient = useQueryClient();

  const { data: rules, isLoading: rulesLoading } = useQuery<GuardRule[]>({
    queryKey: ["guardrails"],
    queryFn: () => api.guardrails.list(),
  });

  const { data: stats } = useQuery<ViolationStats>({
    queryKey: ["guardrails-stats"],
    queryFn: () => api.guardrails.getViolationStats(),
  });

  if (rulesLoading) {
    return <div className="h-64 bg-muted animate-pulse rounded-lg" />;
  }

  return (
    <div className="space-y-6">
      {/* Stats Overview */}
      {stats && stats.totalViolations > 0 && (
        <div className="grid grid-cols-3 gap-4">
          <Card>
            <CardContent className="pt-4">
              <div className="text-xs text-muted-foreground">Total Violations</div>
              <div className="text-2xl font-bold">{stats.totalViolations}</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-4">
              <div className="text-xs text-muted-foreground">Blocked</div>
              <div className="text-2xl font-bold text-destructive">
                {stats.violationsByAction?.block || 0}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-4">
              <div className="text-xs text-muted-foreground">Last 24h</div>
              <div className="text-2xl font-bold">{stats.recentViolations}</div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Rules List */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base flex items-center gap-2">
            <Shield className="h-4 w-4" />
            Guard Rules
          </CardTitle>
          <Button size="sm">
            <Plus className="h-4 w-4 mr-1" />
            Add Rule
          </Button>
        </CardHeader>
        <CardContent>
          {!rules || rules.length === 0 ? (
            <div className="text-center py-8">
              <Shield className="h-12 w-12 mx-auto text-muted-foreground mb-3" />
              <p className="text-sm text-muted-foreground">
                No guardrails configured. Add rules to protect your agents.
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Rule</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Action</TableHead>
                  <TableHead>Config</TableHead>
                  <TableHead>Violations</TableHead>
                  <TableHead className="w-[80px]">Enabled</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rules.map((rule) => {
                  const Icon = ruleTypeIcons[rule.type] || AlertTriangle;
                  return (
                    <TableRow key={rule.id}>
                      <TableCell>
                        <div className="font-medium">{rule.name}</div>
                        {rule.description && (
                          <div className="text-xs text-muted-foreground">
                            {rule.description}
                          </div>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Icon className="h-3 w-3" />
                          <span className="text-sm">{rule.type.replace("_", " ")}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={
                            (actionColors[rule.action] as any) || "outline"
                          }
                        >
                          {rule.action}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-xs font-mono text-muted-foreground">
                        {formatConfig(rule)}
                      </TableCell>
                      <TableCell>
                        {stats?.violationsByRule?.[rule.id] || 0}
                      </TableCell>
                      <TableCell>
                        <Switch checked={rule.enabled} />
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function formatConfig(rule: GuardRule): string {
  const { config, type } = rule;
  switch (type) {
    case "cost_limit":
      return config.maxCostPerTrace != null
        ? `max $${config.maxCostPerTrace}/trace`
        : "—";
    case "latency_limit":
      return config.maxLatencyMs != null
        ? `max ${config.maxLatencyMs}ms`
        : "—";
    case "file_restriction":
      return config.restrictedPaths?.length
        ? `${config.restrictedPaths.length} paths`
        : "—";
    case "pattern_block":
      return config.blockedPatterns?.length
        ? `${config.blockedPatterns.length} patterns`
        : "—";
    default:
      return "—";
  }
}
