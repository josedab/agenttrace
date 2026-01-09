"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Database, Trash2, Eye, ChevronRight, ChevronDown } from "lucide-react";

import { api } from "@/lib/api";
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";

interface DatasetItemTableProps {
  datasetId: string;
}

export function DatasetItemTable({ datasetId }: DatasetItemTableProps) {
  const queryClient = useQueryClient();
  const [selectedItem, setSelectedItem] = React.useState<any>(null);
  const [expandedRows, setExpandedRows] = React.useState<Set<string>>(new Set());

  const { data: items, isLoading } = useQuery({
    queryKey: ["dataset-items", datasetId],
    queryFn: () => api.datasets.getItems(datasetId),
  });

  const deleteMutation = useMutation({
    mutationFn: (itemId: string) => api.datasets.deleteItem(datasetId, itemId),
    onSuccess: () => {
      toast.success("Item deleted");
      queryClient.invalidateQueries({ queryKey: ["dataset-items", datasetId] });
      queryClient.invalidateQueries({ queryKey: ["dataset", datasetId] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to delete item");
    },
  });

  const toggleRow = (id: string) => {
    setExpandedRows((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  if (isLoading) {
    return <ItemTableSkeleton />;
  }

  if (!items || items.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 border rounded-lg bg-card">
        <Database className="h-12 w-12 text-muted-foreground mb-4" />
        <p className="text-lg font-medium">No items yet</p>
        <p className="text-sm text-muted-foreground mt-1">
          Add items to this dataset to start running experiments
        </p>
      </div>
    );
  }

  return (
    <>
      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[40px]"></TableHead>
              <TableHead>Input</TableHead>
              <TableHead>Expected Output</TableHead>
              <TableHead className="w-[100px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => {
              const isExpanded = expandedRows.has(item.id);
              return (
                <React.Fragment key={item.id}>
                  <TableRow>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 w-6 p-0"
                        onClick={() => toggleRow(item.id)}
                      >
                        {isExpanded ? (
                          <ChevronDown className="h-4 w-4" />
                        ) : (
                          <ChevronRight className="h-4 w-4" />
                        )}
                      </Button>
                    </TableCell>
                    <TableCell className="max-w-[300px]">
                      <p className="truncate text-sm">
                        {truncateJson(item.input, 100)}
                      </p>
                    </TableCell>
                    <TableCell className="max-w-[300px]">
                      <p className="truncate text-sm text-muted-foreground">
                        {item.expectedOutput
                          ? truncateJson(item.expectedOutput, 100)
                          : "-"}
                      </p>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setSelectedItem(item)}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => deleteMutation.mutate(item.id)}
                          disabled={deleteMutation.isPending}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                  {isExpanded && (
                    <TableRow>
                      <TableCell colSpan={4} className="bg-muted/50">
                        <div className="grid md:grid-cols-2 gap-4 p-4">
                          <div>
                            <p className="text-sm font-medium mb-2">Input</p>
                            <pre className="text-xs bg-background p-3 rounded-lg overflow-auto max-h-48">
                              {formatJson(item.input)}
                            </pre>
                          </div>
                          <div>
                            <p className="text-sm font-medium mb-2">Expected Output</p>
                            <pre className="text-xs bg-background p-3 rounded-lg overflow-auto max-h-48">
                              {item.expectedOutput
                                ? formatJson(item.expectedOutput)
                                : "Not specified"}
                            </pre>
                          </div>
                          {item.metadata && (
                            <div className="md:col-span-2">
                              <p className="text-sm font-medium mb-2">Metadata</p>
                              <pre className="text-xs bg-background p-3 rounded-lg overflow-auto max-h-32">
                                {formatJson(item.metadata)}
                              </pre>
                            </div>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  )}
                </React.Fragment>
              );
            })}
          </TableBody>
        </Table>
      </div>

      {/* Item detail dialog */}
      <Dialog open={!!selectedItem} onOpenChange={() => setSelectedItem(null)}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-auto">
          <DialogHeader>
            <DialogTitle>Dataset Item</DialogTitle>
            <DialogDescription>
              View full item details
            </DialogDescription>
          </DialogHeader>
          {selectedItem && (
            <div className="space-y-4">
              <div>
                <p className="text-sm font-medium mb-2">Input</p>
                <pre className="text-sm bg-muted p-4 rounded-lg overflow-auto">
                  {formatJson(selectedItem.input)}
                </pre>
              </div>
              <div>
                <p className="text-sm font-medium mb-2">Expected Output</p>
                <pre className="text-sm bg-muted p-4 rounded-lg overflow-auto">
                  {selectedItem.expectedOutput
                    ? formatJson(selectedItem.expectedOutput)
                    : "Not specified"}
                </pre>
              </div>
              {selectedItem.metadata && (
                <div>
                  <p className="text-sm font-medium mb-2">Metadata</p>
                  <pre className="text-sm bg-muted p-4 rounded-lg overflow-auto">
                    {formatJson(selectedItem.metadata)}
                  </pre>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}

function ItemTableSkeleton() {
  return (
    <div className="border rounded-lg">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[40px]"></TableHead>
            <TableHead>Input</TableHead>
            <TableHead>Expected Output</TableHead>
            <TableHead className="w-[100px]">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {[...Array(5)].map((_, i) => (
            <TableRow key={i}>
              <TableCell>
                <Skeleton className="h-4 w-4" />
              </TableCell>
              <TableCell>
                <Skeleton className="h-4 w-48" />
              </TableCell>
              <TableCell>
                <Skeleton className="h-4 w-48" />
              </TableCell>
              <TableCell>
                <Skeleton className="h-8 w-16" />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function truncateJson(value: any, length: number): string {
  const str = typeof value === "string" ? value : JSON.stringify(value);
  if (str.length <= length) return str;
  return str.slice(0, length) + "...";
}

function formatJson(value: any): string {
  if (typeof value === "string") {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  }
  return JSON.stringify(value, null, 2);
}
