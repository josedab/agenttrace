import React from "react";

export interface AgentGraphProps {
  /** Nodes in the agent graph */
  nodes: GraphNode[];
  /** Edges connecting nodes */
  edges: GraphEdge[];
  /** Theme */
  theme?: "light" | "dark";
  /** Graph width */
  width?: number;
  /** Graph height */
  height?: number;
  /** Callback when a node is clicked */
  onNodeClick?: (nodeId: string) => void;
  /** Custom class name */
  className?: string;
}

export interface GraphNode {
  id: string;
  label: string;
  type: "agent" | "tool" | "model" | "input" | "output";
  x?: number;
  y?: number;
  metadata?: Record<string, unknown>;
}

export interface GraphEdge {
  source: string;
  target: string;
  label?: string;
}

const NODE_COLORS: Record<string, string> = {
  agent: "#3b82f6",
  tool: "#10b981",
  model: "#8b5cf6",
  input: "#f59e0b",
  output: "#ef4444",
};

const NODE_ICONS: Record<string, string> = {
  agent: "🤖",
  tool: "🔧",
  model: "🧠",
  input: "📥",
  output: "📤",
};

export function AgentGraph({
  nodes,
  edges,
  theme = "light",
  width = 800,
  height = 500,
  onNodeClick,
  className = "",
}: AgentGraphProps) {
  const isDark = theme === "dark";
  const bg = isDark ? "#1a1a2e" : "#ffffff";
  const border = isDark ? "#2d2d44" : "#e5e7eb";
  const text = isDark ? "#e5e7eb" : "#1f2937";
  const muted = isDark ? "#9ca3af" : "#6b7280";

  // Simple auto-layout if positions not provided
  const positionedNodes = nodes.map((node, i) => ({
    ...node,
    x: node.x ?? 100 + (i % 4) * 180,
    y: node.y ?? 80 + Math.floor(i / 4) * 120,
  }));

  const nodeMap = new Map(positionedNodes.map((n) => [n.id, n]));

  if (nodes.length === 0) {
    return (
      <div
        style={{
          width,
          height,
          backgroundColor: bg,
          border: `1px solid ${border}`,
          borderRadius: "8px",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          color: muted,
          fontFamily: 'sans-serif',
        }}
        className={className}
      >
        No graph data available
      </div>
    );
  }

  return (
    <div
      style={{
        width,
        backgroundColor: bg,
        border: `1px solid ${border}`,
        borderRadius: "8px",
        overflow: "hidden",
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
      }}
      className={className}
    >
      <div
        style={{
          padding: "12px 16px",
          borderBottom: `1px solid ${border}`,
          fontWeight: 600,
          fontSize: "14px",
          color: text,
        }}
      >
        Agent Graph
      </div>

      <svg width={width} height={height - 50} viewBox={`0 0 ${width} ${height - 50}`}>
        {/* Edges */}
        {edges.map((edge, i) => {
          const source = nodeMap.get(edge.source);
          const target = nodeMap.get(edge.target);
          if (!source || !target) return null;

          const sx = (source.x ?? 0) + 60;
          const sy = (source.y ?? 0) + 25;
          const tx = (target.x ?? 0) + 60;
          const ty = (target.y ?? 0) + 25;

          return (
            <g key={`edge-${i}`}>
              <line
                x1={sx}
                y1={sy}
                x2={tx}
                y2={ty}
                stroke={isDark ? "#4b5563" : "#d1d5db"}
                strokeWidth={2}
                markerEnd="url(#arrowhead)"
              />
              {edge.label && (
                <text
                  x={(sx + tx) / 2}
                  y={(sy + ty) / 2 - 8}
                  textAnchor="middle"
                  fontSize="10"
                  fill={muted}
                >
                  {edge.label}
                </text>
              )}
            </g>
          );
        })}

        {/* Arrow marker */}
        <defs>
          <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
            <polygon
              points="0 0, 10 3.5, 0 7"
              fill={isDark ? "#4b5563" : "#d1d5db"}
            />
          </marker>
        </defs>

        {/* Nodes */}
        {positionedNodes.map((node) => {
          const color = NODE_COLORS[node.type] || "#6b7280";
          const icon = NODE_ICONS[node.type] || "●";

          return (
            <g
              key={node.id}
              onClick={() => onNodeClick?.(node.id)}
              style={{ cursor: onNodeClick ? "pointer" : "default" }}
            >
              <rect
                x={node.x}
                y={node.y}
                width={120}
                height={50}
                rx={8}
                fill={isDark ? "#16162a" : "#f9fafb"}
                stroke={color}
                strokeWidth={2}
              />
              <text
                x={(node.x ?? 0) + 16}
                y={(node.y ?? 0) + 22}
                fontSize="16"
              >
                {icon}
              </text>
              <text
                x={(node.x ?? 0) + 36}
                y={(node.y ?? 0) + 24}
                fontSize="12"
                fontWeight={600}
                fill={text}
              >
                {node.label.length > 12 ? node.label.slice(0, 12) + "…" : node.label}
              </text>
              <text
                x={(node.x ?? 0) + 36}
                y={(node.y ?? 0) + 40}
                fontSize="10"
                fill={muted}
              >
                {node.type}
              </text>
            </g>
          );
        })}
      </svg>
    </div>
  );
}
