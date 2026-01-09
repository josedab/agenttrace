"use client";

import * as React from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { Database, MoreHorizontal, Play, Edit, Trash2, FileUp } from "lucide-react";

import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { DatasetListSkeleton } from "@/components/datasets/dataset-list-skeleton";

export function DatasetList() {
  const { data: datasets, isLoading, error } = useQuery({
    queryKey: ["datasets"],
    queryFn: () => api.datasets.list(),
  });

  if (isLoading) {
    return <DatasetListSkeleton />;
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-destructive">Failed to load datasets</p>
      </div>
    );
  }

  if (!datasets || datasets.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 border rounded-lg bg-card">
        <Database className="h-12 w-12 text-muted-foreground mb-4" />
        <p className="text-lg font-medium">No datasets yet</p>
        <p className="text-sm text-muted-foreground mt-1">
          Create your first dataset to start running experiments
        </p>
      </div>
    );
  }

  return (
    <div className="border rounded-lg">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[250px]">Name</TableHead>
            <TableHead>Description</TableHead>
            <TableHead className="w-[100px]">Items</TableHead>
            <TableHead className="w-[100px]">Runs</TableHead>
            <TableHead className="w-[150px]">Last Updated</TableHead>
            <TableHead className="w-[100px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {datasets.map((dataset) => (
            <TableRow key={dataset.id}>
              <TableCell>
                <Link
                  href={`/datasets/${dataset.id}`}
                  className="font-medium hover:underline"
                >
                  {dataset.name}
                </Link>
              </TableCell>
              <TableCell>
                <p className="text-sm text-muted-foreground truncate max-w-[300px]">
                  {dataset.description || "-"}
                </p>
              </TableCell>
              <TableCell>
                <Badge variant="secondary">{dataset.itemCount || 0}</Badge>
              </TableCell>
              <TableCell>
                <Badge variant="outline">{dataset.runCount || 0}</Badge>
              </TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {formatDistanceToNow(new Date(dataset.updatedAt), {
                  addSuffix: true,
                })}
              </TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="sm">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem asChild>
                      <Link href={`/datasets/${dataset.id}`}>
                        <Edit className="h-4 w-4 mr-2" />
                        View Items
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <FileUp className="h-4 w-4 mr-2" />
                      Import Items
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <Play className="h-4 w-4 mr-2" />
                      Run Experiment
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem className="text-destructive">
                      <Trash2 className="h-4 w-4 mr-2" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
