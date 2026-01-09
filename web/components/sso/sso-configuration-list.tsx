"use client";

import * as React from "react";
import { formatDistanceToNow } from "date-fns";
import { toast } from "sonner";
import {
  Shield,
  Check,
  X,
  MoreHorizontal,
  Trash2,
  TestTube,
  ExternalLink,
  Power,
  PowerOff
} from "lucide-react";

import { useSSOConfigurations, useDeleteSSOConfiguration, useTestSSOConfiguration, useUpdateSSOConfiguration } from "@/hooks/use-sso";
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

interface SSOConfigurationListProps {
  organizationId: string;
}

export function SSOConfigurationList({ organizationId }: SSOConfigurationListProps) {
  const { data: configurations, isLoading, error } = useSSOConfigurations(organizationId);
  const deleteMutation = useDeleteSSOConfiguration(organizationId);
  const testMutation = useTestSSOConfiguration(organizationId);
  const updateMutation = useUpdateSSOConfiguration(organizationId);

  const [configToDelete, setConfigToDelete] = React.useState<string | null>(null);

  const handleDelete = () => {
    if (configToDelete) {
      deleteMutation.mutate(undefined, {
        onSuccess: () => {
          toast.success("SSO configuration deleted");
          setConfigToDelete(null);
        },
        onError: (error: Error) => {
          toast.error(error.message || "Failed to delete SSO configuration");
        },
      });
    }
  };

  const handleTest = () => {
    testMutation.mutate(undefined, {
      onSuccess: () => {
        toast.success("SSO configuration test successful");
      },
      onError: (error: Error) => {
        toast.error(error.message || "SSO configuration test failed");
      },
    });
  };

  const handleToggleEnabled = (enabled: boolean) => {
    updateMutation.mutate({ enabled }, {
      onSuccess: () => {
        toast.success(enabled ? "SSO enabled" : "SSO disabled");
      },
      onError: (error: Error) => {
        toast.error(error.message || "Failed to update SSO configuration");
      },
    });
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-96" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-32 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>SSO Configuration</CardTitle>
          <CardDescription>
            Failed to load SSO configurations. Please try again.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  if (!configurations || configurations.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            No SSO Configuration
          </CardTitle>
          <CardDescription>
            Set up Single Sign-On to allow team members to authenticate using your identity provider.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            SSO Configurations
          </CardTitle>
          <CardDescription>
            Manage your organization&apos;s Single Sign-On configurations.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Provider</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Domains</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="w-[50px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {configurations.map((config) => (
                  <TableRow key={config.id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Shield className="h-4 w-4 text-muted-foreground" />
                        <span className="font-medium uppercase">{config.provider}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      {config.enabled ? (
                        <Badge variant="default" className="bg-green-500">
                          <Check className="h-3 w-3 mr-1" />
                          Enabled
                        </Badge>
                      ) : (
                        <Badge variant="secondary">
                          <X className="h-3 w-3 mr-1" />
                          Disabled
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {config.allowedDomains?.slice(0, 3).map((domain) => (
                          <Badge key={domain} variant="outline" className="text-xs">
                            {domain}
                          </Badge>
                        ))}
                        {config.allowedDomains && config.allowedDomains.length > 3 && (
                          <Badge variant="outline" className="text-xs">
                            +{config.allowedDomains.length - 3} more
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-muted-foreground">
                        {formatDistanceToNow(new Date(config.createdAt), {
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
                          <DropdownMenuItem onClick={handleTest}>
                            <TestTube className="h-4 w-4 mr-2" />
                            Test Configuration
                          </DropdownMenuItem>
                          <DropdownMenuItem asChild>
                            <a href={`/api/auth/sso/${organizationId}`} target="_blank" rel="noopener">
                              <ExternalLink className="h-4 w-4 mr-2" />
                              SSO Login URL
                            </a>
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          {config.enabled ? (
                            <DropdownMenuItem onClick={() => handleToggleEnabled(false)}>
                              <PowerOff className="h-4 w-4 mr-2" />
                              Disable SSO
                            </DropdownMenuItem>
                          ) : (
                            <DropdownMenuItem onClick={() => handleToggleEnabled(true)}>
                              <Power className="h-4 w-4 mr-2" />
                              Enable SSO
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            className="text-destructive"
                            onClick={() => setConfigToDelete(config.id)}
                          >
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
        </CardContent>
      </Card>

      <AlertDialog open={!!configToDelete} onOpenChange={() => setConfigToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete SSO Configuration?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this SSO configuration? Team members using
              SSO will no longer be able to sign in until a new configuration is set up.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
