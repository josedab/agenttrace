"use client";

import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

// --- Types ---

interface AgentBottleneck {
  agentId: string;
  agentName: string;
  bottleneckType: string;
  severity: string;
  avgLatencyMs: number;
  messageCount: number;
  errorCount: number;
  suggestion: string;
}

interface MessageFlowEdge {
  sourceAgent: string;
  targetAgent: string;
  messageCount: number;
  avgLatencyMs: number;
  errorRate: number;
}

interface TopologyAnalytics {
  sessionId: string;
  topologyType: string;
  totalAgents: number;
  totalMessages: number;
  totalHandoffs: number;
  avgResponseTimeMs: number;
  criticalPath: string[];
  bottlenecks: AgentBottleneck[];
  messageFlow: MessageFlowEdge[];
  healthScore: number;
}

interface DelegationStep {
  fromAgent: string;
  toAgent: string;
  task: string;
  durationMs: number;
  status: string;
  timestamp: string;
}

interface DelegationChain {
  id: string;
  sessionId: string;
  initiatorId: string;
  steps: DelegationStep[];
  totalTimeMs: number;
  status: string;
  createdAt: string;
}

// --- Mock data (matches API stub output) ---

const MOCK_ANALYTICS: TopologyAnalytics = {
  sessionId: "demo-session",
  topologyType: "hub_spoke",
  totalAgents: 4,
  totalMessages: 28,
  totalHandoffs: 6,
  avgResponseTimeMs: 1250,
  criticalPath: ["orchestrator", "planner", "executor", "reviewer"],
  bottlenecks: [
    {
      agentId: "executor-1",
      agentName: "Code Executor",
      bottleneckType: "high_latency",
      severity: "medium",
      avgLatencyMs: 3200,
      messageCount: 12,
      errorCount: 1,
      suggestion:
        "Consider splitting complex tasks across multiple executor agents",
    },
  ],
  messageFlow: [
    { sourceAgent: "orchestrator", targetAgent: "planner", messageCount: 8, avgLatencyMs: 450, errorRate: 0.0 },
    { sourceAgent: "planner", targetAgent: "executor-1", messageCount: 10, avgLatencyMs: 3200, errorRate: 0.08 },
    { sourceAgent: "executor-1", targetAgent: "reviewer", messageCount: 6, avgLatencyMs: 800, errorRate: 0.0 },
    { sourceAgent: "reviewer", targetAgent: "orchestrator", messageCount: 4, avgLatencyMs: 600, errorRate: 0.0 },
  ],
  healthScore: 78.5,
};

const MOCK_CHAINS: DelegationChain[] = [
  {
    id: "chain-1",
    sessionId: "demo-session",
    initiatorId: "orchestrator",
    steps: [
      { fromAgent: "orchestrator", toAgent: "planner", task: "Plan implementation", durationMs: 2000, status: "completed", timestamp: new Date(Date.now() - 5 * 60000).toISOString() },
      { fromAgent: "planner", toAgent: "executor-1", task: "Execute code changes", durationMs: 8000, status: "completed", timestamp: new Date(Date.now() - 3 * 60000).toISOString() },
      { fromAgent: "executor-1", toAgent: "reviewer", task: "Review changes", durationMs: 3000, status: "completed", timestamp: new Date(Date.now() - 1 * 60000).toISOString() },
    ],
    totalTimeMs: 13000,
    status: "completed",
    createdAt: new Date(Date.now() - 5 * 60000).toISOString(),
  },
];

// --- Agent graph node layout ---

interface AgentNodeLayout {
  id: string;
  label: string;
  role: string;
  x: number;
  y: number;
  latencyMs: number;
  messageCount: number;
  isCritical: boolean;
  bottleneck?: AgentBottleneck;
}

