"use client";

import { useState } from "react";
import {
  DollarSign,
  TrendingUp,
  TrendingDown,
  Minus,
  AlertTriangle,
  Zap,
  Shield,
  Plus,
} from "lucide-react";

interface CostHotspot {
  id: string;
  category: string;
  name: string;
  totalCostUsd: number;
  traceCount: number;
  avgCostPerTrace: number;
  trend: string;
  trendPercent: number;
  topModels?: { model: string; costUsd: number; traceCount: number; percentage: number }[];
}

interface CostPrediction {
  date: string;
  predictedCost: number;
  lowerBound: number;
  upperBound: number;
  confidence: number;
  budgetRemaining: number;
  overrunRisk: number;
}

interface CostAutopilotRule {
  id: string;
  name: string;
  ruleType: string;
  enabled: boolean;
  executionCount: number;
  lastExecuted?: string;
}

interface CostRecommendation {
  id: string;
  currentModel: string;
  recommendedModel: string;
  estimatedSavingsPerMonth: number;
  confidence: number;
  status: string;
}

interface DashboardData {
  currentMonthCost: number;
  monthlyBudget: number;
  budgetUtilization: number;
  projectedOverrun: number;
  hotspots: CostHotspot[];
  predictions: CostPrediction[];
  activeRules: number;
  savingsThisMonth: number;
  recommendations: CostRecommendation[];
}

const TrendIcon = ({ trend }: { trend: string }) => {
  if (trend === "increasing") return <TrendingUp className="h-4 w-4 text-red-500" />;
  if (trend === "decreasing") return <TrendingDown className="h-4 w-4 text-green-500" />;
  return <Minus className="h-4 w-4 text-muted-foreground" />;
};

