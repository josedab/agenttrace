"use client";

import * as React from "react";
import {
  AlertTriangle,
  Bell,
  Search,
  Shield,
  Clock,
  TrendingUp,
  TrendingDown,
  Minus,
  CheckCircle,
  XCircle,
  Eye,
  Settings,
  Plus,
  Send,
  Activity,
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
interface AlertDashboardStats {
  openAnomalies: number;
  criticalAlerts: number;
  activeInvestigations: number;
  mttrMinutes: number;
  alertsSentToday: number;
  anomalyTrend: string;
}

interface CorrelatedAnomaly {
  id: string;
  anomalyType: string;
  severity: string;
  title: string;
  description: string;
  affectedTraces: string[];
  correlation: number;
  rootCauses: { category: string; description: string; impact: number }[];
  remediations: { priority: number; action: string; description: string }[];
  status: string;
  detectedAt: string;
}

interface AlertDeliveryChannel {
  id: string;
  name: string;
  type: string;
  enabled: boolean;
  testStatus: string;
  createdAt: string;
}

interface CorrelationRule {
  id: string;
  name: string;
  anomalyTypes: string[];
  windowMinutes: number;
  minCorrelation: number;
  severity: string;
  enabled: boolean;
}

interface RCAInvestigation {
  id: string;
  anomalyId: string;
  title: string;
  status: string;
  findings: { type: string; description: string; confidence: number }[];
  timeline: { action: string; details: string; timestamp: string }[];
  rootCause?: string;
  resolution?: string;
  createdAt: string;
  updatedAt: string;
}

// Severity badge helper
function SeverityBadge({ severity }: { severity: string }) {
  const variants: Record<string, string> = {
    emergency: "bg-red-600 text-white",
    critical: "bg-red-500 text-white",
    warning: "bg-yellow-500 text-white",
    info: "bg-blue-500 text-white",
  };

  return (
    <Badge className={cn("text-xs", variants[severity] || "bg-gray-500 text-white")}>
      {severity}
    </Badge>
  );
}

// Status badge helper
function StatusBadge({ status }: { status: string }) {
  const config: Record<string, { variant: "default" | "secondary" | "destructive" | "outline"; icon: React.ReactNode }> = {
    open: { variant: "destructive", icon: <AlertTriangle className="h-3 w-3 mr-1" /> },
    acknowledged: { variant: "outline", icon: <Eye className="h-3 w-3 mr-1" /> },
    investigating: { variant: "default", icon: <Search className="h-3 w-3 mr-1" /> },
    resolved: { variant: "secondary", icon: <CheckCircle className="h-3 w-3 mr-1" /> },
    dismissed: { variant: "secondary", icon: <XCircle className="h-3 w-3 mr-1" /> },
  };

  const { variant, icon } = config[status] || { variant: "outline" as const, icon: null };

  return (
    <Badge variant={variant} className="text-xs">
      {icon}
      {status}
    </Badge>
  );
}

// Trend icon helper
function TrendIcon({ trend }: { trend: string }) {
  if (trend === "increasing") return <TrendingUp className="h-4 w-4 text-red-500" />;
  if (trend === "decreasing") return <TrendingDown className="h-4 w-4 text-green-500" />;
  return <Minus className="h-4 w-4 text-muted-foreground" />;
}

// Stats Cards
function StatsCards({ stats }: { stats: AlertDashboardStats }) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">Open Anomalies</CardTitle>
          <AlertTriangle className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.openAnomalies}</div>
          <div className="flex items-center text-xs text-muted-foreground mt-1">
            <TrendIcon trend={stats.anomalyTrend} />
            <span className="ml-1">{stats.anomalyTrend}</span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">Critical Alerts</CardTitle>
          <Bell className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.criticalAlerts}</div>
          <p className="text-xs text-muted-foreground mt-1">Requiring immediate attention</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">Active Investigations</CardTitle>
          <Search className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.activeInvestigations}</div>
          <p className="text-xs text-muted-foreground mt-1">In progress</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">MTTR</CardTitle>
          <Clock className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats.mttrMinutes.toFixed(0)}m</div>
          <p className="text-xs text-muted-foreground mt-1">Mean time to resolve</p>
        </CardContent>
      </Card>
    </div>
  );
}