function buildNodeLayout(analytics: TopologyAnalytics): AgentNodeLayout[] {
  const agents = new Set<string>();
  analytics.messageFlow.forEach((e) => {
    agents.add(e.sourceAgent);
    agents.add(e.targetAgent);
  });

  const agentList = Array.from(agents);
  const centerX = 300;
  const centerY = 220;
  const radius = 160;

  const msgCountByAgent: Record<string, number> = {};
  const latencyByAgent: Record<string, number[]> = {};

  analytics.messageFlow.forEach((e) => {
    msgCountByAgent[e.sourceAgent] = (msgCountByAgent[e.sourceAgent] || 0) + e.messageCount;
    msgCountByAgent[e.targetAgent] = (msgCountByAgent[e.targetAgent] || 0) + e.messageCount;
    if (!latencyByAgent[e.sourceAgent]) latencyByAgent[e.sourceAgent] = [];
    latencyByAgent[e.sourceAgent].push(e.avgLatencyMs);
  });

  const bottleneckMap: Record<string, AgentBottleneck> = {};
  analytics.bottlenecks.forEach((b) => {
    bottleneckMap[b.agentId] = b;
  });

  const criticalSet = new Set(analytics.criticalPath);

  return agentList.map((id, idx) => {
    const angle = (2 * Math.PI * idx) / agentList.length - Math.PI / 2;
    const lats = latencyByAgent[id] || [];
    const avgLat = lats.length > 0 ? lats.reduce((a, b) => a + b, 0) / lats.length : 0;

    return {
      id,
      label: id.charAt(0).toUpperCase() + id.slice(1).replace(/-/g, " "),
      role: id.includes("orchestrator") ? "coordinator" : id.includes("planner") ? "planner" : id.includes("executor") ? "executor" : "reviewer",
      x: centerX + radius * Math.cos(angle),
      y: centerY + radius * Math.sin(angle),
      latencyMs: avgLat,
      messageCount: msgCountByAgent[id] || 0,
      isCritical: criticalSet.has(id),
      bottleneck: bottleneckMap[id],
    };
  });
}

// --- SVG Topology Graph ---

const ROLE_COLORS: Record<string, string> = {
  coordinator: "#6366f1",
  planner: "#3b82f6",
  executor: "#22c55e",
  reviewer: "#f59e0b",
};

const SEVERITY_BORDER: Record<string, string> = {
  critical: "#ef4444",
  high: "#f97316",
  medium: "#eab308",
  low: "#a3a3a3",
};

