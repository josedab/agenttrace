"use client";

import * as React from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { MessageSquare, MoreHorizontal, Play, History, Edit, Trash2 } from "lucide-react";

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
import { PromptListSkeleton } from "@/components/prompts/prompt-list-skeleton";

export function PromptList() {
  const { data: prompts, isLoading, error } = useQuery({
    queryKey: ["prompts"],
    queryFn: () => api.prompts.list(),
  });

  if (isLoading) {
    return <PromptListSkeleton />;
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-destructive">Failed to load prompts</p>
      </div>
    );
  }

  if (!prompts || prompts.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 border rounded-lg bg-card">
        <MessageSquare className="h-12 w-12 text-muted-foreground mb-4" />
        <p className="text-lg font-medium">No prompts yet</p>
        <p className="text-sm text-muted-foreground mt-1">
          Create your first prompt to get started
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
            <TableHead>Labels</TableHead>
            <TableHead className="w-[100px]">Versions</TableHead>
            <TableHead className="w-[150px]">Last Updated</TableHead>
            <TableHead className="w-[100px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {prompts.map((prompt) => (
            <TableRow key={prompt.id}>
              <TableCell>
                <Link
                  href={`/prompts/${encodeURIComponent(prompt.name)}`}
                  className="font-medium hover:underline"
                >
                  {prompt.name}
                </Link>
                {prompt.description && (
                  <p className="text-xs text-muted-foreground mt-1 truncate max-w-[200px]">
                    {prompt.description}
                  </p>
                )}
              </TableCell>
              <TableCell>
                <div className="flex flex-wrap gap-1">
                  {prompt.labels?.production && (
                    <Badge variant="default">production</Badge>
                  )}
                  {prompt.labels?.staging && (
                    <Badge variant="secondary">staging</Badge>
                  )}
                  {prompt.labels?.latest && (
                    <Badge variant="outline">latest</Badge>
                  )}
                </div>
              </TableCell>
              <TableCell>
                <span className="text-muted-foreground">{prompt.versionCount || 1}</span>
              </TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {formatDistanceToNow(new Date(prompt.updatedAt), {
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
                      <Link href={`/prompts/${encodeURIComponent(prompt.name)}`}>
                        <Edit className="h-4 w-4 mr-2" />
                        Edit
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <Link href={`/prompts/${encodeURIComponent(prompt.name)}/playground`}>
                        <Play className="h-4 w-4 mr-2" />
                        Playground
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem asChild>
                      <Link href={`/prompts/${encodeURIComponent(prompt.name)}/versions`}>
                        <History className="h-4 w-4 mr-2" />
                        Version History
                      </Link>
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
