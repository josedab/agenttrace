"use client";

import * as React from "react";
import { format } from "date-fns";
import { CalendarIcon, X } from "lucide-react";

import { AUDIT_ACTIONS, RESOURCE_TYPES, AuditLogFilters as AuditLogFiltersType } from "@/hooks/use-audit-logs";
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
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface AuditLogFiltersProps {
  filters: AuditLogFiltersType;
  onFiltersChange: (filters: AuditLogFiltersType) => void;
}

export function AuditLogFilters({ filters, onFiltersChange }: AuditLogFiltersProps) {
  const [startDate, setStartDate] = React.useState<Date | undefined>(
    filters.startDate ? new Date(filters.startDate) : undefined
  );
  const [endDate, setEndDate] = React.useState<Date | undefined>(
    filters.endDate ? new Date(filters.endDate) : undefined
  );

  const activeFiltersCount = [
    filters.userId,
    filters.action,
    filters.resourceType,
    filters.startDate,
    filters.endDate,
  ].filter(Boolean).length;

  const updateFilter = (key: keyof AuditLogFiltersType, value: string | undefined) => {
    onFiltersChange({
      ...filters,
      [key]: value,
    });
  };

  const clearFilters = () => {
    setStartDate(undefined);
    setEndDate(undefined);
    onFiltersChange({
      userId: undefined,
      action: undefined,
      resourceType: undefined,
      startDate: undefined,
      endDate: undefined,
    });
  };

  const handleStartDateChange = (date: Date | undefined) => {
    setStartDate(date);
    updateFilter("startDate", date?.toISOString());
  };

  const handleEndDateChange = (date: Date | undefined) => {
    setEndDate(date);
    updateFilter("endDate", date?.toISOString());
  };

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex flex-wrap items-end gap-4">
          <div className="space-y-2">
            <Label>User ID</Label>
            <Input
              placeholder="Filter by user..."
              value={filters.userId || ""}
              onChange={(e) => updateFilter("userId", e.target.value || undefined)}
              className="w-[200px]"
            />
          </div>

          <div className="space-y-2">
            <Label>Action</Label>
            <Select
              value={filters.action || "all"}
              onValueChange={(value) => updateFilter("action", value === "all" ? undefined : value)}
            >
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="All actions" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All actions</SelectItem>
                {AUDIT_ACTIONS.map((action) => (
                  <SelectItem key={action.value} value={action.value}>
                    {action.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Resource Type</Label>
            <Select
              value={filters.resourceType || "all"}
              onValueChange={(value) => updateFilter("resourceType", value === "all" ? undefined : value)}
            >
              <SelectTrigger className="w-[150px]">
                <SelectValue placeholder="All types" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All types</SelectItem>
                {RESOURCE_TYPES.map((type) => (
                  <SelectItem key={type.value} value={type.value}>
                    {type.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Start Date</Label>
            <Popover>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  className={cn(
                    "w-[150px] justify-start text-left font-normal",
                    !startDate && "text-muted-foreground"
                  )}
                >
                  <CalendarIcon className="mr-2 h-4 w-4" />
                  {startDate ? format(startDate, "PPP") : "Pick date"}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0" align="start">
                <Calendar
                  mode="single"
                  selected={startDate}
                  onSelect={handleStartDateChange}
                  initialFocus
                />
              </PopoverContent>
            </Popover>
          </div>

          <div className="space-y-2">
            <Label>End Date</Label>
            <Popover>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  className={cn(
                    "w-[150px] justify-start text-left font-normal",
                    !endDate && "text-muted-foreground"
                  )}
                >
                  <CalendarIcon className="mr-2 h-4 w-4" />
                  {endDate ? format(endDate, "PPP") : "Pick date"}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0" align="start">
                <Calendar
                  mode="single"
                  selected={endDate}
                  onSelect={handleEndDateChange}
                  initialFocus
                />
              </PopoverContent>
            </Popover>
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