function TopologyGraph({
  analytics,
  selectedAgent,
  onSelectAgent,
}: {
  analytics: TopologyAnalytics;
  selectedAgent: string | null;
  onSelectAgent: (id: string | null) => void;
}) {
  const nodes = buildNodeLayout(analytics);
  const nodeMap: Record<string, AgentNodeLayout> = {};
  nodes.forEach((n) => (nodeMap[n.id] = n));

  return (
    <svg
      viewBox="0 0 600 440"
      className="w-full h-full"
      style={{ minHeight: 400 }}
    >
      <defs>
        <marker id="arrowhead" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto">
          <polygon points="0 0, 8 3, 0 6" fill="#6b7280" />
        </marker>
        <marker id="arrowhead-error" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto">
          <polygon points="0 0, 8 3, 0 6" fill="#ef4444" />
        </marker>
      </defs>

      {/* Edges */}
      {analytics.messageFlow.map((edge, idx) => {
        const src = nodeMap[edge.sourceAgent];
        const tgt = nodeMap[edge.targetAgent];
        if (!src || !tgt) return null;

        const dx = tgt.x - src.x;
        const dy = tgt.y - src.y;
        const dist = Math.sqrt(dx * dx + dy * dy);
        const nodeRadius = 36;
        const sx = src.x + (dx / dist) * nodeRadius;
        const sy = src.y + (dy / dist) * nodeRadius;
        const ex = tgt.x - (dx / dist) * nodeRadius;
        const ey = tgt.y - (dy / dist) * nodeRadius;

        const hasError = edge.errorRate > 0;
        const strokeColor = hasError ? "#ef4444" : "#94a3b8";
        const strokeWidth = Math.max(1.5, Math.min(4, edge.messageCount / 3));
        const midX = (sx + ex) / 2;
        const midY = (sy + ey) / 2;

        return (
          <g key={`edge-${idx}`}>
            <line
              x1={sx} y1={sy} x2={ex} y2={ey}
              stroke={strokeColor}
              strokeWidth={strokeWidth}
              strokeDasharray={hasError ? "6 3" : undefined}
              markerEnd={hasError ? "url(#arrowhead-error)" : "url(#arrowhead)"}
              opacity={0.7}
            />
            {/* Animated dot along edge */}
            <circle r="3" fill={strokeColor}>
              <animateMotion
                dur={`${2 + idx * 0.5}s`}
                repeatCount="indefinite"
                path={`M${sx},${sy} L${ex},${ey}`}
              />
            </circle>
            {/* Edge label */}
            <text x={midX} y={midY - 6} textAnchor="middle" fontSize="10" fill="#64748b">
              {edge.messageCount} msgs
            </text>
            <text x={midX} y={midY + 6} textAnchor="middle" fontSize="9" fill="#94a3b8">
              {edge.avgLatencyMs.toFixed(0)}ms
            </text>
          </g>
        );
      })}

      {/* Nodes */}
      {nodes.map((node) => {
        const color = ROLE_COLORS[node.role] || "#6b7280";
        const isSelected = selectedAgent === node.id;
        const borderColor = node.bottleneck
          ? SEVERITY_BORDER[node.bottleneck.severity] || color
          : color;
        const borderWidth = node.bottleneck ? 3 : isSelected ? 3 : 2;

        return (
          <g
            key={node.id}
            onClick={() => onSelectAgent(isSelected ? null : node.id)}
            style={{ cursor: "pointer" }}
          >
            {/* Outer glow for bottleneck */}
            {node.bottleneck && (
              <circle
                cx={node.x} cy={node.y} r={42}
                fill="none"
                stroke={SEVERITY_BORDER[node.bottleneck.severity]}
                strokeWidth={1}
                strokeDasharray="4 2"
                opacity={0.5}
              />
            )}
            {/* Node circle */}
            <circle
              cx={node.x} cy={node.y} r={36}
              fill={`${color}15`}
              stroke={borderColor}
              strokeWidth={borderWidth}
            />
            {/* Status dot */}
            <circle cx={node.x + 26} cy={node.y - 26} r={5} fill="#22c55e" stroke="white" strokeWidth={1.5} />
            {/* Label */}
            <text x={node.x} y={node.y - 6} textAnchor="middle" fontSize="11" fontWeight="600" fill="currentColor">
              {node.label}
            </text>
            <text x={node.x} y={node.y + 8} textAnchor="middle" fontSize="9" fill="#64748b">
              {node.messageCount} msgs
            </text>
            <text x={node.x} y={node.y + 20} textAnchor="middle" fontSize="9" fill="#94a3b8">
              {node.latencyMs.toFixed(0)}ms avg
            </text>
            {/* Critical path indicator */}
            {node.isCritical && (
              <text x={node.x} y={node.y + 48} textAnchor="middle" fontSize="9" fill="#6366f1" fontWeight="500">
                ● critical path
              </text>
            )}
          </g>
        );
      })}
    </svg>
  );
}

// --- Delegation Chain Visualization ---

