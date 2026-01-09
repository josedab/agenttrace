"use client";

import * as React from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { Save, Loader2, Copy, Check } from "lucide-react";

import { api } from "@/lib/api";
import { PageHeader } from "@/components/layout/page-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
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

const projectSchema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string().optional(),
  defaultRetentionDays: z.number().min(1).max(365),
  publicDashboard: z.boolean(),
});

type ProjectFormData = z.infer<typeof projectSchema>;

export default function ProjectSettingsPage() {
  const queryClient = useQueryClient();
  const [copied, setCopied] = React.useState(false);

  const { data: project, isLoading } = useQuery({
    queryKey: ["project-settings"],
    queryFn: () => api.project.getSettings(),
  });

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors, isDirty },
  } = useForm<ProjectFormData>({
    resolver: zodResolver(projectSchema),
  });

  React.useEffect(() => {
    if (project) {
      reset({
        name: project.name,
        description: project.description || "",
        defaultRetentionDays: project.defaultRetentionDays,
        publicDashboard: project.publicDashboard,
      });
    }
  }, [project, reset]);

  const updateMutation = useMutation({
    mutationFn: (data: ProjectFormData) => api.project.updateSettings(data),
    onSuccess: () => {
      toast.success("Project settings updated");
      queryClient.invalidateQueries({ queryKey: ["project-settings"] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to update settings");
    },
  });

  const onSubmit = (data: ProjectFormData) => {
    updateMutation.mutate(data);
  };

  const copyProjectId = () => {
    if (project?.id) {
      navigator.clipboard.writeText(project.id);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <PageHeader title="Project Settings" description="Configure your project settings." />
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-32" />
            <Skeleton className="h-4 w-64" />
          </CardHeader>
          <CardContent className="space-y-4">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-20 w-full" />
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Project Settings"
        description="Configure your project settings and preferences."
      />

      <div className="grid gap-6">
        {/* Project ID */}
        <Card>
          <CardHeader>
            <CardTitle>Project ID</CardTitle>
            <CardDescription>
              Your unique project identifier for API access.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <code className="flex-1 px-3 py-2 bg-muted rounded-md text-sm font-mono">
                {project?.id}
              </code>
              <Button variant="outline" size="icon" onClick={copyProjectId}>
                {copied ? (
                  <Check className="h-4 w-4 text-green-500" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* General settings */}
        <Card>
          <CardHeader>
            <CardTitle>General</CardTitle>
            <CardDescription>
              Basic project information and settings.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Project Name</Label>
                <Input id="name" {...register("name")} />
                {errors.name && (
                  <p className="text-sm text-destructive">
                    {errors.name.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  placeholder="Describe your project..."
                  rows={3}
                  {...register("description")}
                />
              </div>

              <div className="flex justify-end">
                <Button
                  type="submit"
                  disabled={!isDirty || updateMutation.isPending}
                >
                  {updateMutation.isPending && (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  )}
                  <Save className="h-4 w-4 mr-2" />
                  Save Changes
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        {/* Data retention */}
        <Card>
          <CardHeader>
            <CardTitle>Data Retention</CardTitle>
            <CardDescription>
              Configure how long data is retained before deletion.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>Default Retention Period</Label>
              <Select
                value={watch("defaultRetentionDays")?.toString()}
                onValueChange={(value) =>
                  setValue("defaultRetentionDays", parseInt(value), {
                    shouldDirty: true,
                  })
                }
              >
                <SelectTrigger className="w-[200px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="7">7 days</SelectItem>
                  <SelectItem value="30">30 days</SelectItem>
                  <SelectItem value="90">90 days</SelectItem>
                  <SelectItem value="180">180 days</SelectItem>
                  <SelectItem value="365">365 days</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Traces older than this will be automatically deleted.
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Public access */}
        <Card>
          <CardHeader>
            <CardTitle>Public Access</CardTitle>
            <CardDescription>
              Configure public access to your project.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Public Dashboard</p>
                <p className="text-sm text-muted-foreground">
                  Allow anyone with the link to view your dashboard.
                </p>
              </div>
              <Switch
                checked={watch("publicDashboard")}
                onCheckedChange={(checked) =>
                  setValue("publicDashboard", checked, { shouldDirty: true })
                }
              />
            </div>
            {watch("publicDashboard") && project?.publicUrl && (
              <div className="flex items-center gap-2 mt-2">
                <Input
                  readOnly
                  value={project.publicUrl}
                  className="font-mono text-sm"
                />
                <Button variant="outline" size="icon">
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Danger zone */}
        <Card className="border-destructive">
          <CardHeader>
            <CardTitle className="text-destructive">Danger Zone</CardTitle>
            <CardDescription>
              Irreversible actions for your project.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Delete All Traces</p>
                <p className="text-sm text-muted-foreground">
                  Permanently delete all traces and observations.
                </p>
              </div>
              <Button variant="destructive" size="sm">
                Delete All Traces
              </Button>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Delete Project</p>
                <p className="text-sm text-muted-foreground">
                  Permanently delete the project and all data.
                </p>
              </div>
              <Button variant="destructive" size="sm">
                Delete Project
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
