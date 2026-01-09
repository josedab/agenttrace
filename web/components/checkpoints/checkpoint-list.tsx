"use client";

import * as React from "react";
import Link from "next/link";
import { formatDistanceToNow, format } from "date-fns";
import { toast } from "sonner";
import {
  GitBranch,
  RotateCcw,
  MoreHorizontal,
  FileCode,
  ChevronRight,
  Loader2,
  Clock,
  Zap,
} from "lucide-react";

import { useCheckpoints, useRestoreCheckpoint, CheckpointFilters as CheckpointFiltersType } from "@/hooks/use-checkpoints";
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
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Skeleton } from "@/components/ui/skeleton";

interface CheckpointListProps {
  projectId: string;
  filters?: CheckpointFiltersType;
}

interface Checkpoint {
  id: string;
  traceId: string;
  traceName: string;
  type: "auto" | "manual";
  description: string | null;
  fileCount: number;
  totalSize: number;
  createdAt: string;
}

export function CheckpointList({ projectId, filters }: CheckpointListProps) {
  const {
    data,
    isLoading,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useCheckpoints(projectId, filters);
  const restoreMutation = useRestoreCheckpoint(projectId);

  const [checkpointToRestore, setCheckpointToRestore] = React.useState<Checkpoint | null>(null);

  const checkpoints = data?.pages.flatMap((page) => page.checkpoints) ?? [];

  const handleRestore = () => {
    if (checkpointToRestore) {
      restoreMutation.mutate(checkpointToRestore.id, {
        onSuccess: () => {
          toast.success("Checkpoint restored successfully");
          setCheckpointToRestore(null);
        },
        onError: (error: Error) => {
          toast.error(error.message || "Failed to restore checkpoint");
        },
      });
    }
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-32" />
          <Skeleton className="h-4 w-64" />
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {[...Array(5)].map((_, i) => (
              <Skeleton key={i} className="h-16 w-full" />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <GitBranch className="h-5 w-5" />
            Checkpoints
          </CardTitle>
          <CardDescription>
            Agent state snapshots for debugging and recovery.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {checkpoints.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No checkpoints found. Checkpoints are created automatically when agents make significant changes.
            </div>
          ) : (
            <>
              <div className="border rounded-lg">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Checkpoint</TableHead>
                      <TableHead>Trace</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Files</TableHead>
                      <TableHead>Size</TableHead>
                      <TableHead>Created</TableHead>
                      <TableHead className="w-[50px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {checkpoints.map((checkpoint) => (
                      <TableRow key={checkpoint.id}>
                        <TableCell>
                          <Link
                            href={`/checkpoints/${checkpoint.id}`}
                            className="flex items-center gap-2 hover:underline"
                          >
                            <GitBranch className="h-4 w-4 text-muted-foreground" />
                            <div>
                              <span className="font-mono text-sm">
                                {checkpoint.id.slice(0, 8)}
                              </span>
                              {checkpoint.description && (
                                <p className="text-xs text-muted-foreground truncate max-w-[200px]">
                                  {checkpoint.description}
                                </p>
                              )}
                            </div>
                          </Link>
                        </TableCell>
                        <TableCell>
                          <Link
                            href={`/traces/${checkpoint.traceId}`}
                            className="text-sm hover:underline"
                          >
                            {checkpoint.traceName}
                          </Link>
                        </TableCell>
                        <TableCell>
                          {checkpoint.type === "auto" ? (
                            <Badge variant="secondary" className="gap-1">
                              <Zap className="h-3 w-3" />
                              Auto
                            </Badge>
                          ) : (
                            <Badge variant="outline" className="gap-1">
                              <Clock className="h-3 w-3" />
                              Manual
                            </Badge>
                          )}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1 text-sm">
                            <FileCode className="h-4 w-4 text-muted-foreground" />
                            {checkpoint.fileCount}
                          </div>
                        </TableCell>
                        <TableCell>
                          <span className="text-sm text-muted-foreground">
                            {formatBytes(checkpoint.totalSize)}
                          </span>
                        </TableCell>
                        <TableCell>
                          <span className="text-sm text-muted-foreground">
                            {formatDistanceToNow(new Date(checkpoint.createdAt), {
                              addSuffix: true,
                            })}
                          </span>
                        </TableCell>
                        <TableCell>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon">
                                <MoreHorizontal className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem asChild>
                                <Link href={`/checkpoints/${checkpoint.id}`}>
                                  <ChevronRight className="h-4 w-4 mr-2" />
                                  View Details
                                </Link>
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => setCheckpointToRestore(checkpoint)}
                              >
                                <RotateCcw className="h-4 w-4 mr-2" />
                                Restore
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {hasNextPage && (
                <div className="flex justify-center mt-4">
                  <Button
                    variant="outline"
                    onClick={() => fetchNextPage()}
                    disabled={isFetchingNextPage}
                  >
                    {isFetchingNextPage ? (
                      <>
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        Loading...
                      </>
                    ) : (
                      "Load More"
                    )}
                  </Button>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Restore Confirmation */}
      <AlertDialog open={!!checkpointToRestore} onOpenChange={() => setCheckpointToRestore(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Restore Checkpoint?</AlertDialogTitle>
            <AlertDialogDescription>
              This will restore the agent state to the checkpoint from{" "}
              {checkpointToRestore &&
                format(new Date(checkpointToRestore.createdAt), "PPpp")}
              . This action will overwrite current files.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleRestore}
              disabled={restoreMutation.isPending}
            >
              {restoreMutation.isPending ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Restoring...
                </>
              ) : (
                <>
                  <RotateCcw className="h-4 w-4 mr-2" />
                  Restore
                </>
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
