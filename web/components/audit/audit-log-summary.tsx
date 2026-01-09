"use client";

import * as React from "react";
import {
  Activity,
  Users,
  Key,
  Shield,
  TrendingUp,
  TrendingDown,
} from "lucide-react";

import { useAuditSummary } from "@/hooks/use-audit-logs";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

interface AuditLogSummaryProps {
  organizationId: string;
}

export function AuditLogSummary({ organizationId }: AuditLogSummaryProps) {
  const { data: summary, isLoading } = useAuditSummary(organizationId);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-24" />
          <Skeleton className="h-4 w-48" />
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[...Array(4)].map((_, i) => (
              <Skeleton key={i} className="h-12 w-full" />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  const stats = [
    {
      label: "Total Events",
      value: summary?.totalEvents ?? 0,
      change: summary?.eventsTrend ?? 0,
      icon: Activity,
    },
    {
      label: "Active Users",
      value: summary?.activeUsers ?? 0,
      change: summary?.usersTrend ?? 0,
      icon: Users,
    },
    {
      label: "API Key Events",
      value: summary?.apiKeyEvents ?? 0,
      change: summary?.apiKeyTrend ?? 0,
      icon: Key,
    },
    {
      label: "Security Events",
      value: summary?.securityEvents ?? 0,
      change: summary?.securityTrend ?? 0,
      icon: Shield,
    },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Summary</CardTitle>
        <CardDescription>Last 30 days overview</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {stats.map((stat) => {
            const Icon = stat.icon;
            const isPositive = stat.change >= 0;

            return (
              <div
                key={stat.label}
                className="flex items-center justify-between p-3 rounded-lg bg-muted/50"
              >
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-background">
                    <Icon className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">{stat.label}</p>
                    <p className="text-lg font-semibold">{stat.value.toLocaleString()}</p>
                  </div>
                </div>
                <div
                  className={`flex items-center gap-1 text-xs ${
                    isPositive ? "text-green-500" : "text-red-500"
                  }`}
                >
                  {isPositive ? (
                    <TrendingUp className="h-3 w-3" />
                  ) : (
                    <TrendingDown className="h-3 w-3" />
                  )}
                  {Math.abs(stat.change)}%
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
