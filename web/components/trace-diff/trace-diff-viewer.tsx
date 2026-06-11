"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { useTraceDiff, type DiffNode } from "@/hooks/use-trace-diff";
import { ArrowLeftRight, Plus, Minus, Pencil, Equal } from "lucide-react";

interface TraceDiffViewerProps {
  leftTraceId?: string;
  rightTraceId?: string;
}

const diffColors: Record<string, string> = {
  added: "bg-green-50 border-green-200 dark:bg-green-950 dark:border-green-800",
  removed: "bg-red-50 border-red-200 dark:bg-red-950 dark:border-red-800",
  modified: "bg-yellow-50 border-yellow-200 dark:bg-yellow-950 dark:border-yellow-800",
  unchanged: "bg-gray-50 border-gray-200 dark:bg-gray-900 dark:border-gray-700",
  reordered: "bg-blue-50 border-blue-200 dark:bg-blue-950 dark:border-blue-800",
};

const diffIcons: Record<string, React.ReactNode> = {
  added: <Plus className="h-3 w-3 text-green-600" />,
  removed: <Minus className="h-3 w-3 text-red-600" />,
  modified: <Pencil className="h-3 w-3 text-yellow-600" />,
  unchanged: <Equal className="h-3 w-3 text-gray-400" />,
};

const diffBadgeVariants: Record<string, string> = {
  added: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
  removed: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
  modified: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
  unchanged: "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400",
};

function DiffNodeRow({ node, depth = 0 }: { node: DiffNode; depth?: number }) {
  const [expanded, setExpanded] = useState(node.diffType !== "unchanged");
  const hasChildren = node.children && node.children.length > 0;

  return (
    <div className="w-full">
      <div
        className={`flex items-center gap-2 px-3 py-2 border-l-2 cursor-pointer hover:opacity-80 ${diffColors[node.diffType]}`}
        style={{ paddingLeft: `${depth * 20 + 12}px` }}
        onClick={() => setExpanded(!expanded)}
      >
        <span className="flex-shrink-0">{diffIcons[node.diffType]}</span>
        {hasChildren && (
          <span className="text-xs text-muted-foreground">
            {expanded ? "▼" : "▶"}
          </span>
        )}
        <span className="font-mono text-sm font-medium">{node.spanName}</span>
        <Badge className={`text-xs ${diffBadgeVariants[node.diffType]}`}>
          {node.diffType}
        </Badge>
        {node.leftValue?.model && (
          <span className="text-xs text-muted-foreground ml-auto">
            {node.leftValue.model}
            {node.rightValue?.model && node.rightValue.model !== node.leftValue.model && (
              <> → {node.rightValue.model}</>
            )}
          </span>
        )}
      </div>

      {expanded && node.propertyDiffs && node.propertyDiffs.length > 0 && (
        <div
          className="bg-muted/50 px-3 py-1 text-xs border-l-2 border-dashed"
          style={{ paddingLeft: `${depth * 20 + 32}px` }}
        >
          {node.propertyDiffs.map((pd, i) => (
            <div key={i} className="flex gap-2 py-0.5">
              <span className="font-medium text-muted-foreground">{pd.property}:</span>
              <span className="text-red-600 line-through">{String(pd.leftValue)}</span>
              <span>→</span>
              <span className="text-green-600">{String(pd.rightValue)}</span>
              <Badge variant="outline" className="text-[10px] h-4">
                {pd.changeType}
              </Badge>
            </div>
          ))}
        </div>
      )}

      {expanded && hasChildren &&
        node.children!.map((child, i) => (
          <DiffNodeRow key={i} node={child} depth={depth + 1} />
        ))}
    </div>
  );
}

export function TraceDiffViewer({ leftTraceId: initialLeft, rightTraceId: initialRight }: TraceDiffViewerProps) {
  const [leftId, setLeftId] = useState(initialLeft || "");
  const [rightId, setRightId] = useState(initialRight || "");
  const [diffInput, setDiffInput] = useState<{ leftTraceId: string; rightTraceId: string } | null>(
    initialLeft && initialRight ? { leftTraceId: initialLeft, rightTraceId: initialRight } : null
  );

  const { data: result, isLoading, error } = useTraceDiff(diffInput);

  const handleCompare = () => {
    if (leftId && rightId) {
      setDiffInput({ leftTraceId: leftId, rightTraceId: rightId });
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ArrowLeftRight className="h-5 w-5" />
            Trace Diff
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2 items-end">
            <div className="flex-1">
              <label className="text-sm font-medium mb-1 block">Left Trace (baseline)</label>
              <Input
                placeholder="Enter trace ID..."
                value={leftId}
                onChange={(e) => setLeftId(e.target.value)}
              />
            </div>
            <div className="flex-1">
              <label className="text-sm font-medium mb-1 block">Right Trace (comparison)</label>
              <Input
                placeholder="Enter trace ID..."
                value={rightId}
                onChange={(e) => setRightId(e.target.value)}
              />
            </div>
            <Button onClick={handleCompare} disabled={!leftId || !rightId || isLoading}>
              {isLoading ? "Computing..." : "Compare"}
            </Button>
          </div>
        </CardContent>
      </Card>

      {error && (
        <Card className="border-red-200">
          <CardContent className="pt-4">
            <p className="text-sm text-red-600">Error: {(error as Error).message}</p>
          </CardContent>
        </Card>
      )}

      {result && (
        <>
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Diff Summary</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">+{result.summary?.addedCount || 0}</div>
                  <div className="text-xs text-muted-foreground">Added</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-red-600">-{result.summary?.removedCount || 0}</div>
                  <div className="text-xs text-muted-foreground">Removed</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-yellow-600">~{result.summary?.modifiedCount || 0}</div>
                  <div className="text-xs text-muted-foreground">Modified</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-gray-500">={result.summary?.unchangedCount || 0}</div>
                  <div className="text-xs text-muted-foreground">Unchanged</div>
                </div>
              </div>
              <div className="grid grid-cols-3 gap-4 mt-4 pt-4 border-t">
                <div className="text-center">
                  <div className="text-sm font-medium">
                    {result.summary?.costDelta > 0 ? "+" : ""}
                    ${result.summary?.costDelta?.toFixed(4) || "0.00"}
                  </div>
                  <div className="text-xs text-muted-foreground">Cost Delta</div>
                </div>
                <div className="text-center">
                  <div className="text-sm font-medium">
                    {result.summary?.latencyDeltaMs > 0 ? "+" : ""}
                    {result.summary?.latencyDeltaMs?.toFixed(0) || "0"}ms
                  </div>
                  <div className="text-xs text-muted-foreground">Latency Delta</div>
                </div>
                <div className="text-center">
                  <div className="text-sm font-medium">
                    {result.summary?.tokenDelta > 0 ? "+" : ""}
                    {result.summary?.tokenDelta || 0}
                  </div>
                  <div className="text-xs text-muted-foreground">Token Delta</div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Structural Diff Tree</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <div className="divide-y">
                {result.rootDiffs?.map((node: DiffNode, i: number) => (
                  <DiffNodeRow key={i} node={node} />
                ))}
              </div>
              {(!result.rootDiffs || result.rootDiffs.length === 0) && (
                <p className="p-4 text-sm text-muted-foreground text-center">No differences found</p>
              )}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