// Anomaly List
function AnomalyList({ anomalies }: { anomalies: CorrelatedAnomaly[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Correlated Anomalies</CardTitle>
        <CardDescription>Anomalies detected across traces with root cause correlation</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Title</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Severity</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Correlation</TableHead>
                <TableHead>Traces</TableHead>
                <TableHead>Detected</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {anomalies.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center text-muted-foreground py-8">
                    No anomalies detected
                  </TableCell>
                </TableRow>
              ) : (
                anomalies.map((anomaly) => (
                  <TableRow key={anomaly.id}>
                    <TableCell>
                      <span className="font-medium">{anomaly.title}</span>
                      {anomaly.description && (
                        <p className="text-xs text-muted-foreground mt-0.5 truncate max-w-[250px]">
                          {anomaly.description}
                        </p>
                      )}
                    </TableCell>
                    <TableCell>
                      <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{anomaly.anomalyType}</code>
                    </TableCell>
                    <TableCell>
                      <SeverityBadge severity={anomaly.severity} />
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={anomaly.status} />
                    </TableCell>
                    <TableCell>
                      <span className="text-sm">{(anomaly.correlation * 100).toFixed(0)}%</span>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm">{anomaly.affectedTraces.length}</span>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-muted-foreground">
                        {new Date(anomaly.detectedAt).toLocaleDateString()}
                      </span>
                    </TableCell>
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

// Alert Channels
function AlertChannelList({ channels }: { channels: AlertDeliveryChannel[] }) {
  const channelIcons: Record<string, React.ReactNode> = {
    slack: <Send className="h-4 w-4" />,
    pagerduty: <Bell className="h-4 w-4" />,
    opsgenie: <Shield className="h-4 w-4" />,
    email: <Send className="h-4 w-4" />,
    webhook: <Activity className="h-4 w-4" />,
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle>Alert Channels</CardTitle>
          <CardDescription>Configure delivery channels for alerts</CardDescription>
        </div>
        <Button size="sm">
          <Plus className="h-4 w-4 mr-1" />
          Add Channel
        </Button>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {channels.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-4">No alert channels configured</p>
          ) : (
            channels.map((channel) => (
              <div key={channel.id} className="flex items-center justify-between p-3 border rounded-lg">
                <div className="flex items-center gap-3">
                  {channelIcons[channel.type] || <Settings className="h-4 w-4" />}
                  <div>
                    <p className="text-sm font-medium">{channel.name}</p>
                    <p className="text-xs text-muted-foreground">{channel.type}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant={channel.enabled ? "default" : "secondary"}>
                    {channel.enabled ? "Active" : "Disabled"}
                  </Badge>
                  <Badge
                    variant="outline"
                    className={cn(
                      channel.testStatus === "success" && "border-green-500 text-green-600",
                      channel.testStatus === "failed" && "border-red-500 text-red-600"
                    )}
                  >
                    {channel.testStatus || "untested"}
                  </Badge>
                </div>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}

// Correlation Rules
function CorrelationRuleList({ rules }: { rules: CorrelationRule[] }) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle>Correlation Rules</CardTitle>
          <CardDescription>Define how anomalies are correlated and alerted</CardDescription>
        </div>
        <Button size="sm">
          <Plus className="h-4 w-4 mr-1" />
          Add Rule
        </Button>
      </CardHeader>
      <CardContent>
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Anomaly Types</TableHead>
                <TableHead>Window</TableHead>
                <TableHead>Min Correlation</TableHead>
                <TableHead>Severity</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rules.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                    No correlation rules configured
                  </TableCell>
                </TableRow>
              ) : (
                rules.map((rule) => (
                  <TableRow key={rule.id}>
                    <TableCell className="font-medium">{rule.name}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {rule.anomalyTypes.map((type) => (
                          <code key={type} className="text-xs bg-muted px-1.5 py-0.5 rounded">{type}</code>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell>{rule.windowMinutes}m</TableCell>
                    <TableCell>{(rule.minCorrelation * 100).toFixed(0)}%</TableCell>
                    <TableCell>
                      <SeverityBadge severity={rule.severity} />
                    </TableCell>
                    <TableCell>
                      <Badge variant={rule.enabled ? "default" : "secondary"}>
                        {rule.enabled ? "Enabled" : "Disabled"}
                      </Badge>
                    </TableCell>
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

// Investigation Timeline
function InvestigationList({ investigations }: { investigations: RCAInvestigation[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Investigations</CardTitle>
        <CardDescription>Active and recent root cause investigations</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {investigations.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-4">No investigations in progress</p>
          ) : (
            investigations.map((investigation) => (
              <div key={investigation.id} className="border rounded-lg p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-medium">{investigation.title}</h4>
                  <StatusBadge status={investigation.status} />
                </div>
                {investigation.rootCause && (
                  <div className="text-xs bg-muted p-2 rounded">
                    <span className="font-medium">Root Cause:</span> {investigation.rootCause}
                  </div>
                )}
                {investigation.timeline.length > 0 && (
                  <div className="space-y-1">
                    {investigation.timeline.slice(-3).map((event, i) => (
                      <div key={i} className="flex items-center gap-2 text-xs text-muted-foreground">
                        <div className="h-1.5 w-1.5 rounded-full bg-muted-foreground" />
                        <span className="font-medium">{event.action}</span>
                        <span>{event.details}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}

// Root Cause Breakdown (from anomalies)
function RootCauseBreakdown({ anomalies }: { anomalies: CorrelatedAnomaly[] }) {
  const categoryCounts: Record<string, number> = {};
  anomalies.forEach((a) => {
    a.rootCauses.forEach((rc) => {
      categoryCounts[rc.category] = (categoryCounts[rc.category] || 0) + 1;
    });
  });

  const sorted = Object.entries(categoryCounts).sort((a, b) => b[1] - a[1]);
  const total = sorted.reduce((sum, [, count]) => sum + count, 0) || 1;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Root Cause Breakdown</CardTitle>
        <CardDescription>Distribution of root causes across anomalies</CardDescription>
      </CardHeader>
      <CardContent>
        {sorted.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-4">No root cause data available</p>
        ) : (
          <div className="space-y-3">
            {sorted.map(([category, count]) => {
              const pct = (count / total) * 100;
              return (
                <div key={category} className="space-y-1">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium">{category}</span>
                    <span className="text-muted-foreground">{count} ({pct.toFixed(0)}%)</span>
                  </div>
                  <div className="h-2 bg-muted rounded-full overflow-hidden">
                    <div
                      className="h-full bg-primary rounded-full transition-all"
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// Main Dashboard
export function RCADashboard() {
  const [stats, setStats] = React.useState<AlertDashboardStats>({
    openAnomalies: 0,
    criticalAlerts: 0,
    activeInvestigations: 0,
    mttrMinutes: 0,
    alertsSentToday: 0,
    anomalyTrend: "stable",
  });
  const [anomalies, setAnomalies] = React.useState<CorrelatedAnomaly[]>([]);
  const [channels, setChannels] = React.useState<AlertDeliveryChannel[]>([]);
  const [rules, setRules] = React.useState<CorrelationRule[]>([]);
  const [investigations, setInvestigations] = React.useState<RCAInvestigation[]>([]);

  React.useEffect(() => {
    // Fetch all data on mount
    Promise.all([
      fetch("/api/public/rca/dashboard").then((r) => r.ok ? r.json() : null),
      fetch("/api/public/rca/anomalies").then((r) => r.ok ? r.json() : null),
      fetch("/api/public/rca/alert-channels").then((r) => r.ok ? r.json() : null),
      fetch("/api/public/rca/correlation-rules").then((r) => r.ok ? r.json() : null),
      fetch("/api/public/rca/investigations").then((r) => r.ok ? r.json() : null),
    ]).then(([dashStats, anomalyData, channelData, ruleData, invData]) => {
      if (dashStats) setStats(dashStats);
      if (anomalyData?.anomalies) setAnomalies(anomalyData.anomalies);
      if (channelData?.channels) setChannels(channelData.channels);
      if (ruleData?.rules) setRules(ruleData.rules);
      if (invData?.investigations) setInvestigations(invData.investigations);
    }).catch(() => {
      // Silently handle fetch errors in development
    });
  }, []);

  return (
    <div className="space-y-6">
      <StatsCards stats={stats} />

      <AnomalyList anomalies={anomalies} />

      <div className="grid gap-6 lg:grid-cols-2">
        <AlertChannelList channels={channels} />
        <RootCauseBreakdown anomalies={anomalies} />
      </div>

      <CorrelationRuleList rules={rules} />

      <InvestigationList investigations={investigations} />
    </div>
  );
}
