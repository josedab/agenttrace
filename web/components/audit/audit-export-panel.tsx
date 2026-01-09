"use client";

import * as React from "react";
import { format, subDays } from "date-fns";
import { toast } from "sonner";
import {
  Download,
  FileJson,
  FileSpreadsheet,
  Loader2,
  CheckCircle,
  Clock,
  AlertCircle,
} from "lucide-react";

import {
  useAuditExportJobs,
  useCreateAuditExport,
  useDownloadAuditExport,
} from "@/hooks/use-audit-logs";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";

interface AuditExportPanelProps {
  organizationId: string;
}

export function AuditExportPanel({ organizationId }: AuditExportPanelProps) {
  const [dateRange, setDateRange] = React.useState("30");
  const [exportFormat, setExportFormat] = React.useState<"json" | "csv">("json");

  const { data: exportJobs, isLoading: jobsLoading } = useAuditExportJobs(organizationId);
  const createExportMutation = useCreateAuditExport(organizationId);
  const downloadMutation = useDownloadAuditExport(organizationId);

  const handleCreateExport = () => {
    const endDate = new Date();
    const startDate = subDays(endDate, parseInt(dateRange));

    createExportMutation.mutate(
      {
        startDate: startDate.toISOString(),
        endDate: endDate.toISOString(),
        format: exportFormat,
      },
      {
        onSuccess: () => {
          toast.success("Export job created. You can download it when ready.");
        },
        onError: (error: Error) => {
          toast.error(error.message || "Failed to create export job");
        },
      }
    );
  };

  const handleDownload = (jobId: string) => {
    downloadMutation.mutate(jobId, {
      onSuccess: (data) => {
        // Trigger download
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `audit-logs-${jobId}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      },
      onError: (error: Error) => {
        toast.error(error.message || "Failed to download export");
      },
    });
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "completed":
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case "pending":
      case "processing":
        return <Clock className="h-4 w-4 text-yellow-500" />;
      case "failed":
        return <AlertCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Clock className="h-4 w-4" />;
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Download className="h-4 w-4" />
          Export Logs
        </CardTitle>
        <CardDescription>
          Download audit logs for compliance and analysis.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-3">
          <div className="space-y-2">
            <Label>Date Range</Label>
            <Select value={dateRange} onValueChange={setDateRange}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="7">Last 7 days</SelectItem>
                <SelectItem value="30">Last 30 days</SelectItem>
                <SelectItem value="90">Last 90 days</SelectItem>
                <SelectItem value="365">Last year</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Format</Label>
            <Select value={exportFormat} onValueChange={(v) => setExportFormat(v as "json" | "csv")}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="json">
                  <div className="flex items-center gap-2">
                    <FileJson className="h-4 w-4" />
                    JSON
                  </div>
                </SelectItem>
                <SelectItem value="csv">
                  <div className="flex items-center gap-2">
                    <FileSpreadsheet className="h-4 w-4" />
                    CSV
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <Button
            onClick={handleCreateExport}
            disabled={createExportMutation.isPending}
            className="w-full"
          >
            {createExportMutation.isPending ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Creating...
              </>
            ) : (
              <>
                <Download className="h-4 w-4 mr-2" />
                Create Export
              </>
            )}
          </Button>
        </div>

        {/* Recent Exports */}
        {jobsLoading ? (
          <div className="space-y-2 pt-4 border-t">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        ) : exportJobs && exportJobs.length > 0 ? (
          <div className="space-y-2 pt-4 border-t">
            <Label>Recent Exports</Label>
            <div className="space-y-2">
              {exportJobs.slice(0, 3).map((job) => (
                <div
                  key={job.id}
                  className="flex items-center justify-between p-2 rounded-md bg-muted/50"
                >
                  <div className="flex items-center gap-2">
                    {getStatusIcon(job.status)}
                    <div className="text-sm">
                      <p className="font-medium">
                        {format(new Date(job.createdAt), "MMM d, yyyy")}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {job.format.toUpperCase()}
                      </p>
                    </div>
                  </div>
                  {job.status === "completed" && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDownload(job.id)}
                      disabled={downloadMutation.isPending}
                    >
                      <Download className="h-4 w-4" />
                    </Button>
                  )}
                  {job.status === "processing" && (
                    <Badge variant="secondary">Processing</Badge>
                  )}
                  {job.status === "failed" && (
                    <Badge variant="destructive">Failed</Badge>
                  )}
                </div>
              ))}
            </div>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
