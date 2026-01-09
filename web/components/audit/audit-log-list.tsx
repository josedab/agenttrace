"use client";

import * as React from "react";
import { formatDistanceToNow, format } from "date-fns";
import {
  User,
  Key,
  Building2,
  Shield,
  Settings,
  Users,
  LogIn,
  LogOut,
  Plus,
  Trash2,
  Edit,
  ChevronRight,
  Loader2,
} from "lucide-react";

import { useAuditLogs, AuditLogFilters } from "@/hooks/use-audit-logs";
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
import { ScrollArea } from "@/components/ui/scroll-area";

interface AuditLogListProps {
  organizationId: string;
  filters: AuditLogFilters;
}

const actionIcons: Record<string, React.ReactNode> = {
  login: <LogIn className="h-4 w-4" />,
  logout: <LogOut className="h-4 w-4" />,
  "api_key.create": <Plus className="h-4 w-4" />,
  "api_key.delete": <Trash2 className="h-4 w-4" />,
  "project.create": <Plus className="h-4 w-4" />,
  "project.update": <Edit className="h-4 w-4" />,
  "project.delete": <Trash2 className="h-4 w-4" />,
  "member.invite": <Plus className="h-4 w-4" />,
  "member.remove": <Trash2 className="h-4 w-4" />,
  "member.role_change": <Edit className="h-4 w-4" />,
  "sso.configure": <Shield className="h-4 w-4" />,
  "sso.enable": <Shield className="h-4 w-4" />,
  "sso.disable": <Shield className="h-4 w-4" />,
  "settings.update": <Settings className="h-4 w-4" />,
};

const resourceIcons: Record<string, React.ReactNode> = {
  user: <User className="h-4 w-4" />,
  project: <Building2 className="h-4 w-4" />,
  api_key: <Key className="h-4 w-4" />,
  organization: <Building2 className="h-4 w-4" />,
  sso_config: <Shield className="h-4 w-4" />,
  member: <Users className="h-4 w-4" />,
};

interface AuditLog {
  id: string;
  userId: string;
  userName: string;
  userEmail: string;
  action: string;
  resourceType: string;
  resourceId: string;
  metadata: Record<string, unknown>;
  ipAddress: string;
  userAgent: string;
  timestamp: string;
}

export function AuditLogList({ organizationId, filters }: AuditLogListProps) {
  const {
    data,
    isLoading,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useAuditLogs(organizationId, filters);

  const [selectedLog, setSelectedLog] = React.useState<AuditLog | null>(null);

  const logs = data?.pages.flatMap((page) => page.logs) ?? [];

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
              <Skeleton key={i} className="h-12 w-full" />
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
          <CardTitle>Activity Log</CardTitle>
          <CardDescription>
            Recent security and administrative events in your organization.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {logs.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No audit logs found for the selected filters.
            </div>
          ) : (
            <>
              <div className="border rounded-lg">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Event</TableHead>
                      <TableHead>User</TableHead>
                      <TableHead>Resource</TableHead>
                      <TableHead>IP Address</TableHead>
                      <TableHead>Time</TableHead>
                      <TableHead className="w-[40px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {logs.map((log) => (
                      <TableRow
                        key={log.id}
                        className="cursor-pointer hover:bg-muted/50"
                        onClick={() => setSelectedLog(log)}
                      >
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <div className="p-1.5 bg-muted rounded">
                              {actionIcons[log.action] || <Settings className="h-4 w-4" />}
                            </div>
                            <span className="font-medium">
                              {formatAction(log.action)}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col">
                            <span className="text-sm font-medium">{log.userName}</span>
                            <span className="text-xs text-muted-foreground">
                              {log.userEmail}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            {resourceIcons[log.resourceType] || <Building2 className="h-4 w-4" />}
                            <Badge variant="outline" className="text-xs">
                              {log.resourceType}
                            </Badge>
                          </div>
                        </TableCell>
                        <TableCell>
                          <code className="text-xs bg-muted px-1.5 py-0.5 rounded">
                            {log.ipAddress}
                          </code>
                        </TableCell>
                        <TableCell>
                          <span className="text-sm text-muted-foreground">
                            {formatDistanceToNow(new Date(log.timestamp), {
                              addSuffix: true,
                            })}
                          </span>
                        </TableCell>
                        <TableCell>
                          <ChevronRight className="h-4 w-4 text-muted-foreground" />
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

      {/* Detail Dialog */}
      <Dialog open={!!selectedLog} onOpenChange={() => setSelectedLog(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              {selectedLog && actionIcons[selectedLog.action]}
              {selectedLog && formatAction(selectedLog.action)}
            </DialogTitle>
            <DialogDescription>
              {selectedLog && format(new Date(selectedLog.timestamp), "PPpp")}
            </DialogDescription>
          </DialogHeader>

          {selectedLog && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="text-sm font-medium text-muted-foreground">User</label>
                  <p className="text-sm">{selectedLog.userName}</p>
                  <p className="text-xs text-muted-foreground">{selectedLog.userEmail}</p>
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium text-muted-foreground">IP Address</label>
                  <p className="text-sm font-mono">{selectedLog.ipAddress}</p>
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium text-muted-foreground">Resource Type</label>
                  <Badge variant="outline">{selectedLog.resourceType}</Badge>
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium text-muted-foreground">Resource ID</label>
                  <p className="text-sm font-mono">{selectedLog.resourceId}</p>
                </div>
              </div>

              <div className="space-y-1">
                <label className="text-sm font-medium text-muted-foreground">User Agent</label>
                <p className="text-xs text-muted-foreground break-all">{selectedLog.userAgent}</p>
              </div>

              {Object.keys(selectedLog.metadata).length > 0 && (
                <div className="space-y-1">
                  <label className="text-sm font-medium text-muted-foreground">Metadata</label>
                  <ScrollArea className="h-[200px] rounded-md border p-4">
                    <pre className="text-xs">
                      {JSON.stringify(selectedLog.metadata, null, 2)}
                    </pre>
                  </ScrollArea>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}

function formatAction(action: string): string {
  const actionLabels: Record<string, string> = {
    login: "User Login",
    logout: "User Logout",
    "api_key.create": "API Key Created",
    "api_key.delete": "API Key Deleted",
    "project.create": "Project Created",
    "project.update": "Project Updated",
    "project.delete": "Project Deleted",
    "member.invite": "Member Invited",
    "member.remove": "Member Removed",
    "member.role_change": "Role Changed",
    "sso.configure": "SSO Configured",
    "sso.enable": "SSO Enabled",
    "sso.disable": "SSO Disabled",
    "settings.update": "Settings Updated",
  };

  return actionLabels[action] || action;
}
