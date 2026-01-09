"use client";

import * as React from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";

import { PageHeader } from "@/components/layout/page-header";
import { CheckpointDetail } from "@/components/checkpoints/checkpoint-detail";
import { Button } from "@/components/ui/button";

export default function CheckpointDetailPage() {
  const params = useParams();
  const checkpointId = params.checkpointId as string;
  const projectId = "project-1"; // In a real app, from context

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/checkpoints">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <PageHeader
          title="Checkpoint Detail"
          description="View checkpoint files and restore agent state."
        />
      </div>

      <CheckpointDetail projectId={projectId} checkpointId={checkpointId} />
    </div>
  );
}