function DelegationChainView({ chains }: { chains: DelegationChain[] }) {
  if (chains.length === 0) {
    return <p className="text-sm text-muted-foreground">No delegation chains found.</p>;
  }

  return (
    <div className="space-y-4">
      {chains.map((chain) => (
        <Card key={chain.id}>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm flex items-center gap-2">
              Delegation Chain
              <Badge variant={chain.status === "completed" ? "default" : chain.status === "active" ? "secondary" : "destructive"}>
                {chain.status}
              </Badge>
              <span className="text-xs text-muted-foreground ml-auto">
                {(chain.totalTimeMs / 1000).toFixed(1)}s total
              </span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-1 flex-wrap">
              {chain.steps.map((step, idx) => (
                <React.Fragment key={idx}>
                  <div className="flex flex-col items-center px-2 py-1 rounded border bg-muted/50 min-w-[100px]">
                    <span className="text-xs font-medium">{step.fromAgent}</span>
                    <span className="text-[10px] text-muted-foreground mt-0.5">{step.task}</span>
                    <span className="text-[10px] text-muted-foreground">{(step.durationMs / 1000).toFixed(1)}s</span>
                  </div>
                  {idx < chain.steps.length - 1 && (
                    <svg width="24" height="16" viewBox="0 0 24 16" className="text-muted-foreground shrink-0">
                      <line x1="0" y1="8" x2="18" y2="8" stroke="currentColor" strokeWidth="1.5" />
                      <polygon points="18,4 24,8 18,12" fill="currentColor" />
                    </svg>
                  )}
                  {idx === chain.steps.length - 1 && (
                    <>
                      <svg width="24" height="16" viewBox="0 0 24 16" className="text-muted-foreground shrink-0">
                        <line x1="0" y1="8" x2="18" y2="8" stroke="currentColor" strokeWidth="1.5" />
                        <polygon points="18,4 24,8 18,12" fill="currentColor" />
                      </svg>
                      <div className="flex flex-col items-center px-2 py-1 rounded border bg-muted/50 min-w-[100px]">
                        <span className="text-xs font-medium">{step.toAgent}</span>
                        <span className="text-[10px] text-muted-foreground mt-0.5">final</span>
                      </div>
                    </>
                  )}
                </React.Fragment>
              ))}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

// --- Health Score Gauge ---

function HealthScoreGauge({ score }: { score: number }) {
  const color = score >= 80 ? "#22c55e" : score >= 60 ? "#eab308" : "#ef4444";
  const circumference = 2 * Math.PI * 40;
  const offset = circumference - (score / 100) * circumference;

  return (
    <div className="flex flex-col items-center">
      <svg width="100" height="100" viewBox="0 0 100 100">
        <circle cx="50" cy="50" r="40" fill="none" stroke="#e5e7eb" strokeWidth="8" />
        <circle
          cx="50" cy="50" r="40" fill="none" stroke={color} strokeWidth="8"
          strokeDasharray={circumference} strokeDashoffset={offset}
          strokeLinecap="round" transform="rotate(-90 50 50)"
        />
        <text x="50" y="46" textAnchor="middle" fontSize="20" fontWeight="700" fill="currentColor">
          {score.toFixed(0)}
        </text>
        <text x="50" y="62" textAnchor="middle" fontSize="10" fill="#64748b">
          Health
        </text>
      </svg>
    </div>
  );
}

// --- Main Dashboard ---

type TabType = "topology" | "delegations";

export function TopologyDashboard() {
  const [activeTab, setActiveTab] = React.useState<TabType>("topology");
  const [selectedAgent, setSelectedAgent] = React.useState<string | null>(null);
  const [messageTypeFilter, setMessageTypeFilter] = React.useState<string>("all");

  const analytics = MOCK_ANALYTICS;
  const chains = MOCK_CHAINS;

  const filteredFlow =
    messageTypeFilter === "all"
      ? analytics.messageFlow
      : analytics.messageFlow.filter((e) =>
          messageTypeFilter === "errors" ? e.errorRate > 0 : true
        );

  const filteredAnalytics = { ...analytics, messageFlow: filteredFlow };

  const selectedBottleneck = analytics.bottlenecks.find(
    (b) => b.agentId === selectedAgent
  );

  return (
    <div className="space-y-6">
      {/* Top stats */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        <Card>
          <CardContent className="pt-4 pb-3 text-center">
            <div className="text-2xl font-bold">{analytics.totalAgents}</div>
            <div className="text-xs text-muted-foreground">Agents</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4 pb-3 text-center">
            <div className="text-2xl font-bold">{analytics.totalMessages}</div>
            <div className="text-xs text-muted-foreground">Messages</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4 pb-3 text-center">
            <div className="text-2xl font-bold">{analytics.totalHandoffs}</div>
            <div className="text-xs text-muted-foreground">Handoffs</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4 pb-3 text-center">
            <div className="text-2xl font-bold">{analytics.avgResponseTimeMs.toFixed(0)}ms</div>
            <div className="text-xs text-muted-foreground">Avg Response</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-4 pb-3">
            <HealthScoreGauge score={analytics.healthScore} />
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b pb-2">
        <Button
          variant={activeTab === "topology" ? "default" : "ghost"}
          size="sm"
          onClick={() => setActiveTab("topology")}
        >
          Topology View
        </Button>
        <Button
          variant={activeTab === "delegations" ? "default" : "ghost"}
          size="sm"
          onClick={() => setActiveTab("delegations")}
        >
          Delegation Chains
        </Button>
      </div>

      {activeTab === "topology" && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Graph area */}
          <div className="lg:col-span-2">
            <Card>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">Agent Topology</CardTitle>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline">{analytics.topologyType.replace("_", " ")}</Badge>
                    <select
                      className="text-xs border rounded px-2 py-1 bg-background"
                      value={messageTypeFilter}
                      onChange={(e) => setMessageTypeFilter(e.target.value)}
                    >
                      <option value="all">All Messages</option>
                      <option value="errors">With Errors</option>
                    </select>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <TopologyGraph
                  analytics={filteredAnalytics}
                  selectedAgent={selectedAgent}
                  onSelectAgent={setSelectedAgent}
                />
              </CardContent>
            </Card>
          </div>

          {/* Analytics sidebar */}
          <div className="space-y-4">
            {/* Critical Path */}
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">Critical Path</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-1 flex-wrap">
                  {analytics.criticalPath.map((agent, idx) => (
                    <React.Fragment key={agent}>
                      <Badge
                        variant={selectedAgent === agent ? "default" : "outline"}
                        className="cursor-pointer text-xs"
                        onClick={() => setSelectedAgent(selectedAgent === agent ? null : agent)}
                      >
                        {agent}
                      </Badge>
                      {idx < analytics.criticalPath.length - 1 && (
                        <span className="text-muted-foreground text-xs">→</span>
                      )}
                    </React.Fragment>
                  ))}
                </div>
              </CardContent>
            </Card>

            {/* Bottlenecks */}
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">Bottlenecks</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {analytics.bottlenecks.length === 0 ? (
                  <p className="text-xs text-muted-foreground">No bottlenecks detected</p>
                ) : (
                  analytics.bottlenecks.map((b) => (
                    <div
                      key={b.agentId}
                      className={`p-2 rounded border text-xs space-y-1 cursor-pointer ${
                        selectedAgent === b.agentId ? "ring-2 ring-primary" : ""
                      }`}
                      onClick={() => setSelectedAgent(selectedAgent === b.agentId ? null : b.agentId)}
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-medium">{b.agentName}</span>
                        <Badge
                          variant={b.severity === "critical" || b.severity === "high" ? "destructive" : "secondary"}
                          className="text-[10px]"
                        >
                          {b.severity}
                        </Badge>
                      </div>
                      <div className="text-muted-foreground">{b.bottleneckType.replace("_", " ")}</div>
                      <div className="flex gap-3 text-muted-foreground">
                        <span>{b.avgLatencyMs.toFixed(0)}ms</span>
                        <span>{b.messageCount} msgs</span>
                        <span>{b.errorCount} errors</span>
                      </div>
                      <div className="text-muted-foreground italic">{b.suggestion}</div>
                    </div>
                  ))
                )}
              </CardContent>
            </Card>

            {/* Selected agent detail */}
            {selectedAgent && (
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm">Agent: {selectedAgent}</CardTitle>
                </CardHeader>
                <CardContent className="text-xs space-y-2">
                  {selectedBottleneck ? (
                    <>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Type</span>
                        <span>{selectedBottleneck.bottleneckType}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Avg Latency</span>
                        <span>{selectedBottleneck.avgLatencyMs}ms</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Messages</span>
                        <span>{selectedBottleneck.messageCount}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Errors</span>
                        <span>{selectedBottleneck.errorCount}</span>
                      </div>
                    </>
                  ) : (
                    <p className="text-muted-foreground">No issues detected for this agent.</p>
                  )}
                  {/* Message flows involving this agent */}
                  <div className="pt-2 border-t">
                    <span className="font-medium">Message Flows</span>
                    {analytics.messageFlow
                      .filter((e) => e.sourceAgent === selectedAgent || e.targetAgent === selectedAgent)
                      .map((e, i) => (
                        <div key={i} className="flex justify-between text-muted-foreground mt-1">
                          <span>{e.sourceAgent} → {e.targetAgent}</span>
                          <span>{e.messageCount} msgs / {e.avgLatencyMs}ms</span>
                        </div>
                      ))}
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      )}

      {activeTab === "delegations" && (
        <DelegationChainView chains={chains} />
      )}
    </div>
  );
}
