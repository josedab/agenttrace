"use client";

import * as React from "react";
import { useRouter, useSearchParams, usePathname } from "next/navigation";
import { Search, Filter, X, Calendar } from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";

const LEVEL_OPTIONS = [
  { value: "all", label: "All Levels" },
  { value: "DEBUG", label: "Debug" },
  { value: "DEFAULT", label: "Default" },
  { value: "WARNING", label: "Warning" },
  { value: "ERROR", label: "Error" },
];

export function TraceFilters() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const [search, setSearch] = React.useState(searchParams.get("q") || "");
  const [level, setLevel] = React.useState(searchParams.get("level") || "all");
  const [isFilterOpen, setIsFilterOpen] = React.useState(false);

  // Advanced filters
  const [minLatency, setMinLatency] = React.useState(searchParams.get("minLatency") || "");
  const [maxLatency, setMaxLatency] = React.useState(searchParams.get("maxLatency") || "");
  const [minCost, setMinCost] = React.useState(searchParams.get("minCost") || "");
  const [maxCost, setMaxCost] = React.useState(searchParams.get("maxCost") || "");
  const [startDate, setStartDate] = React.useState(searchParams.get("startDate") || "");
  const [endDate, setEndDate] = React.useState(searchParams.get("endDate") || "");

  // Count active filters
  const activeFilters = [
    level !== "all" ? "level" : null,
    minLatency ? "minLatency" : null,
    maxLatency ? "maxLatency" : null,
    minCost ? "minCost" : null,
    maxCost ? "maxCost" : null,
    startDate ? "startDate" : null,
    endDate ? "endDate" : null,
  ].filter(Boolean).length;

  const updateSearchParams = React.useCallback(
    (updates: Record<string, string | null>) => {
      const params = new URLSearchParams(searchParams.toString());

      Object.entries(updates).forEach(([key, value]) => {
        if (value === null || value === "" || value === "all") {
          params.delete(key);
        } else {
          params.set(key, value);
        }
      });

      router.push(`${pathname}?${params.toString()}`);
    },
    [pathname, router, searchParams]
  );

  // Debounced search
  React.useEffect(() => {
    const timer = setTimeout(() => {
      updateSearchParams({ q: search || null });
    }, 300);

    return () => clearTimeout(timer);
  }, [search, updateSearchParams]);

  const handleLevelChange = (value: string) => {
    setLevel(value);
    updateSearchParams({ level: value === "all" ? null : value });
  };

  const applyFilters = () => {
    updateSearchParams({
      minLatency: minLatency || null,
      maxLatency: maxLatency || null,
      minCost: minCost || null,
      maxCost: maxCost || null,
      startDate: startDate || null,
      endDate: endDate || null,
    });
    setIsFilterOpen(false);
  };

  const clearFilters = () => {
    setLevel("all");
    setMinLatency("");
    setMaxLatency("");
    setMinCost("");
    setMaxCost("");
    setStartDate("");
    setEndDate("");
    updateSearchParams({
      level: null,
      minLatency: null,
      maxLatency: null,
      minCost: null,
      maxCost: null,
      startDate: null,
      endDate: null,
    });
  };

  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
      {/* Search */}
      <div className="relative flex-1">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search traces by name, input, or output..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Level filter */}
      <Select value={level} onValueChange={handleLevelChange}>
        <SelectTrigger className="w-[140px]">
          <SelectValue placeholder="Level" />
        </SelectTrigger>
        <SelectContent>
          {LEVEL_OPTIONS.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Advanced filters */}
      <Popover open={isFilterOpen} onOpenChange={setIsFilterOpen}>
        <PopoverTrigger asChild>
          <Button variant="outline" className="gap-2">
            <Filter className="h-4 w-4" />
            Filters
            {activeFilters > 0 && (
              <Badge variant="secondary" className="ml-1">
                {activeFilters}
              </Badge>
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-80" align="end">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h4 className="font-medium">Advanced Filters</h4>
              {activeFilters > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={clearFilters}
                  className="h-8 px-2"
                >
                  Clear all
                </Button>
              )}
            </div>

            {/* Latency range */}
            <div className="space-y-2">
              <Label>Latency (ms)</Label>
              <div className="flex items-center gap-2">
                <Input
                  type="number"
                  placeholder="Min"
                  value={minLatency}
                  onChange={(e) => setMinLatency(e.target.value)}
                  className="h-8"
                />
                <span className="text-muted-foreground">-</span>
                <Input
                  type="number"
                  placeholder="Max"
                  value={maxLatency}
                  onChange={(e) => setMaxLatency(e.target.value)}
                  className="h-8"
                />
              </div>
            </div>

            {/* Cost range */}
            <div className="space-y-2">
              <Label>Cost ($)</Label>
              <div className="flex items-center gap-2">
                <Input
                  type="number"
                  step="0.0001"
                  placeholder="Min"
                  value={minCost}
                  onChange={(e) => setMinCost(e.target.value)}
                  className="h-8"
                />
                <span className="text-muted-foreground">-</span>
                <Input
                  type="number"
                  step="0.0001"
                  placeholder="Max"
                  value={maxCost}
                  onChange={(e) => setMaxCost(e.target.value)}
                  className="h-8"
                />
              </div>
            </div>

            {/* Date range */}
            <div className="space-y-2">
              <Label>Date Range</Label>
              <div className="flex items-center gap-2">
                <Input
                  type="datetime-local"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="h-8"
                />
                <span className="text-muted-foreground">-</span>
                <Input
                  type="datetime-local"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  className="h-8"
                />
              </div>
            </div>

            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsFilterOpen(false)}
              >
                Cancel
              </Button>
              <Button size="sm" onClick={applyFilters}>
                Apply
              </Button>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}