function RuleTypeBadge({ type }: { type: string }) {
  const colors: Record<string, string> = {
    model_downgrade: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
    cache_enable: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
    rate_limit: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
    budget_alert: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
  };
  return (
    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${colors[type] || "bg-muted text-muted-foreground"}`}>
      {type.replace(/_/g, " ")}
    </span>
  );
}

export function CostAutopilotDashboard() {
  const [dashboard] = useState<DashboardData>({
    currentMonthCost: 325.40,
    monthlyBudget: 500.00,
    budgetUtilization: 65.08,
    projectedOverrun: 0,
    hotspots: [
      {
        id: "1",
        category: "model",
        name: "GPT-4 usage in simple tasks",
        totalCostUsd: 245.80,
        traceCount: 1250,
        avgCostPerTrace: 0.197,
        trend: "increasing",
        trendPercent: 12.5,
        topModels: [
          { model: "gpt-4", costUsd: 200.0, traceCount: 800, percentage: 81.4 },
          { model: "gpt-3.5-turbo", costUsd: 45.80, traceCount: 450, percentage: 18.6 },
        ],
      },
      {
        id: "2",
        category: "prompt",
        name: "Verbose system prompts",
        totalCostUsd: 89.50,
        traceCount: 500,
        avgCostPerTrace: 0.179,
        trend: "stable",
        trendPercent: 1.2,
      },
    ],
    predictions: [],
    activeRules: 2,
    savingsThisMonth: 42.50,
    recommendations: [
      {
        id: "r1",
        currentModel: "gpt-4",
        recommendedModel: "gpt-4o-mini",
        estimatedSavingsPerMonth: 160.0,
        confidence: 0.85,
        status: "PENDING",
      },
    ],
  });

  const [rules] = useState<CostAutopilotRule[]>([
    {
      id: "rule-1",
      name: "Downgrade simple classification tasks",
      ruleType: "model_downgrade",
      enabled: true,
      executionCount: 45,
      lastExecuted: new Date().toISOString(),
    },
    {
      id: "rule-2",
      name: "Alert when daily cost exceeds $25",
      ruleType: "budget_alert",
      enabled: true,
      executionCount: 3,
    },
  ]);

  const [showNewRule, setShowNewRule] = useState(false);

  const budgetPct = dashboard.budgetUtilization;
  const budgetColor =
    budgetPct >= 90 ? "bg-red-500" : budgetPct >= 70 ? "bg-yellow-500" : "bg-green-500";

  return (
    <div className="space-y-6">
      {/* Budget Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="rounded-lg border bg-card p-4">
          <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
            <DollarSign className="h-4 w-4" />
            Current Month Spend
          </div>
          <div className="text-2xl font-bold">${dashboard.currentMonthCost.toFixed(2)}</div>
          <div className="mt-2 h-2 rounded-full bg-muted overflow-hidden">
            <div className={`h-full rounded-full ${budgetColor}`} style={{ width: `${Math.min(budgetPct, 100)}%` }} />
          </div>
          <p className="text-xs text-muted-foreground mt-1">{budgetPct.toFixed(1)}% of budget</p>
        </div>

        <div className="rounded-lg border bg-card p-4">
          <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
            <Shield className="h-4 w-4" />
            Monthly Budget
          </div>
          <div className="text-2xl font-bold">${dashboard.monthlyBudget.toFixed(2)}</div>
          <p className="text-xs text-muted-foreground mt-1">
            ${(dashboard.monthlyBudget - dashboard.currentMonthCost).toFixed(2)} remaining
          </p>
        </div>

        <div className="rounded-lg border bg-card p-4">
          <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
            <Zap className="h-4 w-4" />
            Savings This Month
          </div>
          <div className="text-2xl font-bold text-green-600">${dashboard.savingsThisMonth.toFixed(2)}</div>
          <p className="text-xs text-muted-foreground mt-1">{dashboard.activeRules} active rules</p>
        </div>

        <div className="rounded-lg border bg-card p-4">
          <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
            <AlertTriangle className="h-4 w-4" />
            Projected Overrun
          </div>
          <div className={`text-2xl font-bold ${dashboard.projectedOverrun > 0 ? "text-red-600" : "text-green-600"}`}>
            {dashboard.projectedOverrun > 0 ? `+$${dashboard.projectedOverrun.toFixed(2)}` : "On Track"}
          </div>
          <p className="text-xs text-muted-foreground mt-1">Based on current trends</p>
        </div>
      </div>

      {/* Cost Hotspots */}
      <div className="rounded-lg border bg-card">
        <div className="border-b px-4 py-3">
          <h3 className="font-semibold">Cost Hotspots</h3>
          <p className="text-sm text-muted-foreground">Areas with highest cost impact</p>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-4 py-2 text-left font-medium">Category</th>
                <th className="px-4 py-2 text-left font-medium">Name</th>
                <th className="px-4 py-2 text-right font-medium">Total Cost</th>
                <th className="px-4 py-2 text-right font-medium">Traces</th>
                <th className="px-4 py-2 text-right font-medium">Avg/Trace</th>
                <th className="px-4 py-2 text-center font-medium">Trend</th>
              </tr>
            </thead>
            <tbody>
              {dashboard.hotspots.map((hotspot) => (
                <tr key={hotspot.id} className="border-b last:border-0 hover:bg-muted/30">
                  <td className="px-4 py-3">
                    <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-muted">
                      {hotspot.category}
                    </span>
                  </td>
                  <td className="px-4 py-3 font-medium">{hotspot.name}</td>
                  <td className="px-4 py-3 text-right font-mono">${hotspot.totalCostUsd.toFixed(2)}</td>
                  <td className="px-4 py-3 text-right">{hotspot.traceCount.toLocaleString()}</td>
                  <td className="px-4 py-3 text-right font-mono">${hotspot.avgCostPerTrace.toFixed(3)}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-center gap-1">
                      <TrendIcon trend={hotspot.trend} />
                      <span className={`text-xs ${hotspot.trend === "increasing" ? "text-red-500" : hotspot.trend === "decreasing" ? "text-green-500" : "text-muted-foreground"}`}>
                        {hotspot.trendPercent > 0 ? "+" : ""}{hotspot.trendPercent}%
                      </span>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Model Downgrade Recommendations */}
      {dashboard.recommendations.length > 0 && (
        <div className="rounded-lg border bg-card">
          <div className="border-b px-4 py-3">
            <h3 className="font-semibold">Model Downgrade Recommendations</h3>
            <p className="text-sm text-muted-foreground">Switch to cheaper models for simple tasks</p>
          </div>
          <div className="divide-y">
            {dashboard.recommendations.map((rec) => (
              <div key={rec.id} className="px-4 py-3 flex items-center justify-between">
                <div>
                  <div className="flex items-center gap-2 text-sm">
                    <code className="px-1.5 py-0.5 rounded bg-muted text-xs">{rec.currentModel}</code>
                    <span className="text-muted-foreground">→</span>
                    <code className="px-1.5 py-0.5 rounded bg-green-100 dark:bg-green-900 text-xs">{rec.recommendedModel}</code>
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">
                    {(rec.confidence * 100).toFixed(0)}% confidence · {rec.status.toLowerCase()}
                  </p>
                </div>
                <div className="text-right">
                  <div className="text-sm font-semibold text-green-600">
                    -${rec.estimatedSavingsPerMonth.toFixed(2)}/mo
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Autopilot Rules */}
      <div className="rounded-lg border bg-card">
        <div className="border-b px-4 py-3 flex items-center justify-between">
          <div>
            <h3 className="font-semibold">Autopilot Rules</h3>
            <p className="text-sm text-muted-foreground">Automated cost optimization rules</p>
          </div>
          <button
            onClick={() => setShowNewRule(!showNewRule)}
            className="inline-flex items-center gap-1 px-3 py-1.5 rounded-md text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90"
          >
            <Plus className="h-3.5 w-3.5" />
            Add Rule
          </button>
        </div>

        {showNewRule && (
          <div className="border-b px-4 py-4 bg-muted/30">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium">Rule Name</label>
                <input
                  type="text"
                  placeholder="e.g., Downgrade GPT-4 for simple tasks"
                  className="mt-1 w-full rounded-md border bg-background px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="text-sm font-medium">Rule Type</label>
                <select className="mt-1 w-full rounded-md border bg-background px-3 py-2 text-sm">
                  <option value="model_downgrade">Model Downgrade</option>
                  <option value="cache_enable">Enable Caching</option>
                  <option value="rate_limit">Rate Limit</option>
                  <option value="budget_alert">Budget Alert</option>
                </select>
              </div>
              <div>
                <label className="text-sm font-medium">Trigger Metric</label>
                <select className="mt-1 w-full rounded-md border bg-background px-3 py-2 text-sm">
                  <option value="daily_cost">Daily Cost</option>
                  <option value="per_trace_cost">Per-Trace Cost</option>
                  <option value="token_usage">Token Usage</option>
                  <option value="error_rate">Error Rate</option>
                </select>
              </div>
              <div>
                <label className="text-sm font-medium">Threshold ($)</label>
                <input
                  type="number"
                  placeholder="25.00"
                  className="mt-1 w-full rounded-md border bg-background px-3 py-2 text-sm"
                />
              </div>
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <button
                onClick={() => setShowNewRule(false)}
                className="px-3 py-1.5 rounded-md text-sm border hover:bg-muted"
              >
                Cancel
              </button>
              <button
                onClick={() => setShowNewRule(false)}
                className="px-3 py-1.5 rounded-md text-sm bg-primary text-primary-foreground hover:bg-primary/90"
              >
                Create Rule
              </button>
            </div>
          </div>
        )}

        <div className="divide-y">
          {rules.map((rule) => (
            <div key={rule.id} className="px-4 py-3 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className={`h-2 w-2 rounded-full ${rule.enabled ? "bg-green-500" : "bg-muted-foreground"}`} />
                <div>
                  <div className="text-sm font-medium">{rule.name}</div>
                  <div className="flex items-center gap-2 mt-1">
                    <RuleTypeBadge type={rule.ruleType} />
                    <span className="text-xs text-muted-foreground">
                      {rule.executionCount} executions
                    </span>
                  </div>
                </div>
              </div>
              <button className="text-xs px-2 py-1 rounded border hover:bg-muted">
                {rule.enabled ? "Disable" : "Enable"}
              </button>
            </div>
          ))}
          {rules.length === 0 && (
            <div className="px-4 py-8 text-center text-sm text-muted-foreground">
              No autopilot rules configured. Add a rule to start automating cost optimizations.
            </div>
          )}
        </div>
      </div>

      {/* Savings Summary */}
      <div className="rounded-lg border bg-card p-4">
        <h3 className="font-semibold mb-3">Savings Summary</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="text-center p-3 rounded-lg bg-green-50 dark:bg-green-950">
            <div className="text-2xl font-bold text-green-600">${dashboard.savingsThisMonth.toFixed(2)}</div>
            <div className="text-xs text-muted-foreground mt-1">Saved This Month</div>
          </div>
          <div className="text-center p-3 rounded-lg bg-blue-50 dark:bg-blue-950">
            <div className="text-2xl font-bold text-blue-600">
              ${dashboard.recommendations.reduce((sum, r) => sum + r.estimatedSavingsPerMonth, 0).toFixed(2)}
            </div>
            <div className="text-xs text-muted-foreground mt-1">Potential Monthly Savings</div>
          </div>
          <div className="text-center p-3 rounded-lg bg-purple-50 dark:bg-purple-950">
            <div className="text-2xl font-bold text-purple-600">{dashboard.activeRules}</div>
            <div className="text-xs text-muted-foreground mt-1">Active Optimization Rules</div>
          </div>
        </div>
      </div>
    </div>
  );
}
