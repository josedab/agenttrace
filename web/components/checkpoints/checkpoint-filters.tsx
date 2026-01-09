"use client";

import * as React from "react";
import { X, Filter } from "lucide-react";

import { CheckpointFilters as CheckpointFiltersType } from "@/hooks/use-checkpoints";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";

interface CheckpointFiltersProps {
  filters: CheckpointFiltersType;
  onFiltersChange: (filters: CheckpointFiltersType) => void;
}

export function CheckpointFilters({ filters, onFiltersChange }: CheckpointFiltersProps) {
  const activeFiltersCount = [filters.traceId, filters.type].filter(Boolean).length;

  const updateFilter = (key: keyof CheckpointFiltersType, value: string | undefined) => {
    onFiltersChange({
      ...filters,
      [key]: value,
    });
  };

  const clearFilters = () => {
    onFiltersChange({
      traceId: undefined,
      type: undefined,
    });
  };

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex flex-wrap items-end gap-4">
          <div className="flex items-center gap-2 text-muted-foreground">
            <Filter className="h-4 w-4" />
            <span className="text-sm font-medium">Filters</span>
          </div>

          <div className="space-y-2">
            <Label>Trace ID</Label>
            <Input
              placeholder="Filter by trace..."
              value={filters.traceId || ""}
              onChange={(e) => updateFilter("traceId", e.target.value || undefined)}
              className="w-[250px]"
            />
          </div>

          <div className="space-y-2">
            <Label>Type</Label>
            <Select
              value={filters.type || "all"}
              onValueChange={(value) =>
                updateFilter("type", value === "all" ? undefined : (value as "auto" | "manual"))
              }
            >
              <SelectTrigger className="w-[150px]">
                <SelectValue placeholder="All types" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All types</SelectItem>
                <SelectItem value="auto">Auto</SelectItem>
                <SelectItem value="manual">Manual</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {activeFiltersCount > 0 && (
            <Button
              variant="ghost"
              size="sm"
              onClick={clearFilters}
              className="gap-1"
            >
              Clear
              <Badge variant="secondary" className="ml-1 px-1.5">
                {activeFiltersCount}
              </Badge>
              <X className="h-3 w-3" />
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
