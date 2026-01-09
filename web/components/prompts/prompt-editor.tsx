"use client";

import * as React from "react";
import Link from "next/link";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { Save, Play, History, Tag, Loader2, AlertCircle, Copy } from "lucide-react";

import { api } from "@/lib/api";
import { extractPromptVariables } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import { PromptVersionList } from "@/components/prompts/prompt-version-list";
import { PromptLabels } from "@/components/prompts/prompt-labels";

const promptSchema = z.object({
  prompt: z.string().min(1, "Prompt content is required"),
  config: z.string().optional(),
});

type PromptFormData = z.infer<typeof promptSchema>;

interface PromptEditorProps {
  promptName: string;
}

export function PromptEditor({ promptName }: PromptEditorProps) {
  const queryClient = useQueryClient();

  const { data: prompt, isLoading, error } = useQuery({
    queryKey: ["prompt", promptName],
    queryFn: () => api.prompts.getByName(promptName),
  });

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isDirty },
  } = useForm<PromptFormData>({
    resolver: zodResolver(promptSchema),
    defaultValues: {
      prompt: "",
      config: "",
    },
  });

  // Update form when prompt loads
  React.useEffect(() => {
    if (prompt?.activeVersion) {
      setValue("prompt", prompt.activeVersion.prompt || "");
      setValue("config", prompt.activeVersion.config ? JSON.stringify(prompt.activeVersion.config, null, 2) : "");
    }
  }, [prompt, setValue]);

  const promptValue = watch("prompt");
  const variables = React.useMemo(
    () => extractPromptVariables(promptValue || ""),
    [promptValue]
  );

  const saveMutation = useMutation({
    mutationFn: (data: PromptFormData) =>
      api.prompts.createVersion(promptName, {
        prompt: data.prompt,
        config: data.config ? JSON.parse(data.config) : undefined,
      }),
    onSuccess: () => {
      toast.success("New version saved");
      queryClient.invalidateQueries({ queryKey: ["prompt", promptName] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to save version");
    },
  });

  const onSubmit = (data: PromptFormData) => {
    try {
      if (data.config) {
        JSON.parse(data.config); // Validate JSON
      }
      saveMutation.mutate(data);
    } catch {
      toast.error("Invalid JSON in config");
    }
  };

  if (isLoading) {
    return <PromptEditorSkeleton />;
  }

  if (error || !prompt) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <AlertCircle className="h-12 w-12 text-destructive mb-4" />
        <p className="text-destructive">Failed to load prompt</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold">{prompt.name}</h1>
          {prompt.description && (
            <p className="text-muted-foreground mt-1">{prompt.description}</p>
          )}
          <div className="flex items-center gap-2 mt-2">
            <Badge variant="outline">v{prompt.activeVersion?.version || 1}</Badge>
            {prompt.labels && Object.entries(prompt.labels).map(([label, version]) => (
              <Badge key={label} variant="secondary">
                {label}: v{version}
              </Badge>
            ))}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" asChild>
            <Link href={`/prompts/${encodeURIComponent(promptName)}/playground`}>
              <Play className="h-4 w-4 mr-2" />
              Playground
            </Link>
          </Button>
          <Button onClick={handleSubmit(onSubmit)} disabled={saveMutation.isPending || !isDirty}>
            {saveMutation.isPending ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Save className="h-4 w-4 mr-2" />
            )}
            Save Version
          </Button>
        </div>
      </div>

      <div className="grid lg:grid-cols-3 gap-6">
        {/* Main editor */}
        <div className="lg:col-span-2">
          <Tabs defaultValue="editor">
            <TabsList>
              <TabsTrigger value="editor">Editor</TabsTrigger>
              <TabsTrigger value="config">Config</TabsTrigger>
              <TabsTrigger value="versions">
                <History className="h-4 w-4 mr-1" />
                Versions
              </TabsTrigger>
            </TabsList>

            <TabsContent value="editor" className="mt-4">
              <Card>
                <CardHeader>
                  <CardTitle>Prompt Template</CardTitle>
                  <CardDescription>
                    Use {"{{variable}}"} syntax for template variables
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Textarea
                    {...register("prompt")}
                    rows={15}
                    className="font-mono text-sm"
                    placeholder="You are a helpful assistant. User query: {{input}}"
                  />
                  {errors.prompt && (
                    <p className="text-sm text-destructive mt-2">
                      {errors.prompt.message}
                    </p>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="config" className="mt-4">
              <Card>
                <CardHeader>
                  <CardTitle>Model Configuration</CardTitle>
                  <CardDescription>
                    Optional JSON configuration for model parameters
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Textarea
                    {...register("config")}
                    rows={10}
                    className="font-mono text-sm"
                    placeholder={`{
  "model": "gpt-4",
  "temperature": 0.7,
  "max_tokens": 1000
}`}
                  />
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="versions" className="mt-4">
              <PromptVersionList promptName={promptName} />
            </TabsContent>
          </Tabs>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Variables */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Variables</CardTitle>
              <CardDescription>
                Detected template variables
              </CardDescription>
            </CardHeader>
            <CardContent>
              {variables.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No variables detected
                </p>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {variables.map((variable) => (
                    <Badge key={variable} variant="outline">
                      {variable}
                    </Badge>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Labels */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <Tag className="h-4 w-4" />
                Labels
              </CardTitle>
              <CardDescription>
                Assign versions to labels
              </CardDescription>
            </CardHeader>
            <CardContent>
              <PromptLabels promptName={promptName} />
            </CardContent>
          </Card>

          {/* Usage */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Usage</CardTitle>
              <CardDescription>
                Fetch this prompt in your code
              </CardDescription>
            </CardHeader>
            <CardContent>
              <CodeBlock
                code={`from agenttrace import AgentTrace

at = AgentTrace()
prompt = at.get_prompt("${promptName}")

# With variables
compiled = prompt.compile(input="Hello")`}
                language="python"
              />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function CodeBlock({ code, language }: { code: string; language: string }) {
  const copyCode = () => {
    navigator.clipboard.writeText(code);
    toast.success("Code copied to clipboard");
  };

  return (
    <div className="relative">
      <pre className="text-xs bg-muted p-3 rounded-lg overflow-auto">
        <code>{code}</code>
      </pre>
      <Button
        variant="ghost"
        size="sm"
        className="absolute top-2 right-2 h-6 w-6 p-0"
        onClick={copyCode}
      >
        <Copy className="h-3 w-3" />
      </Button>
    </div>
  );
}

export function PromptEditorSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <div className="h-8 w-48 bg-muted animate-pulse rounded" />
          <div className="h-4 w-64 bg-muted animate-pulse rounded mt-2" />
        </div>
        <div className="flex gap-2">
          <div className="h-10 w-32 bg-muted animate-pulse rounded" />
          <div className="h-10 w-32 bg-muted animate-pulse rounded" />
        </div>
      </div>
      <div className="grid lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <div className="h-96 bg-muted animate-pulse rounded" />
        </div>
        <div className="space-y-6">
          <div className="h-32 bg-muted animate-pulse rounded" />
          <div className="h-32 bg-muted animate-pulse rounded" />
        </div>
      </div>
    </div>
  );
}
