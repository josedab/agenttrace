"use client";

import { useState } from "react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Trophy,
  TrendingUp,
  DollarSign,
  Zap,
  Activity,
  Clock,
  BarChart3,
} from "lucide-react";
import {
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  Radar,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
} from "recharts";
import { cn } from "@/lib/utils";

// --- Mock Data ---

const agents = [
  { name: "Claude Code", color: "#8b5cf6", key: "claudeCode" },
  { name: "Copilot", color: "#3b82f6", key: "copilot" },
  { name: "Cursor", color: "#22c55e", key: "cursor" },
  { name: "Aider", color: "#f97316", key: "aider" },
] as const;

type AgentKey = (typeof agents)[number]["key"];

const agentBadgeClasses: Record<AgentKey, string> = {
  claudeCode: "bg-purple-500/15 text-purple-400 border-purple-500/30",
  copilot: "bg-blue-500/15 text-blue-400 border-blue-500/30",
  cursor: "bg-green-500/15 text-green-400 border-green-500/30",
  aider: "bg-orange-500/15 text-orange-400 border-orange-500/30",
};

const radarData = [
  { metric: "Speed", claudeCode: 82, copilot: 78, cursor: 91, aider: 70 },
  { metric: "Quality", claudeCode: 95, copilot: 85, cursor: 88, aider: 80 },
  { metric: "Cost Eff.", claudeCode: 68, copilot: 90, cursor: 75, aider: 92 },
  { metric: "Accuracy", claudeCode: 93, copilot: 82, cursor: 86, aider: 77 },
  { metric: "Token Eff.", claudeCode: 74, copilot: 88, cursor: 80, aider: 85 },
  { metric: "Reliability", claudeCode: 90, copilot: 84, cursor: 87, aider: 76 },
];

const costData = [
  { name: "Claude Code", cost: 0.042, fill: "#8b5cf6" },
  { name: "Copilot", cost: 0.018, fill: "#3b82f6" },
  { name: "Cursor", cost: 0.031, fill: "#22c55e" },
  { name: "Aider", cost: 0.015, fill: "#f97316" },
];

interface AgentMetrics {
  costPerTrace: string;
  avgLatency: string;
  qualityScore: number;
  tokenEfficiency: string;
  errorRate: string;
  rank: number;
}

const agentMetrics: Record<AgentKey, AgentMetrics> = {
  claudeCode: {
    costPerTrace: "$0.042",
    avgLatency: "1.8s",
    qualityScore: 95,
    tokenEfficiency: "74%",
    errorRate: "2.1%",
    rank: 1,
  },
  copilot: {
    costPerTrace: "$0.018",
    avgLatency: "2.1s",
    qualityScore: 85,
    tokenEfficiency: "88%",
    errorRate: "3.4%",
    rank: 3,
  },
  cursor: {
    costPerTrace: "$0.031",
    avgLatency: "1.2s",
    qualityScore: 88,
    tokenEfficiency: "80%",
    errorRate: "2.8%",
    rank: 2,
  },
  aider: {
    costPerTrace: "$0.015",
    avgLatency: "2.5s",
    qualityScore: 80,
    tokenEfficiency: "85%",
    errorRate: "4.2%",
    rank: 4,
  },
};

// --- Components ---

function StatCards() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium">Total Profiles</CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">4</div>
          <p className="text-xs text-muted-foreground">Active agent profiles</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium">Comparisons Run</CardTitle>
          <BarChart3 className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">128</div>
          <p className="text-xs text-muted-foreground">
            <span className="text-green-500">+12%</span> from last week
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium">Top Agent</CardTitle>
          <Trophy className="h-4 w-4 text-yellow-500" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">Claude Code</div>
          <p className="text-xs text-muted-foreground">Highest quality score</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium">Cost Savings</CardTitle>
          <DollarSign className="h-4 w-4 text-green-500" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">$1,240</div>
          <p className="text-xs text-muted-foreground">
            By optimizing agent selection
          </p>
        </CardContent>
      </Card>
    </div>
  );
}

