"use client";

import * as React from "react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Shield,
  AlertTriangle,
  Activity,
  Clock,
  TrendingDown,
  TrendingUp,
  CheckCircle,
  XCircle,
  Eye,
} from "lucide-react";

// --- Mock Data ---

const summaryCards = [
  { title: "Active Monitors", value: 12, icon: Shield },
  { title: "Regressions Detected", value: 3, icon: AlertTriangle },
  { title: "Unresolved Issues", value: 5, icon: XCircle },
  { title: "Avg Detection Time", value: "4.2 min", icon: Clock },
];

const metricHealth = [
  {
    name: "Quality",
    status: "healthy" as const,
    current: "94.2%",
    baseline: "93.8%",
    trend: "up" as const,
  },
  {
    name: "Latency",
    status: "warning" as const,
    current: "320ms",
    baseline: "280ms",
    trend: "up" as const,
  },
  {
    name: "Cost",
    status: "critical" as const,
    current: "$1.42",
    baseline: "$0.98",
    trend: "up" as const,
  },
  {
    name: "Error Rate",
    status: "healthy" as const,
    current: "0.8%",
    baseline: "1.1%",
    trend: "down" as const,
  },
];

const recentDetections = [
  {
    id: "det-001",
    severity: "critical" as const,
    metric: "Cost per Trace",
    delta: "+44.9%",
    method: "CUSUM",
    status: "unresolved",
    detectedAt: "2024-12-18 14:32",
  },
  {
    id: "det-002",
    severity: "high" as const,
    metric: "P95 Latency",
    delta: "+28.3%",
    method: "Z-Score",
    status: "acknowledged",
    detectedAt: "2024-12-18 12:15",
  },
  {
    id: "det-003",
    severity: "medium" as const,
    metric: "Error Rate",
    delta: "+12.1%",
    method: "Moving Average",
    status: "unresolved",
    detectedAt: "2024-12-18 09:44",
  },
  {
    id: "det-004",
    severity: "low" as const,
    metric: "Token Usage",
    delta: "+5.2%",
    method: "IQR",
    status: "resolved",
    detectedAt: "2024-12-17 22:10",
  },
  {
    id: "det-005",
    severity: "high" as const,
    metric: "Quality Score",
    delta: "-8.7%",
    method: "CUSUM",
    status: "unresolved",
    detectedAt: "2024-12-17 18:55",
  },
  {
    id: "det-006",
    severity: "medium" as const,
    metric: "Completion Rate",
    delta: "-6.3%",
    method: "Z-Score",
    status: "false_positive",
    detectedAt: "2024-12-17 15:20",
  },
];

// --- Helpers ---

const statusColors: Record<string, string> = {
  healthy: "bg-green-500/15 text-green-700 border-green-500/25",
  warning: "bg-yellow-500/15 text-yellow-700 border-yellow-500/25",
  critical: "bg-red-500/15 text-red-700 border-red-500/25",
};

const severityColors: Record<string, string> = {
  critical: "bg-red-500/15 text-red-700 border-red-500/25",
  high: "bg-orange-500/15 text-orange-700 border-orange-500/25",
  medium: "bg-yellow-500/15 text-yellow-700 border-yellow-500/25",
  low: "bg-blue-500/15 text-blue-700 border-blue-500/25",
};

// --- Component ---

export function RegressionDetectionDashboard() {
  return (
    <div className="space-y-6">
      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {summaryCards.map((card) => (
          <Card key={card.title}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {card.title}
              </CardTitle>
              <card.icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{card.value}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Metric Health Overview */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Metric Health Overview</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {metricHealth.map((metric) => (
              <div
                key={metric.name}
                className="flex items-center justify-between rounded-lg border p-3"
              >
                <div className="flex items-center gap-3">
                  <Badge className={statusColors[metric.status]}>
                    {metric.status}
                  </Badge>
                  <span className="font-medium">{metric.name}</span>
                </div>
                <div className="flex items-center gap-6 text-sm">
                  <div>
                    <span className="text-muted-foreground">Current: </span>
                    <span className="font-medium">{metric.current}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Baseline: </span>
                    <span className="font-medium">{metric.baseline}</span>
                  </div>
                  <div className="flex items-center gap-1">
                    {metric.trend === "up" ? (
                      <TrendingUp className="h-4 w-4 text-muted-foreground" />
                    ) : (
                      <TrendingDown className="h-4 w-4 text-muted-foreground" />
                    )}
                    {metric.status === "healthy" ? (
                      <CheckCircle className="h-4 w-4 text-green-600" />
                    ) : metric.status === "warning" ? (
                      <AlertTriangle className="h-4 w-4 text-yellow-600" />
                    ) : (
                      <XCircle className="h-4 w-4 text-red-600" />
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Recent Detections */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Recent Detections</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-2 pr-4 font-medium">Severity</th>
                  <th className="pb-2 pr-4 font-medium">Metric</th>
                  <th className="pb-2 pr-4 font-medium">Delta%</th>
                  <th className="pb-2 pr-4 font-medium">Method</th>
                  <th className="pb-2 pr-4 font-medium">Status</th>
                  <th className="pb-2 pr-4 font-medium">Detected At</th>
                  <th className="pb-2 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {recentDetections.map((d) => (
                  <tr key={d.id} className="border-b last:border-0">
                    <td className="py-3 pr-4">
                      <Badge className={severityColors[d.severity]}>
                        {d.severity}
                      </Badge>
                    </td>
                    <td className="py-3 pr-4 font-medium">{d.metric}</td>
                    <td className="py-3 pr-4 font-mono">{d.delta}</td>
                    <td className="py-3 pr-4">{d.method}</td>
                    <td className="py-3 pr-4">
                      <Badge variant="outline" className="capitalize">
                        {d.status.replace("_", " ")}
                      </Badge>
                    </td>
                    <td className="py-3 pr-4 text-muted-foreground">
                      {d.detectedAt}
                    </td>
                    <td className="py-3">
                      <div className="flex items-center gap-1">
                        <Button variant="ghost" size="sm" title="Acknowledge">
                          <Eye className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="sm" title="Resolve">
                          <CheckCircle className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          title="Mark False Positive"
                        >
                          <XCircle className="h-4 w-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Detection Configuration */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-lg">Detection Configuration</CardTitle>
          <Button>
            <Activity className="mr-2 h-4 w-4" />
            Create New Monitor
          </Button>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Configure regression detection monitors with custom thresholds,
            detection methods, and alert channels.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
