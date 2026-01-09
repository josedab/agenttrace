"use client";

import * as React from "react";
import Link from "next/link";
import { format } from "date-fns";
import { toast } from "sonner";
import {
  GitBranch,
  RotateCcw,
  FileCode,
  Clock,
  Zap,
  Loader2,
  Copy,
  Check,
  ExternalLink,
  FolderTree,
} from "lucide-react";

import { useCheckpoint, useRestoreCheckpoint } from "@/hooks/use-checkpoints";
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
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { ScrollArea } from "@/components/ui/scroll-area";

interface CheckpointDetailProps {
  projectId: string;
  checkpointId: string;
}

interface CheckpointFile {
  path: string;
  content: string;
  hash: string;
  size: number;
}

export function CheckpointDetail({ projectId, checkpointId }: CheckpointDetailProps) {
  const { data: checkpoint, isLoading } = useCheckpoint(projectId, checkpointId);
  const restoreMutation = useRestoreCheckpoint(projectId);

  const [showRestoreDialog, setShowRestoreDialog] = React.useState(false);
  const [selectedFile, setSelectedFile] = React.useState<CheckpointFile | null>(null);
  const [copiedHash, setCopiedHash] = React.useState<string | null>(null);

  const handleRestore = () => {
    restoreMutation.mutate(checkpointId, {
      onSuccess: () => {
        toast.success("Checkpoint restored successfully");
        setShowRestoreDialog(false);
      },
      onError: (error: Error) => {
        toast.error(error.message || "Failed to restore checkpoint");
      },
    });
  };

  const copyHash = (hash: string) => {
    navigator.clipboard.writeText(hash);
    setCopiedHash(hash);
    setTimeout(() => setCopiedHash(null), 2000);
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
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-96" />
          </CardHeader>
          <CardContent>
            <Skeleton className="h-48 w-full" />
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!checkpoint) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground">
          Checkpoint not found.
        </CardContent>
      </Card>
    );
  }

  // Select first file by default
  React.useEffect(() => {
    if (checkpoint?.files?.length > 0 && !selectedFile) {
      setSelectedFile(checkpoint.files[0]);
    }
  }, [checkpoint, selectedFile]);

  return (
    <>
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Checkpoint Info */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <GitBranch className="h-5 w-5" />
              Checkpoint Info
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-1">
              <label className="text-sm font-medium text-muted-foreground">ID</label>
              <p className="text-sm font-mono">{checkpoint.id}</p>
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium text-muted-foreground">Type</label>
              <div>
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
              </div>
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium text-muted-foreground">Created</label>
              <p className="text-sm">{format(new Date(checkpoint.createdAt), "PPpp")}</p>
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium text-muted-foreground">Trace</label>
              <Link
                href={`/traces/${checkpoint.traceId}`}
                className="text-sm text-primary hover:underline flex items-center gap-1"
              >
                {checkpoint.traceName}
                <ExternalLink className="h-3 w-3" />
              </Link>
            </div>

            {checkpoint.description && (
              <div className="space-y-1">
                <label className="text-sm font-medium text-muted-foreground">Description</label>
                <p className="text-sm">{checkpoint.description}</p>
              </div>
            )}

            <div className="space-y-1">
              <label className="text-sm font-medium text-muted-foreground">Files</label>
              <p className="text-sm">{checkpoint.files?.length ?? 0} files</p>
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium text-muted-foreground">Total Size</label>
              <p className="text-sm">
                {formatBytes(checkpoint.files?.reduce((acc, f) => acc + f.size, 0) ?? 0)}
              </p>
            </div>

            <Button
              onClick={() => setShowRestoreDialog(true)}
              className="w-full mt-4"
            >
              <RotateCcw className="h-4 w-4 mr-2" />
              Restore Checkpoint
            </Button>
          </CardContent>
        </Card>

        {/* File Browser */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FolderTree className="h-5 w-5" />
              Files
            </CardTitle>
            <CardDescription>
              Browse files included in this checkpoint.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="files" className="w-full">
              <TabsList>
                <TabsTrigger value="files">File List</TabsTrigger>
                <TabsTrigger value="content">Content</TabsTrigger>
              </TabsList>

              <TabsContent value="files" className="mt-4">
                <ScrollArea className="h-[400px] rounded-md border">
                  <div className="p-4 space-y-2">
                    {checkpoint.files?.map((file) => (
                      <div
                        key={file.path}
                        className={`flex items-center justify-between p-3 rounded-md cursor-pointer transition-colors ${
                          selectedFile?.path === file.path
                            ? "bg-primary/10 border border-primary/20"
                            : "hover:bg-muted"
                        }`}
                        onClick={() => setSelectedFile(file)}
                      >
                        <div className="flex items-center gap-3">
                          <FileCode className="h-4 w-4 text-muted-foreground" />
                          <div>
                            <p className="text-sm font-medium font-mono">{file.path}</p>
                            <p className="text-xs text-muted-foreground">
                              {formatBytes(file.size)}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          <code className="text-xs bg-muted px-1.5 py-0.5 rounded">
                            {file.hash.slice(0, 8)}
                          </code>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6"
                            onClick={(e) => {
                              e.stopPropagation();
                              copyHash(file.hash);
                            }}
                          >
                            {copiedHash === file.hash ? (
                              <Check className="h-3 w-3 text-green-500" />
                            ) : (
                              <Copy className="h-3 w-3" />
                            )}
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                </ScrollArea>
              </TabsContent>

              <TabsContent value="content" className="mt-4">
                {selectedFile ? (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <FileCode className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm font-mono">{selectedFile.path}</span>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          navigator.clipboard.writeText(selectedFile.content);
                          toast.success("Content copied to clipboard");
                        }}
                      >
                        <Copy className="h-4 w-4 mr-2" />
                        Copy
                      </Button>
                    </div>
                    <ScrollArea className="h-[350px] rounded-md border bg-muted/30">
                      <pre className="p-4 text-sm font-mono whitespace-pre-wrap">
                        {selectedFile.content}
                      </pre>
                    </ScrollArea>
                  </div>
                ) : (
                  <div className="h-[350px] flex items-center justify-center text-muted-foreground">
                    Select a file to view its content
                  </div>
                )}
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>

      {/* Restore Dialog */}
      <AlertDialog open={showRestoreDialog} onOpenChange={setShowRestoreDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Restore Checkpoint?</AlertDialogTitle>
            <AlertDialogDescription>
              This will restore the agent state to this checkpoint from{" "}
              {format(new Date(checkpoint.createdAt), "PPpp")}.
              Current files will be overwritten with the checkpoint contents.
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
