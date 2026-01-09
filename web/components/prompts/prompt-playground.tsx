"use client";

import * as React from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { Play, Loader2, Copy, RefreshCw } from "lucide-react";

import { api } from "@/lib/api";
import { extractPromptVariables, compilePrompt } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";

interface PromptPlaygroundProps {
  promptName: string;
}

export function PromptPlayground({ promptName }: PromptPlaygroundProps) {
  const [selectedVersion, setSelectedVersion] = React.useState<string>("latest");
  const [variables, setVariables] = React.useState<Record<string, string>>({});
  const [compiledPrompt, setCompiledPrompt] = React.useState<string>("");
  const [output, setOutput] = React.useState<string>("");

  const { data: prompt, isLoading } = useQuery({
    queryKey: ["prompt", promptName],
    queryFn: () => api.prompts.getByName(promptName),
  });

  // Get the prompt template based on selected version
  const promptTemplate = React.useMemo(() => {
    if (!prompt) return "";
    if (selectedVersion === "latest") {
      return prompt.activeVersion?.prompt || "";
    }
    const version = prompt.versions?.find(
      (v) => v.version.toString() === selectedVersion
    );
    return version?.prompt || "";
  }, [prompt, selectedVersion]);

  // Extract variables from template
  const templateVariables = React.useMemo(
    () => extractPromptVariables(promptTemplate),
    [promptTemplate]
  );

  // Initialize variables when template changes
  React.useEffect(() => {
    const newVars: Record<string, string> = {};
    templateVariables.forEach((v) => {
      newVars[v] = variables[v] || "";
    });
    setVariables(newVars);
  }, [templateVariables]);

  // Update compiled prompt when variables change
  React.useEffect(() => {
    if (promptTemplate) {
      setCompiledPrompt(compilePrompt(promptTemplate, variables));
    }
  }, [promptTemplate, variables]);

  const runMutation = useMutation({
    mutationFn: async () => {
      // This would call your backend to actually run the prompt
      const response = await api.prompts.run(promptName, {
        version: selectedVersion === "latest" ? undefined : parseInt(selectedVersion),
        variables,
      });
      return response;
    },
    onSuccess: (data) => {
      setOutput(data.output || "No output");
      toast.success("Prompt executed successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to run prompt");
    },
  });

  const copyCompiled = () => {
    navigator.clipboard.writeText(compiledPrompt);
    toast.success("Copied to clipboard");
  };

  if (isLoading) {
    return <PlaygroundSkeleton />;
  }

  if (!prompt) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Prompt not found</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{prompt.name} - Playground</h1>
          <p className="text-muted-foreground mt-1">
            Test your prompt with different variables
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={selectedVersion} onValueChange={setSelectedVersion}>
            <SelectTrigger className="w-32">
              <SelectValue placeholder="Version" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="latest">Latest</SelectItem>
              {prompt.versions?.map((v) => (
                <SelectItem key={v.version} value={v.version.toString()}>
                  v{v.version}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            onClick={() => runMutation.mutate()}
            disabled={runMutation.isPending}
          >
            {runMutation.isPending ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Play className="h-4 w-4 mr-2" />
            )}
            Run
          </Button>
        </div>
      </div>

      <div className="grid lg:grid-cols-2 gap-6">
        {/* Left side - Variables and Template */}
        <div className="space-y-6">
          {/* Variables */}
          <Card>
            <CardHeader>
              <CardTitle>Variables</CardTitle>
              <CardDescription>
                Fill in the template variables
              </CardDescription>
            </CardHeader>
            <CardContent>
              {templateVariables.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No variables in this template
                </p>
              ) : (
                <div className="space-y-4">
                  {templateVariables.map((variable) => (
                    <div key={variable} className="space-y-2">
                      <Label htmlFor={variable}>
                        <Badge variant="outline" className="mr-2">
                          {variable}
                        </Badge>
                      </Label>
                      <Textarea
                        id={variable}
                        value={variables[variable] || ""}
                        onChange={(e) =>
                          setVariables((prev) => ({
                            ...prev,
                            [variable]: e.target.value,
                          }))
                        }
                        rows={3}
                        placeholder={`Enter value for ${variable}`}
                      />
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Template */}
          <Card>
            <CardHeader>
              <CardTitle>Template</CardTitle>
              <CardDescription>
                Original prompt template
              </CardDescription>
            </CardHeader>
            <CardContent>
              <pre className="text-sm font-mono bg-muted p-4 rounded-lg whitespace-pre-wrap">
                {promptTemplate || "No template"}
              </pre>
            </CardContent>
          </Card>
        </div>

        {/* Right side - Compiled and Output */}
        <div className="space-y-6">
          {/* Compiled prompt */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <div>
                <CardTitle>Compiled Prompt</CardTitle>
                <CardDescription>
                  Preview with variables filled in
                </CardDescription>
              </div>
              <Button variant="ghost" size="sm" onClick={copyCompiled}>
                <Copy className="h-4 w-4" />
              </Button>
            </CardHeader>
            <CardContent>
              <pre className="text-sm font-mono bg-muted p-4 rounded-lg whitespace-pre-wrap max-h-64 overflow-auto">
                {compiledPrompt || "Fill in variables to see compiled prompt"}
              </pre>
            </CardContent>
          </Card>

          {/* Output */}
          <Card>
            <CardHeader>
              <CardTitle>Output</CardTitle>
              <CardDescription>
                Model response
              </CardDescription>
            </CardHeader>
            <CardContent>
              {runMutation.isPending ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
              ) : output ? (
                <pre className="text-sm font-mono bg-muted p-4 rounded-lg whitespace-pre-wrap max-h-96 overflow-auto">
                  {output}
                </pre>
              ) : (
                <p className="text-sm text-muted-foreground text-center py-8">
                  Click &quot;Run&quot; to execute the prompt
                </p>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function PlaygroundSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-48 mt-2" />
        </div>
        <div className="flex gap-2">
          <Skeleton className="h-10 w-32" />
          <Skeleton className="h-10 w-20" />
        </div>
      </div>
      <div className="grid lg:grid-cols-2 gap-6">
        <div className="space-y-6">
          <Skeleton className="h-64 w-full" />
          <Skeleton className="h-48 w-full" />
        </div>
        <div className="space-y-6">
          <Skeleton className="h-48 w-full" />
          <Skeleton className="h-64 w-full" />
        </div>
      </div>
    </div>
  );
}
