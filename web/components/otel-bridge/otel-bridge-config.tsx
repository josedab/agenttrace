"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import {
  Settings,
  Globe,
  Download,
  BarChart3,
  Plus,
  ToggleLeft,
  ToggleRight,
  Activity,
  CheckCircle2,
  XCircle,
  Clock,
} from "lucide-react";

interface OTelConfig {
  exportEnabled: boolean;
  importEnabled: boolean;
  samplingRate: number;
  resourceAttributes: Record<string, string>;
}

interface ExportDestination {
  id: string;
  type: "jaeger" | "tempo" | "datadog" | "otlp" | "zipkin";
  name: string;
  endpoint: string;
  enabled: boolean;
}

interface BridgeStats {
  exportStats: {
    totalSpans: number;
    successCount: number;
    errorCount: number;
    avgLatencyMs: number;
    last24hCount: number;
  };
  importStats: {
    totalSpans: number;
    successCount: number;
    errorCount: number;
    avgLatencyMs: number;
    last24hCount: number;
  };
}

const DESTINATION_TYPES = ["jaeger", "tempo", "datadog", "otlp", "zipkin"] as const;

const DEST_ICONS: Record<string, string> = {
  jaeger: "🔵",
  tempo: "🟠",
  datadog: "🟣",
  otlp: "🟢",
  zipkin: "🔴",
};

