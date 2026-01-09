"use client";

import * as React from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { ClipboardList, Users } from "lucide-react";

import { api } from "@/lib/api";
import { PageHeader } from "@/components/layout/page-header";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Progress } from "@/components/ui/progress";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface AnnotationQueue {
  id: string;
  name: string;
  scoreName: string;
  pendingCount: number;
  completedCount: number;
  totalCount: number;
  createdAt: string;
}

export default function AnnotationQueuesPage() {
  const { data: queues, isLoading, error } = useQuery({
    queryKey: ["annotation-queues"],
    queryFn: () => api.annotationQueues.list(),
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Annotation Queues"
        description="Review and annotate traces manually."
      />

      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[...Array(6)].map((_, i) => (
            <Card key={i}>
              <CardHeader>
                <Skeleton className="h-5 w-32" />
                <Skeleton className="h-4 w-24" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-2 w-full mb-2" />
                <Skeleton className="h-4 w-20" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : error ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="text-destructive">Failed to load annotation queues</p>
        </div>
      ) : queues && queues.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {queues.map((queue: AnnotationQueue) => {
            const progress =
              queue.totalCount > 0
                ? (queue.completedCount / queue.totalCount) * 100
                : 0;
            return (
              <Link key={queue.id} href={`/evals/queues/${queue.id}`}>
                <Card className="hover:border-primary/50 transition-colors cursor-pointer">
                  <CardHeader className="pb-2">
                    <div className="flex items-center justify-between">
                      <CardTitle className="text-base">{queue.name}</CardTitle>
                      {queue.pendingCount > 0 && (
                        <Badge variant="secondary">
                          {queue.pendingCount} pending
                        </Badge>
                      )}
                    </div>
                    <CardDescription>
                      Score: {queue.scoreName}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      <Progress value={progress} className="h-2" />
                      <div className="flex items-center justify-between text-xs text-muted-foreground">
                        <span>
                          {queue.completedCount} / {queue.totalCount} reviewed
                        </span>
                        <span>
                          {formatDistanceToNow(new Date(queue.createdAt), {
                            addSuffix: true,
                          })}
                        </span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </Link>
            );
          })}
        </div>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <ClipboardList className="h-12 w-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-semibold">No annotation queues</h3>
            <p className="text-sm text-muted-foreground mt-1">
              Create a human evaluator to start an annotation queue.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
