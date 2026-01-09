"use client";

import * as React from "react";
import { GitBranch } from "lucide-react";

import { PageHeader } from "@/components/layout/page-header";
import { CheckpointList } from "@/components/checkpoints/checkpoint-list";
import { CheckpointFilters } from "@/components/checkpoints/checkpoint-filters";

export default function CheckpointsPage() {
  // In a real app, this would come from context
  const projectId = "project-1";

  const [filters, setFilters] = React.useState({
    traceId: undefined as string | undefined,
    type: undefined as "auto" | "manual" | undefined,
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Checkpoints"
        description="Browse and restore agent state checkpoints for debugging and recovery."
        icon={GitBranch}
      />

      <CheckpointFilters filters={filters} onFiltersChange={setFilters} />
      <CheckpointList projectId={projectId} filters={filters} />
    </div>
  );
}
