"use client";

import * as React from "react";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api";
import { PageHeader } from "@/components/layout/page-header";
import { AnnotationInterface } from "@/components/evals/annotation-interface";
import { AnnotationInterfaceSkeleton } from "@/components/evals/annotation-interface-skeleton";

export default function AnnotationQueuePage() {
  const params = useParams();
  const queueId = params.queueId as string;

  const { data: queue, isLoading, error } = useQuery({
    queryKey: ["annotation-queue", queueId],
    queryFn: () => api.annotationQueues.get(queueId),
  });

  if (isLoading) {
    return <AnnotationInterfaceSkeleton />;
  }

  if (error || !queue) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-destructive">Failed to load annotation queue</p>
        <p className="text-sm text-muted-foreground mt-1">
          The queue may have been deleted or you don't have access.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={queue.name}
        description={`${queue.pendingCount} items pending review`}
      />
      <AnnotationInterface queue={queue} />
    </div>
  );
}
