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
  Layers,
  Plus,
  Play,
  ToggleLeft,
  ToggleRight,
  Zap,
  Clock,
} from "lucide-react";

interface EnrichmentRule {
  id: string;
  name: string;
  triggerEvent: string;
  sourceType: string;
  condition: string;
  transform: string;
  priority: number;
  enabled: boolean;
  fireCount: number;
  lastFired: string | null;
}

const TRIGGER_EVENTS = [
  "trace.created",
  "trace.completed",
  "span.created",
  "span.error",
  "eval.completed",
  "cost.threshold",
];

const SOURCE_TYPES = [
  "metadata",
  "span_attributes",
  "model_output",
  "external_api",
  "knowledge_base",
  "custom",
];

export function EnrichmentRuleBuilder() {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = React.useState(false);
  const [formData, setFormData] = React.useState({
    name: "",
    triggerEvent: "",
    sourceType: "",
    condition: "",
    transform: "",
    priority: 0,
  });

  const { data: rules, isLoading } = useQuery<EnrichmentRule[]>({
    queryKey: ["enrichment-rules"],
    queryFn: () => api.enrichment.listRules(),
  });

  const createMutation = useMutation({
    mutationFn: (data: typeof formData) => api.enrichment.createRule(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["enrichment-rules"] });
      setShowForm(false);
      setFormData({ name: "", triggerEvent: "", sourceType: "", condition: "", transform: "", priority: 0 });
    },
  });

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      api.enrichment.updateRule(id, { enabled }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["enrichment-rules"] });
    },
  });

  const testMutation = useMutation({
    mutationFn: (id: string) => api.enrichment.testRule(id),
  });

  if (isLoading) {
    return <EnrichmentRuleBuilderSkeleton />;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Layers className="h-5 w-5 text-muted-foreground" />
          <h2 className="text-lg font-semibold">Enrichment Rules</h2>
          <Badge variant="secondary">{rules?.length ?? 0} rules</Badge>
        </div>
        <Button onClick={() => setShowForm(!showForm)} size="sm">
          <Plus className="h-4 w-4 mr-1" />
          Create Rule
        </Button>
      </div>

      {/* Create Rule Form */}
      {showForm && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">New Enrichment Rule</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1">
                <label className="text-sm font-medium">Name</label>
                <Input
                  placeholder="Rule name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                />
              </div>
              <div className="space-y-1">
                <label className="text-sm font-medium">Priority</label>
                <Input
                  type="number"
                  placeholder="0"
                  value={formData.priority}
                  onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) || 0 })}
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1">
                <label className="text-sm font-medium">Trigger Event</label>
                <Select
                  value={formData.triggerEvent}
                  onValueChange={(v) => setFormData({ ...formData, triggerEvent: v })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select trigger" />
                  </SelectTrigger>
                  <SelectContent>
                    {TRIGGER_EVENTS.map((evt) => (
                      <SelectItem key={evt} value={evt}>
                        {evt}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1">
                <label className="text-sm font-medium">Source Type</label>
                <Select
                  value={formData.sourceType}
                  onValueChange={(v) => setFormData({ ...formData, sourceType: v })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select source" />
                  </SelectTrigger>
                  <SelectContent>
                    {SOURCE_TYPES.map((src) => (
                      <SelectItem key={src} value={src}>
                        {src.replace(/_/g, " ")}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium">Condition</label>
              <Input
                placeholder='e.g. span.attributes.model == "gpt-4"'
                value={formData.condition}
                onChange={(e) => setFormData({ ...formData, condition: e.target.value })}
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium">Transform</label>
              <Input
                placeholder='e.g. set metadata.tier = "premium"'
                value={formData.transform}
                onChange={(e) => setFormData({ ...formData, transform: e.target.value })}
              />
            </div>
            <div className="flex gap-2 justify-end">
              <Button variant="outline" size="sm" onClick={() => setShowForm(false)}>
                Cancel
              </Button>
              <Button
                size="sm"
                onClick={() => createMutation.mutate(formData)}
                disabled={createMutation.isPending || !formData.name || !formData.triggerEvent}
              >
                Create Rule
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Rules List */}
      <div className="space-y-3">
        {rules?.map((rule) => (
          <Card key={rule.id}>
            <CardContent className="pt-4">
              <div className="flex items-center gap-4">
                <button
                  onClick={() => toggleMutation.mutate({ id: rule.id, enabled: !rule.enabled })}
                  className="shrink-0"
                  disabled={toggleMutation.isPending}
                >
                  {rule.enabled ? (
                    <ToggleRight className="h-6 w-6 text-green-500" />
                  ) : (
                    <ToggleLeft className="h-6 w-6 text-muted-foreground" />
                  )}
                </button>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="font-medium truncate">{rule.name}</span>
                    <Badge variant="outline" className="text-xs">
                      priority: {rule.priority}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-3 text-xs text-muted-foreground mt-1">
                    <span className="flex items-center gap-1">
                      <Zap className="h-3 w-3" />
                      {rule.triggerEvent}
                    </span>
                    <span>source: {rule.sourceType}</span>
                    <span>{rule.fireCount.toLocaleString()} fires</span>
                    {rule.lastFired && (
                      <span className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        {new Date(rule.lastFired).toLocaleDateString()}
                      </span>
                    )}
                  </div>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => testMutation.mutate(rule.id)}
                  disabled={testMutation.isPending}
                >
                  <Play className="h-3 w-3 mr-1" />
                  Test Rule
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}

        {(!rules || rules.length === 0) && (
          <div className="text-center py-12 text-muted-foreground">
            <Layers className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">No enrichment rules configured</p>
            <p className="text-xs mt-1">Create a rule to automatically enrich traces and spans</p>
          </div>
        )}
      </div>
    </div>
  );
}

function EnrichmentRuleBuilderSkeleton() {
  return (
    <div className="space-y-4">
      <div className="h-10 bg-muted animate-pulse rounded-lg" />
      <div className="h-24 bg-muted animate-pulse rounded-lg" />
      <div className="h-24 bg-muted animate-pulse rounded-lg" />
    </div>
  );
}
