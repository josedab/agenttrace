"use client";

import * as React from "react";
import { ChevronRight, ChevronDown, Cpu, Zap, Wrench } from "lucide-react";

import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

interface Observation {
  id: string;
  type: "SPAN" | "GENERATION" | "EVENT";
  name: string | null;
  startTime: string;
  endTime: string | null;
  latency: number | null;
  level: string;
  parentObservationId: string | null;
  model: string | null;
  totalCost: number | null;
  usage: {
    promptTokens: number | null;
    completionTokens: number | null;
    totalTokens: number | null;
  } | null;
}

interface TraceTreeProps {
  observations: Observation[];
  selectedId: string | null;
  onSelect: (id: string | null) => void;
}

interface TreeNode {
  observation: Observation;
  children: TreeNode[];
}

export function TraceTree({ observations, selectedId, onSelect }: TraceTreeProps) {
  // Build tree structure
  const tree = React.useMemo(() => {
    const nodeMap = new Map<string, TreeNode>();
    const roots: TreeNode[] = [];

    // Create nodes
    observations.forEach((obs) => {
      nodeMap.set(obs.id, { observation: obs, children: [] });
    });

    // Build tree
    observations.forEach((obs) => {
      const node = nodeMap.get(obs.id)!;
      if (obs.parentObservationId && nodeMap.has(obs.parentObservationId)) {
        nodeMap.get(obs.parentObservationId)!.children.push(node);
      } else {
        roots.push(node);
      }
    });

    // Sort children by start time
    const sortNodes = (nodes: TreeNode[]) => {
      nodes.sort(
        (a, b) =>
          new Date(a.observation.startTime).getTime() -
          new Date(b.observation.startTime).getTime()
      );
      nodes.forEach((node) => sortNodes(node.children));
    };
    sortNodes(roots);

    return roots;
  }, [observations]);

  return (
    <Card>
      <CardContent className="pt-6">
        {tree.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-8">
            No observations found for this trace
          </p>
        ) : (
          <div className="space-y-1">
            {tree.map((node) => (
              <TreeNodeComponent
                key={node.observation.id}
                node={node}
                depth={0}
                selectedId={selectedId}
                onSelect={onSelect}
              />
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

interface TreeNodeComponentProps {
  node: TreeNode;
  depth: number;
  selectedId: string | null;
  onSelect: (id: string | null) => void;
}

function TreeNodeComponent({
  node,
  depth,
  selectedId,
  onSelect,
}: TreeNodeComponentProps) {
  const [isExpanded, setIsExpanded] = React.useState(true);
  const hasChildren = node.children.length > 0;
  const { observation } = node;

  return (
    <div>
      <div
        className={cn(
          "flex items-center gap-2 py-1.5 px-2 rounded cursor-pointer transition-colors",
          selectedId === observation.id
            ? "bg-primary/10 border border-primary"
            : "hover:bg-muted"
        )}
        style={{ paddingLeft: `${depth * 24 + 8}px` }}
        onClick={() => onSelect(observation.id)}
      >
        {/* Expand/collapse button */}
        {hasChildren ? (
          <button
            onClick={(e) => {
              e.stopPropagation();
              setIsExpanded(!isExpanded);
            }}
            className="p-0.5 hover:bg-muted rounded"
          >
            {isExpanded ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </button>
        ) : (
          <div className="w-5" />
        )}

        {/* Icon */}
        <ObservationIcon type={observation.type} />

        {/* Name */}
        <span className="font-medium truncate flex-1">
          {observation.name || observation.type}
        </span>

        {/* Badges */}
        <div className="flex items-center gap-2">
          <TypeBadge type={observation.type} />
          {observation.model && (
            <Badge variant="outline" className="text-xs">
              {observation.model}
            </Badge>
          )}
          {observation.latency && (
            <span className="text-xs text-muted-foreground">
              {formatLatency(observation.latency)}
            </span>
          )}
        </div>
      </div>

      {/* Children */}
      {hasChildren && isExpanded && (
        <div>
          {node.children.map((child) => (
            <TreeNodeComponent
              key={child.observation.id}
              node={child}
              depth={depth + 1}
              selectedId={selectedId}
              onSelect={onSelect}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function ObservationIcon({ type }: { type: Observation["type"] }) {
  switch (type) {
    case "GENERATION":
      return <Cpu className="h-4 w-4 text-blue-500 flex-shrink-0" />;
    case "EVENT":
      return <Zap className="h-4 w-4 text-yellow-500 flex-shrink-0" />;
    default:
      return <Wrench className="h-4 w-4 text-green-500 flex-shrink-0" />;
  }
}

function TypeBadge({ type }: { type: Observation["type"] }) {
  return (
    <Badge
      variant="outline"
      className={cn(
        "text-xs",
        type === "GENERATION" && "border-blue-500 text-blue-500",
        type === "EVENT" && "border-yellow-500 text-yellow-500",
        type === "SPAN" && "border-green-500 text-green-500"
      )}
    >
      {type}
    </Badge>
  );
}

function formatLatency(ms: number): string {
  if (ms >= 1000) {
    return (ms / 1000).toFixed(2) + "s";
  }
  return ms.toFixed(0) + "ms";
}
