"use client";

import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { Shield, AlertTriangle, FileText, Terminal } from "lucide-react";

interface ReviewCardProps {
  review: {
    id: string;
    status: string;
    riskLevel: string;
    riskScore: number;
    proposedActions: {
      id: string;
      type: string;
      target: string;
      description: string;
      riskLevel: string;
    }[];
    createdAt: string;
  };
}

const riskColors: Record<string, string> = {
  low: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
  medium:
    "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400",
  high: "bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400",
  critical:
    "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400",
};

const actionIcons: Record<string, React.ElementType> = {
  file_write: FileText,
  file_delete: FileText,
  command_exec: Terminal,
  network_request: Shield,
  env_access: AlertTriangle,
};

export function ReviewCard({ review }: ReviewCardProps) {
  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center justify-between text-sm">
          <span className="flex items-center gap-2">
            <Shield className="h-4 w-4" />
            Review #{review.id.slice(0, 8)}
          </span>
          <div className="flex items-center gap-2">
            <Badge
              className={cn("text-xs", riskColors[review.riskLevel])}
            >
              {review.riskLevel} risk ({Math.round(review.riskScore)})
            </Badge>
            <Badge variant="outline" className="text-xs capitalize">
              {review.status}
            </Badge>
          </div>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {review.proposedActions.map((action) => {
            const Icon = actionIcons[action.type] || Shield;
            return (
              <div
                key={action.id}
                className={cn(
                  "flex items-start gap-2 p-2 rounded text-sm",
                  riskColors[action.riskLevel]
                )}
              >
                <Icon className="h-4 w-4 shrink-0 mt-0.5" />
                <div>
                  <div className="font-medium capitalize">
                    {action.type.replace(/_/g, " ")}: {action.target}
                  </div>
                  <div className="text-xs opacity-80">
                    {action.description}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
