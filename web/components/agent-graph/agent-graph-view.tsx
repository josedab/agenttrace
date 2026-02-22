"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import ReactFlow, {
  Node,
  Edge,
  Background,
  Controls,
  MiniMap,
  Position,
} from "reactflow";
import "reactflow/dist/style.css";
import { api } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface AgentNode {
  id: string;
  name: string;
  type: string;
  model?: string;
  tokensUsed: number;
  cost: number;
  durationMs: number;
  status: string;
}

interface AgentEdge {
  sourceId: string;
  targetId: string;
  label?: string;
  messageCount: number;
  tokenCount: number;
}

interface AgentGraph {
  traceId: string;
  agents: AgentNode[];
  edges: AgentEdge[];
  totalCost: number;
  totalDurationMs: number;
}

interface AgentGraphViewProps {
  traceId: string;
}

const nodeColors: Record<string, string> = {
  orchestrator: "#6366f1",
  worker: "#22c55e",
  tool: "#f59e0b",
};

export function AgentGraphView({ traceId }: AgentGraphViewProps) {
  const [selectedNode, setSelectedNode] = React.useState<AgentNode | null>(null);

  const { data: graph, isLoading } = useQuery<AgentGraph>({
    queryKey: ["agent-graph", traceId],
    queryFn: () => api.traces.getGraph(traceId),
    enabled: !!traceId,
  });

  if (isLoading || !graph) {
    return (
      <div className="h-96 bg-muted animate-pulse rounded-lg flex items-center justify-center">
        <span className="text-muted-foreground">Loading agent graph...</span>
      </div>
    );
  }

  if (graph.agents.length === 0) {
    return (
      <div className="h-96 border rounded-lg flex items-center justify-center">
        <span className="text-muted-foreground">
          No multi-agent structure detected in this trace
        </span>
      </div>
    );
  }

  // Convert to ReactFlow nodes/edges
  const nodes: Node[] = graph.agents.map((agent, idx) => ({
    id: agent.id,
    position: {
      x: (idx % 3) * 300 + 50,
      y: Math.floor(idx / 3) * 200 + 50,
    },
    data: {
      label: (
        <div className="text-center p-2">
          <div className="font-medium text-sm">{agent.name}</div>
          <div className="text-xs text-muted-foreground mt-1">
            {agent.model || agent.type}
          </div>
          <div className="flex gap-2 mt-2 justify-center text-xs">
            <span>${agent.cost.toFixed(4)}</span>
            <span>·</span>
            <span>{agent.tokensUsed.toLocaleString()} tok</span>
          </div>
        </div>
      ),
    },
    sourcePosition: Position.Bottom,
    targetPosition: Position.Top,
    style: {
      background: `${nodeColors[agent.type] || "#6b7280"}15`,
      border: `2px solid ${nodeColors[agent.type] || "#6b7280"}`,
      borderRadius: "8px",
      minWidth: "180px",
    },
  }));

  const edges: Edge[] = graph.edges.map((edge, idx) => ({
    id: `edge-${idx}`,
    source: edge.sourceId,
    target: edge.targetId,
    label: edge.label || `${edge.messageCount} msgs`,
    animated: true,
    style: { stroke: "#6b7280" },
    labelStyle: { fontSize: 11, fill: "#6b7280" },
  }));

  return (
    <div className="space-y-4">
      {/* Graph Summary */}
      <div className="flex items-center gap-4 text-sm">
        <Badge variant="outline">{graph.agents.length} agents</Badge>
        <Badge variant="outline">{graph.edges.length} connections</Badge>
        <Badge variant="outline">${graph.totalCost.toFixed(4)} total cost</Badge>
        <Badge variant="outline">
          {(graph.totalDurationMs / 1000).toFixed(1)}s duration
        </Badge>
      </div>

      {/* Graph Canvas */}
      <div className="h-[500px] border rounded-lg bg-card">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodeClick={(_, node) => {
            const agent = graph.agents.find((a) => a.id === node.id);
            setSelectedNode(agent || null);
          }}
          fitView
          attributionPosition="bottom-left"
        >
          <Background />
          <Controls />
          <MiniMap />
        </ReactFlow>
      </div>

      {/* Selected Node Detail */}
      {selectedNode && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2">
              <div
                className="w-3 h-3 rounded-full"
                style={{
                  backgroundColor: nodeColors[selectedNode.type] || "#6b7280",
                }}
              />
              {selectedNode.name}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-4 gap-4 text-sm">
              <div>
                <span className="text-muted-foreground">Type</span>
                <div className="font-medium">{selectedNode.type}</div>
              </div>
              <div>
                <span className="text-muted-foreground">Model</span>
                <div className="font-medium font-mono">
                  {selectedNode.model || "—"}
                </div>
              </div>
              <div>
                <span className="text-muted-foreground">Cost</span>
                <div className="font-medium">${selectedNode.cost.toFixed(4)}</div>
              </div>
              <div>
                <span className="text-muted-foreground">Tokens</span>
                <div className="font-medium">
                  {selectedNode.tokensUsed.toLocaleString()}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