export function OTelBridgeConfig() {
  const queryClient = useQueryClient();
  const [showDestForm, setShowDestForm] = React.useState(false);
  const [destForm, setDestForm] = React.useState({
    type: "otlp" as (typeof DESTINATION_TYPES)[number],
    name: "",
    endpoint: "",
  });
  const [importForm, setImportForm] = React.useState({
    correlateByTraceId: true,
    createMissingTraces: false,
  });

  const { data: config } = useQuery<OTelConfig>({
    queryKey: ["otel-config"],
    queryFn: () => api.otelBridge.getConfig(),
  });

  const { data: destinations } = useQuery<ExportDestination[]>({
    queryKey: ["otel-destinations"],
    queryFn: () => api.otelBridge.listDestinations(),
  });

  const { data: stats } = useQuery<BridgeStats>({
    queryKey: ["otel-stats"],
    queryFn: () => api.otelBridge.getStats(),
  });

  const updateConfigMutation = useMutation({
    mutationFn: (data: Partial<OTelConfig>) => api.otelBridge.updateConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["otel-config"] });
    },
  });

  const addDestMutation = useMutation({
    mutationFn: (data: typeof destForm) => api.otelBridge.addDestination(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["otel-destinations"] });
      setShowDestForm(false);
      setDestForm({ type: "otlp", name: "", endpoint: "" });
    },
  });

  const toggleDestMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      api.otelBridge.updateDestination(id, { enabled }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["otel-destinations"] });
    },
  });

  const importMutation = useMutation({
    mutationFn: (data: typeof importForm) => api.otelBridge.importSpans(data),
  });

  const [samplingRate, setSamplingRate] = React.useState(config?.samplingRate ?? 100);
  const [attrKey, setAttrKey] = React.useState("");
  const [attrValue, setAttrValue] = React.useState("");

  React.useEffect(() => {
    if (config) setSamplingRate(config.samplingRate);
  }, [config]);

  return (
    <Tabs defaultValue="configuration" className="space-y-6">
      <TabsList>
        <TabsTrigger value="configuration" className="flex items-center gap-1">
          <Settings className="h-4 w-4" />
          Configuration
        </TabsTrigger>
        <TabsTrigger value="destinations" className="flex items-center gap-1">
          <Globe className="h-4 w-4" />
          Destinations
        </TabsTrigger>
        <TabsTrigger value="import" className="flex items-center gap-1">
          <Download className="h-4 w-4" />
          Import
        </TabsTrigger>
        <TabsTrigger value="stats" className="flex items-center gap-1">
          <BarChart3 className="h-4 w-4" />
          Stats
        </TabsTrigger>
      </TabsList>

      {/* Configuration Tab */}
      <TabsContent value="configuration" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Bridge Settings</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <span className="text-sm font-medium">Export Enabled</span>
                <p className="text-xs text-muted-foreground">Send spans to configured destinations</p>
              </div>
              <button
                onClick={() =>
                  updateConfigMutation.mutate({ exportEnabled: !config?.exportEnabled })
                }
              >
                {config?.exportEnabled ? (
                  <ToggleRight className="h-6 w-6 text-green-500" />
                ) : (
                  <ToggleLeft className="h-6 w-6 text-muted-foreground" />
                )}
              </button>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <span className="text-sm font-medium">Import Enabled</span>
                <p className="text-xs text-muted-foreground">Accept incoming OTel spans</p>
              </div>
              <button
                onClick={() =>
                  updateConfigMutation.mutate({ importEnabled: !config?.importEnabled })
                }
              >
                {config?.importEnabled ? (
                  <ToggleRight className="h-6 w-6 text-green-500" />
                ) : (
                  <ToggleLeft className="h-6 w-6 text-muted-foreground" />
                )}
              </button>
            </div>

            {/* Sampling Rate */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="text-sm font-medium">Sampling Rate</label>
                <span className="text-sm text-muted-foreground">{samplingRate}%</span>
              </div>
              <input
                type="range"
                min={0}
                max={100}
                value={samplingRate}
                onChange={(e) => setSamplingRate(parseInt(e.target.value))}
                onMouseUp={() => updateConfigMutation.mutate({ samplingRate })}
                className="w-full"
              />
            </div>

            {/* Resource Attributes */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Resource Attributes</label>
              {config?.resourceAttributes && Object.entries(config.resourceAttributes).length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {Object.entries(config.resourceAttributes).map(([key, value]) => (
                    <Badge key={key} variant="secondary">
                      {key}: {value}
                    </Badge>
                  ))}
                </div>
              )}
              <div className="flex gap-2">
                <Input
                  placeholder="Key"
                  value={attrKey}
                  onChange={(e) => setAttrKey(e.target.value)}
                  className="flex-1"
                />
                <Input
                  placeholder="Value"
                  value={attrValue}
                  onChange={(e) => setAttrValue(e.target.value)}
                  className="flex-1"
                />
                <Button
                  variant="outline"
                  size="sm"
                  disabled={!attrKey || !attrValue}
                  onClick={() => {
                    updateConfigMutation.mutate({
                      resourceAttributes: {
                        ...config?.resourceAttributes,
                        [attrKey]: attrValue,
                      },
                    });
                    setAttrKey("");
                    setAttrValue("");
                  }}
                >
                  <Plus className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      {/* Destinations Tab */}
      <TabsContent value="destinations" className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium">Export Destinations</h3>
          <Button size="sm" onClick={() => setShowDestForm(!showDestForm)}>
            <Plus className="h-4 w-4 mr-1" />
            Add Destination
          </Button>
        </div>

        {showDestForm && (
          <Card>
            <CardContent className="pt-4 space-y-4">
              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-1">
                  <label className="text-sm font-medium">Type</label>
                  <Select
                    value={destForm.type}
                    onValueChange={(v) =>
                      setDestForm({ ...destForm, type: v as (typeof DESTINATION_TYPES)[number] })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {DESTINATION_TYPES.map((t) => (
                        <SelectItem key={t} value={t}>
                          {DEST_ICONS[t]} {t.charAt(0).toUpperCase() + t.slice(1)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium">Name</label>
                  <Input
                    placeholder="Destination name"
                    value={destForm.name}
                    onChange={(e) => setDestForm({ ...destForm, name: e.target.value })}
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium">Endpoint</label>
                  <Input
                    placeholder="https://..."
                    value={destForm.endpoint}
                    onChange={(e) => setDestForm({ ...destForm, endpoint: e.target.value })}
                  />
                </div>
              </div>
              <div className="flex gap-2 justify-end">
                <Button variant="outline" size="sm" onClick={() => setShowDestForm(false)}>
                  Cancel
                </Button>
                <Button
                  size="sm"
                  onClick={() => addDestMutation.mutate(destForm)}
                  disabled={addDestMutation.isPending || !destForm.name || !destForm.endpoint}
                >
                  Add Destination
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        <div className="space-y-3">
          {destinations?.map((dest) => (
            <Card key={dest.id}>
              <CardContent className="pt-4">
                <div className="flex items-center gap-4">
                  <span className="text-lg">{DEST_ICONS[dest.type] ?? "⚪"}</span>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{dest.name}</span>
                      <Badge variant="outline" className="text-xs">
                        {dest.type}
                      </Badge>
                    </div>
                    <p className="text-xs text-muted-foreground font-mono truncate mt-0.5">
                      {dest.endpoint}
                    </p>
                  </div>
                  <button
                    onClick={() => toggleDestMutation.mutate({ id: dest.id, enabled: !dest.enabled })}
                    disabled={toggleDestMutation.isPending}
                  >
                    {dest.enabled ? (
                      <ToggleRight className="h-6 w-6 text-green-500" />
                    ) : (
                      <ToggleLeft className="h-6 w-6 text-muted-foreground" />
                    )}
                  </button>
                </div>
              </CardContent>
            </Card>
          ))}

          {(!destinations || destinations.length === 0) && (
            <div className="text-center py-8 text-muted-foreground text-sm">
              No export destinations configured
            </div>
          )}
        </div>
      </TabsContent>

      {/* Import Tab */}
      <TabsContent value="import" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Download className="h-4 w-4" />
              Import OTel Spans
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <span className="text-sm font-medium">Correlate by Trace ID</span>
                <p className="text-xs text-muted-foreground">Match imported spans to existing traces</p>
              </div>
              <button
                onClick={() =>
                  setImportForm({ ...importForm, correlateByTraceId: !importForm.correlateByTraceId })
                }
              >
                {importForm.correlateByTraceId ? (
                  <ToggleRight className="h-6 w-6 text-green-500" />
                ) : (
                  <ToggleLeft className="h-6 w-6 text-muted-foreground" />
                )}
              </button>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <span className="text-sm font-medium">Create Missing Traces</span>
                <p className="text-xs text-muted-foreground">Auto-create traces for unmatched span trace IDs</p>
              </div>
              <button
                onClick={() =>
                  setImportForm({ ...importForm, createMissingTraces: !importForm.createMissingTraces })
                }
              >
                {importForm.createMissingTraces ? (
                  <ToggleRight className="h-6 w-6 text-green-500" />
                ) : (
                  <ToggleLeft className="h-6 w-6 text-muted-foreground" />
                )}
              </button>
            </div>
            <div className="flex justify-end">
              <Button
                size="sm"
                onClick={() => importMutation.mutate(importForm)}
                disabled={importMutation.isPending}
              >
                <Download className="h-4 w-4 mr-1" />
                Start Import
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      {/* Stats Tab */}
      <TabsContent value="stats" className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Export Stats */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <Activity className="h-4 w-4" />
                Export Stats
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {stats ? (
                <>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Total Spans</span>
                    <span className="font-medium">{stats.exportStats.totalSpans.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <CheckCircle2 className="h-3 w-3 text-green-500" /> Success
                    </span>
                    <span className="font-medium">{stats.exportStats.successCount.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <XCircle className="h-3 w-3 text-red-500" /> Errors
                    </span>
                    <span className="font-medium">{stats.exportStats.errorCount.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <Clock className="h-3 w-3" /> Avg Latency
                    </span>
                    <span className="font-medium">{stats.exportStats.avgLatencyMs.toFixed(1)}ms</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Last 24h</span>
                    <span className="font-medium">{stats.exportStats.last24hCount.toLocaleString()}</span>
                  </div>
                </>
              ) : (
                <div className="h-32 bg-muted animate-pulse rounded-lg" />
              )}
            </CardContent>
          </Card>

          {/* Import Stats */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <Download className="h-4 w-4" />
                Import Stats
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {stats ? (
                <>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Total Spans</span>
                    <span className="font-medium">{stats.importStats.totalSpans.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <CheckCircle2 className="h-3 w-3 text-green-500" /> Success
                    </span>
                    <span className="font-medium">{stats.importStats.successCount.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <XCircle className="h-3 w-3 text-red-500" /> Errors
                    </span>
                    <span className="font-medium">{stats.importStats.errorCount.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <Clock className="h-3 w-3" /> Avg Latency
                    </span>
                    <span className="font-medium">{stats.importStats.avgLatencyMs.toFixed(1)}ms</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Last 24h</span>
                    <span className="font-medium">{stats.importStats.last24hCount.toLocaleString()}</span>
                  </div>
                </>
              ) : (
                <div className="h-32 bg-muted animate-pulse rounded-lg" />
              )}
            </CardContent>
          </Card>
        </div>
      </TabsContent>
    </Tabs>
  );
}
