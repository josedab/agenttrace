"use client";

import * as React from "react";
import Link from "next/link";
import { formatDistanceToNow } from "date-fns";
import { MoreHorizontal, Settings, Trash2, Play, Bot, User, Code } from "lucide-react";

import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

interface Evaluator {
  id: string;
  name: string;
  description?: string;
  type: "LLM_AS_JUDGE" | "CODE" | "HUMAN";
  status: "ACTIVE" | "INACTIVE";
  scoreName: string;
  createdAt: string;
  lastRunAt?: string;
  totalRuns: number;
}

interface EvaluatorListProps {
  evaluators: Evaluator[];
}

const evaluatorTypeIcons = {
  LLM_AS_JUDGE: Bot,
  CODE: Code,
  HUMAN: User,
};

const evaluatorTypeLabels = {
  LLM_AS_JUDGE: "LLM-as-Judge",
  CODE: "Code",
  HUMAN: "Human",
};

export function EvaluatorList({ evaluators }: EvaluatorListProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>All Evaluators</CardTitle>
        <CardDescription>
          Manage your automated and manual evaluation configurations.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Score Name</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Total Runs</TableHead>
                <TableHead>Last Run</TableHead>
                <TableHead className="w-[50px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {evaluators.map((evaluator) => {
                const TypeIcon = evaluatorTypeIcons[evaluator.type];
                return (
                  <TableRow key={evaluator.id}>
                    <TableCell>
                      <Link
                        href={`/evals/${evaluator.id}`}
                        className="font-medium hover:underline"
                      >
                        {evaluator.name}
                      </Link>
                      {evaluator.description && (
                        <p className="text-xs text-muted-foreground mt-0.5 truncate max-w-[200px]">
                          {evaluator.description}
                        </p>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <TypeIcon className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm">
                          {evaluatorTypeLabels[evaluator.type]}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <code className="text-sm bg-muted px-1.5 py-0.5 rounded">
                        {evaluator.scoreName}
                      </code>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={evaluator.status === "ACTIVE" ? "default" : "secondary"}
                      >
                        {evaluator.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm">{evaluator.totalRuns}</span>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-muted-foreground">
                        {evaluator.lastRunAt
                          ? formatDistanceToNow(new Date(evaluator.lastRunAt), {
                              addSuffix: true,
                            })
                          : "Never"}
                      </span>
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon">
                            <MoreHorizontal className="h-4 w-4" />
                            <span className="sr-only">Open menu</span>
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem asChild>
                            <Link href={`/evals/${evaluator.id}`}>
                              <Settings className="h-4 w-4 mr-2" />
                              Configure
                            </Link>
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <Play className="h-4 w-4 mr-2" />
                            Run Now
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
                );
              })}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}