function RadarComparisonChart() {
  return (
    <Card className="col-span-1 lg:col-span-2">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <TrendingUp className="h-5 w-5" />
          Normalized Comparison
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={380}>
          <RadarChart data={radarData} cx="50%" cy="50%" outerRadius="75%">
            <PolarGrid stroke="hsl(var(--border))" />
            <PolarAngleAxis
              dataKey="metric"
              tick={{ fill: "hsl(var(--muted-foreground))", fontSize: 12 }}
            />
            <PolarRadiusAxis
              angle={30}
              domain={[0, 100]}
              tick={{ fill: "hsl(var(--muted-foreground))", fontSize: 10 }}
            />
            {agents.map((agent) => (
              <Radar
                key={agent.key}
                name={agent.name}
                dataKey={agent.key}
                stroke={agent.color}
                fill={agent.color}
                fillOpacity={0.15}
              />
            ))}
            <Legend />
          </RadarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}

function CostComparisonChart() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <DollarSign className="h-5 w-5" />
          Cost per Trace
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={380}>
          <BarChart data={costData} layout="vertical">
            <CartesianGrid
              strokeDasharray="3 3"
              stroke="hsl(var(--border))"
            />
            <XAxis
              type="number"
              tick={{ fill: "hsl(var(--muted-foreground))", fontSize: 12 }}
              tickFormatter={(v: number) => `$${v.toFixed(3)}`}
            />
            <YAxis
              dataKey="name"
              type="category"
              width={100}
              tick={{ fill: "hsl(var(--muted-foreground))", fontSize: 12 }}
            />
            <Tooltip
              formatter={(value: number) => [`$${value.toFixed(3)}`, "Cost"]}
              contentStyle={{
                backgroundColor: "hsl(var(--card))",
                border: "1px solid hsl(var(--border))",
                borderRadius: "8px",
              }}
              labelStyle={{ color: "hsl(var(--foreground))" }}
            />
            <Bar dataKey="cost" radius={[0, 4, 4, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}

function rankBadgeClass(rank: number) {
  switch (rank) {
    case 1:
      return "bg-yellow-500/15 text-yellow-400 border-yellow-500/30";
    case 2:
      return "bg-slate-300/15 text-slate-300 border-slate-300/30";
    case 3:
      return "bg-amber-700/15 text-amber-600 border-amber-700/30";
    default:
      return "bg-muted text-muted-foreground";
  }
}

function ComparisonTable() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Zap className="h-5 w-5" />
          Detailed Comparison
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left py-3 px-4 font-medium text-muted-foreground">
                  Rank
                </th>
                <th className="text-left py-3 px-4 font-medium text-muted-foreground">
                  Agent
                </th>
                <th className="text-left py-3 px-4 font-medium text-muted-foreground">
                  <DollarSign className="inline h-3 w-3 mr-1" />
                  Cost/Trace
                </th>
                <th className="text-left py-3 px-4 font-medium text-muted-foreground">
                  <Clock className="inline h-3 w-3 mr-1" />
                  Avg Latency
                </th>
                <th className="text-left py-3 px-4 font-medium text-muted-foreground">
                  <Trophy className="inline h-3 w-3 mr-1" />
                  Quality Score
                </th>
                <th className="text-left py-3 px-4 font-medium text-muted-foreground">
                  <Zap className="inline h-3 w-3 mr-1" />
                  Token Efficiency
                </th>
                <th className="text-left py-3 px-4 font-medium text-muted-foreground">
                  <Activity className="inline h-3 w-3 mr-1" />
                  Error Rate
                </th>
              </tr>
            </thead>
            <tbody>
              {agents
                .slice()
                .sort(
                  (a, b) =>
                    agentMetrics[a.key].rank - agentMetrics[b.key].rank
                )
                .map((agent) => {
                  const m = agentMetrics[agent.key];
                  return (
                    <tr
                      key={agent.key}
                      className="border-b border-border/50 hover:bg-muted/50 transition-colors"
                    >
                      <td className="py-3 px-4">
                        <Badge
                          variant="outline"
                          className={cn("text-xs", rankBadgeClass(m.rank))}
                        >
                          #{m.rank}
                        </Badge>
                      </td>
                      <td className="py-3 px-4">
                        <Badge
                          variant="outline"
                          className={cn(
                            "text-xs",
                            agentBadgeClasses[agent.key]
                          )}
                        >
                          {agent.name}
                        </Badge>
                      </td>
                      <td className="py-3 px-4 font-mono">{m.costPerTrace}</td>
                      <td className="py-3 px-4 font-mono">{m.avgLatency}</td>
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-2">
                          <div className="h-2 flex-1 max-w-[100px] rounded-full bg-muted overflow-hidden">
                            <div
                              className="h-full rounded-full"
                              style={{
                                width: `${m.qualityScore}%`,
                                backgroundColor: agent.color,
                              }}
                            />
                          </div>
                          <span className="font-mono text-xs">
                            {m.qualityScore}
                          </span>
                        </div>
                      </td>
                      <td className="py-3 px-4 font-mono">
                        {m.tokenEfficiency}
                      </td>
                      <td className="py-3 px-4 font-mono">{m.errorRate}</td>
                    </tr>
                  );
                })}
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  );
}

function NewComparisonForm({
  onClose,
}: {
  onClose: () => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Run New Comparison</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Select Agents</label>
            <div className="space-y-2">
              {agents.map((agent) => (
                <label
                  key={agent.key}
                  className="flex items-center gap-2 text-sm"
                >
                  <input
                    type="checkbox"
                    defaultChecked
                    className="rounded border-border"
                  />
                  <Badge
                    variant="outline"
                    className={cn("text-xs", agentBadgeClasses[agent.key])}
                  >
                    {agent.name}
                  </Badge>
                </label>
              ))}
            </div>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Metrics</label>
            <div className="space-y-2">
              {["Cost", "Latency", "Quality", "Token Usage", "Error Rate"].map(
                (metric) => (
                  <label
                    key={metric}
                    className="flex items-center gap-2 text-sm"
                  >
                    <input
                      type="checkbox"
                      defaultChecked
                      className="rounded border-border"
                    />
                    {metric}
                  </label>
                )
              )}
            </div>
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={onClose}>Run Comparison</Button>
        </div>
      </CardContent>
    </Card>
  );
}

// --- Main Dashboard ---

export function AgentComparisonDashboard() {
  const [showForm, setShowForm] = useState(false);

  return (
    <div className="space-y-6">
      <StatCards />

      <div className="flex justify-end">
        <Button onClick={() => setShowForm(!showForm)}>
          <BarChart3 className="h-4 w-4 mr-2" />
          Run New Comparison
        </Button>
      </div>

      {showForm && <NewComparisonForm onClose={() => setShowForm(false)} />}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <RadarComparisonChart />
        <CostComparisonChart />
      </div>

      <ComparisonTable />
    </div>
  );
}
