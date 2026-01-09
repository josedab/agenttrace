"use client";

import * as React from "react";
import Link from "next/link";
import { formatDistanceToNow } from "date-fns";
import { Bot, User, Code, ExternalLink } from "lucide-react";

import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

interface Score {
  id: string;
  traceId: string;
  observationId?: string;
  name: string;
  value: number | boolean | string;
  dataType: "NUMERIC" | "BOOLEAN" | "CATEGORICAL";
  source: "API" | "EVAL" | "ANNOTATION";
  comment?: string;
  createdAt: string;
  evaluatorId?: string;
  userId?: string;
}

interface ScoreListProps {
  scores: Score[];
}

const sourceIcons = {
  API: Code,
  EVAL: Bot,
  ANNOTATION: User,
};

const sourceLabels = {
  API: "API",
  EVAL: "Evaluator",
  ANNOTATION: "Human",
};

export function ScoreList({ scores }: ScoreListProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>All Scores</CardTitle>
        <CardDescription>
          Scores submitted via API, evaluators, or human annotation.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Value</TableHead>
                <TableHead>Source</TableHead>
                <TableHead>Trace</TableHead>
                <TableHead>Comment</TableHead>
                <TableHead>Created</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {scores.map((score) => {
                const SourceIcon = sourceIcons[score.source];
                return (
                  <TableRow key={score.id}>
                    <TableCell>
                      <code className="text-sm bg-muted px-1.5 py-0.5 rounded">
                        {score.name}
                      </code>
                    </TableCell>
                    <TableCell>
                      <ScoreValue score={score} />
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <SourceIcon className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm">
                          {sourceLabels[score.source]}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Link
                        href={`/traces/${score.traceId}`}
                        className="flex items-center gap-1 text-sm text-primary hover:underline"
                      >
                        {score.traceId.slice(0, 8)}...
                        <ExternalLink className="h-3 w-3" />
                      </Link>
                    </TableCell>
                    <TableCell>
                      {score.comment ? (
                        <p className="text-sm text-muted-foreground truncate max-w-[200px]">
                          {score.comment}
                        </p>
                      ) : (
                        <span className="text-sm text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-muted-foreground">
                        {formatDistanceToNow(new Date(score.createdAt), {
                          addSuffix: true,
                        })}
                      </span>
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

function ScoreValue({ score }: { score: Score }) {
  if (score.dataType === "BOOLEAN") {
    return (
      <Badge variant={score.value ? "default" : "destructive"}>
        {score.value ? "Pass" : "Fail"}
      </Badge>
    );
  }

  if (score.dataType === "NUMERIC") {
    const numValue = typeof score.value === "number" ? score.value : 0;
    return (
      <Badge
        variant={
          numValue >= 0.7
            ? "default"
            : numValue >= 0.4
            ? "secondary"
            : "destructive"
        }
      >
        {numValue.toFixed(2)}
      </Badge>
    );
  }

  // Categorical
  return <Badge variant="outline">{String(score.value)}</Badge>;
}
