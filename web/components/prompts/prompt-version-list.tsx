"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";
import { formatDistanceToNow, format } from "date-fns";
import { Clock, Tag, Check } from "lucide-react";

import { cn } from "@/lib/utils";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

interface PromptVersionListProps {
  promptName: string;
}

export function PromptVersionList({ promptName }: PromptVersionListProps) {
  const { data: prompt, isLoading } = useQuery({
    queryKey: ["prompt", promptName],
    queryFn: () => api.prompts.getByName(promptName),
  });

  if (isLoading) {
    return <VersionListSkeleton />;
  }

  const versions = prompt?.versions || [];

  if (versions.length === 0) {
    return (
      <Card>
        <CardContent className="pt-6">
          <p className="text-sm text-muted-foreground text-center">
            No versions yet
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Version History</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {versions.map((version) => {
            const isActive = prompt?.activeVersion?.id === version.id;
            const labels = getVersionLabels(version.version, prompt?.labels || {});

            return (
              <div
                key={version.id}
                className={cn(
                  "flex items-start justify-between p-4 rounded-lg border",
                  isActive && "border-primary bg-primary/5"
                )}
              >
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">v{version.version}</span>
                    {isActive && (
                      <Badge variant="default" className="text-xs">
                        <Check className="h-3 w-3 mr-1" />
                        Active
                      </Badge>
                    )}
                    {labels.map((label) => (
                      <Badge key={label} variant="secondary" className="text-xs">
                        <Tag className="h-3 w-3 mr-1" />
                        {label}
                      </Badge>
                    ))}
                  </div>
                  <div className="flex items-center gap-4 text-xs text-muted-foreground">
                    <span className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      {formatDistanceToNow(new Date(version.createdAt), {
                        addSuffix: true,
                      })}
                    </span>
                    {version.author && (
                      <span>by {version.author}</span>
                    )}
                  </div>
                  {version.commitMessage && (
                    <p className="text-sm text-muted-foreground">
                      {version.commitMessage}
                    </p>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  <Button variant="outline" size="sm">
                    View
                  </Button>
                  {!isActive && (
                    <Button variant="outline" size="sm">
                      Restore
                    </Button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}

function getVersionLabels(
  version: number,
  labels: Record<string, number>
): string[] {
  return Object.entries(labels)
    .filter(([_, v]) => v === version)
    .map(([label]) => label);
}

function VersionListSkeleton() {
  return (
    <Card>
      <CardHeader>
        <Skeleton className="h-6 w-32" />
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <div
              key={i}
              className="flex items-start justify-between p-4 rounded-lg border"
            >
              <div className="space-y-2">
                <Skeleton className="h-5 w-24" />
                <Skeleton className="h-3 w-32" />
              </div>
              <Skeleton className="h-8 w-16" />
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
