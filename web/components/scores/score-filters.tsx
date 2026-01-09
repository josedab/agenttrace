"use client";

import * as React from "react";
import { Search, Filter, X } from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface ScoreFilters {
  scoreName: string;
  source: string;
  minScore?: number;
  maxScore?: number;
}

interface ScoreFiltersProps {
  filters: ScoreFilters;
  onFiltersChange: (filters: ScoreFilters) => void;
}

export function ScoreFilters({ filters, onFiltersChange }: ScoreFiltersProps) {
  const activeFilterCount = [
    filters.scoreName,
    filters.source,
    filters.minScore !== undefined,
    filters.maxScore !== undefined,
  ].filter(Boolean).length;

  const clearFilters = () => {
    onFiltersChange({
      scoreName: "",
      source: "",
      minScore: undefined,
      maxScore: undefined,
    });
  };

  return (
    <div className="flex items-center gap-4">
      {/* Search by score name */}
      <div className="relative flex-1 max-w-sm">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Search by score name..."
          value={filters.scoreName}
          onChange={(e) =>
            onFiltersChange({ ...filters, scoreName: e.target.value })
          }
          className="pl-9"
        />
      </div>

      {/* Source filter */}
      <Select
        value={filters.source || "all"}
        onValueChange={(value) =>
          onFiltersChange({ ...filters, source: value === "all" ? "" : value })
        }
      >
        <SelectTrigger className="w-[150px]">
          <SelectValue placeholder="Source" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All sources</SelectItem>
          <SelectItem value="API">API</SelectItem>
          <SelectItem value="EVAL">Evaluator</SelectItem>
          <SelectItem value="ANNOTATION">Human</SelectItem>
        </SelectContent>
      </Select>

      {/* Advanced filters */}
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="outline" className="gap-2">
            <Filter className="h-4 w-4" />
            Filters
            {activeFilterCount > 0 && (
              <Badge variant="secondary" className="ml-1">
                {activeFilterCount}
              </Badge>
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-80" align="end">
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Score Range</Label>
              <div className="flex items-center gap-2">
                <Input
                  type="number"
                  placeholder="Min"
                  min={0}
                  max={1}
                  step={0.1}
                  value={filters.minScore ?? ""}
                  onChange={(e) =>
                    onFiltersChange({
                      ...filters,
                      minScore: e.target.value
                        ? parseFloat(e.target.value)
                        : undefined,
                    })
                  }
                  className="w-20"
                />
                <span className="text-muted-foreground">to</span>
                <Input
                  type="number"
                  placeholder="Max"
                  min={0}
                  max={1}
                  step={0.1}
                  value={filters.maxScore ?? ""}
                  onChange={(e) =>
                    onFiltersChange({
                      ...filters,
                      maxScore: e.target.value
                        ? parseFloat(e.target.value)
                        : undefined,
                    })
                  }
                  className="w-20"
                />
              </div>
            </div>

            {activeFilterCount > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={clearFilters}
                className="w-full"
              >
                <X className="h-4 w-4 mr-2" />
                Clear all filters
              </Button>
            )}
          </div>
        </PopoverContent>
      </Popover>

      {/* Active filter badges */}
      {activeFilterCount > 0 && (
        <div className="flex items-center gap-2">
          {filters.scoreName && (
            <Badge variant="secondary" className="gap-1">
              Name: {filters.scoreName}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => onFiltersChange({ ...filters, scoreName: "" })}
              />
            </Badge>
          )}
          {filters.source && (
            <Badge variant="secondary" className="gap-1">
              Source: {filters.source}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() => onFiltersChange({ ...filters, source: "" })}
              />
            </Badge>
          )}
          {(filters.minScore !== undefined || filters.maxScore !== undefined) && (
            <Badge variant="secondary" className="gap-1">
              Score: {filters.minScore ?? 0} - {filters.maxScore ?? 1}
              <X
                className="h-3 w-3 cursor-pointer"
                onClick={() =>
                  onFiltersChange({
                    ...filters,
                    minScore: undefined,
                    maxScore: undefined,
                  })
                }
              />
            </Badge>
          )}
        </div>
      )}
    </div>
  );
}
