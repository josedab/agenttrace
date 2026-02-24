"use client";

import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { DiffFinding } from "@/hooks/use-diff-analysis";
import { AlertTriangle, AlertCircle, Info, ShieldAlert } from "lucide-react";

interface FindingsListProps {
  findings: DiffFinding[];
}

const severityConfig: Record<
  string,
  { icon: React.ElementType; color: string; bg: string }
> = {
  critical: {
    icon: ShieldAlert,
    color: "text-red-600",
    bg: "bg-red-100 dark:bg-red-900/30",
  },
  error: {
    icon: AlertCircle,
    color: "text-orange-600",
    bg: "bg-orange-100 dark:bg-orange-900/30",
  },
  warning: {
    icon: AlertTriangle,
    color: "text-yellow-600",
    bg: "bg-yellow-100 dark:bg-yellow-900/30",
  },
  info: {
    icon: Info,
    color: "text-blue-600",
    bg: "bg-blue-100 dark:bg-blue-900/30",
  },
};

export function FindingsList({ findings }: FindingsListProps) {
  const grouped = React.useMemo(() => {
    const groups: Record<string, DiffFinding[]> = {};
    for (const f of findings) {
      (groups[f.severity] ||= []).push(f);
    }
    return groups;
  }, [findings]);

  const severityOrder = ["critical", "error", "warning", "info"];

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>Findings</span>
          <Badge variant="outline">{findings.length}</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {findings.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-4">
            No findings detected ✓
          </p>
        ) : (
          <div className="space-y-4">
            {severityOrder.map((severity) => {
              const items = grouped[severity];
              if (!items?.length) return null;
              const config = severityConfig[severity];
              const Icon = config.icon;
              return (
                <div key={severity}>
                  <div className="flex items-center gap-1 mb-2">
                    <Icon className={cn("h-4 w-4", config.color)} />
                    <span className="text-sm font-medium capitalize">
                      {severity}
                    </span>
                    <Badge variant="secondary" className="text-xs ml-1">
                      {items.length}
                    </Badge>
                  </div>
                  <div className="space-y-1">
                    {items.map((finding) => (
                      <div
                        key={finding.id}
                        className={cn("p-2 rounded-md text-sm", config.bg)}
                      >
                        <div className="font-medium">{finding.title}</div>
                        {finding.description && (
                          <p className="text-xs text-muted-foreground mt-0.5">
                            {finding.description}
                          </p>
                        )}
                        <div className="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
                          <span>{finding.filePath}</span>
                          {finding.startLine && (
                            <span>L{finding.startLine}</span>
                          )}
                          <span>•</span>
                          <span>{finding.category}</span>
                          <span>•</span>
                          <span>
                            {Math.round(finding.confidence * 100)}% confidence
                          </span>
                        </div>
                      </div>
                    ))}
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
