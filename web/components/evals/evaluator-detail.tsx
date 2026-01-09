"use client";

import * as React from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { Save, Loader2, Play, Bot, Code, User, Trash2 } from "lucide-react";

import { api } from "@/lib/api";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScoreDistributionChart } from "./score-distribution-chart";
import { EvaluatorRunHistory } from "./evaluator-run-history";

interface Evaluator {
  id: string;
  name: string;
  description?: string;
  type: "LLM_AS_JUDGE" | "CODE" | "HUMAN";
  status: "ACTIVE" | "INACTIVE";
  scoreName: string;
  scoreDataType: "NUMERIC" | "BOOLEAN" | "CATEGORICAL";
  config: {
    model?: string;
    prompt?: string;
    code?: string;
    categories?: string[];
    minValue?: number;
    maxValue?: number;
  };
  createdAt: string;
  updatedAt: string;
}

interface EvaluatorDetailProps {
  evaluator: Evaluator;
}

const evaluatorSchema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string().optional(),
  scoreName: z.string().min(1, "Score name is required"),
  scoreDataType: z.enum(["NUMERIC", "BOOLEAN", "CATEGORICAL"]),
  status: z.enum(["ACTIVE", "INACTIVE"]),
  model: z.string().optional(),
  prompt: z.string().optional(),
  code: z.string().optional(),
});

type EvaluatorFormData = z.infer<typeof evaluatorSchema>;

const evaluatorTypeIcons = {
  LLM_AS_JUDGE: Bot,
  CODE: Code,
  HUMAN: User,
};

const evaluatorTypeLabels = {
  LLM_AS_JUDGE: "LLM-as-Judge",
  CODE: "Code",
  HUMAN: "Human",
};

export function EvaluatorDetail({ evaluator }: EvaluatorDetailProps) {
  const queryClient = useQueryClient();
  const TypeIcon = evaluatorTypeIcons[evaluator.type];

  const { data: stats } = useQuery({
    queryKey: ["evaluator-stats", evaluator.id],
    queryFn: () => api.evaluators.getStats(evaluator.id),
  });

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors, isDirty },
  } = useForm<EvaluatorFormData>({
    resolver: zodResolver(evaluatorSchema),
    defaultValues: {
      name: evaluator.name,
      description: evaluator.description || "",
      scoreName: evaluator.scoreName,
      scoreDataType: evaluator.scoreDataType,
      status: evaluator.status,
      model: evaluator.config.model || "gpt-4",
      prompt: evaluator.config.prompt || "",
      code: evaluator.config.code || "",
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: EvaluatorFormData) =>
      api.evaluators.update(evaluator.id, {
        name: data.name,
        description: data.description,
        scoreName: data.scoreName,
        scoreDataType: data.scoreDataType,
        status: data.status,
        config: {
          model: data.model,
          prompt: data.prompt,
          code: data.code,
        },
      }),
    onSuccess: () => {
      toast.success("Evaluator updated successfully");
      queryClient.invalidateQueries({ queryKey: ["evaluator", evaluator.id] });
      queryClient.invalidateQueries({ queryKey: ["evaluators"] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to update evaluator");
    },
  });

  const runMutation = useMutation({
    mutationFn: () => api.evaluators.run(evaluator.id),
    onSuccess: () => {
      toast.success("Evaluation started");
      queryClient.invalidateQueries({ queryKey: ["evaluator-stats", evaluator.id] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to start evaluation");
    },
  });

  const onSubmit = (data: EvaluatorFormData) => {
    updateMutation.mutate(data);
  };

  const scoreDataType = watch("scoreDataType");
  const status = watch("status");

  return (
    <div className="space-y-6">
      {/* Type badge and actions */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Badge variant="outline" className="gap-1.5">
            <TypeIcon className="h-3 w-3" />
            {evaluatorTypeLabels[evaluator.type]}
          </Badge>
          <Badge variant={evaluator.status === "ACTIVE" ? "default" : "secondary"}>
            {evaluator.status}
          </Badge>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => runMutation.mutate()}
            disabled={runMutation.isPending || evaluator.status !== "ACTIVE"}
          >
            {runMutation.isPending ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Play className="h-4 w-4 mr-2" />
            )}
            Run Now
          </Button>
        </div>
      </div>

      <Tabs defaultValue="config" className="space-y-4">
        <TabsList>
          <TabsTrigger value="config">Configuration</TabsTrigger>
          <TabsTrigger value="stats">Statistics</TabsTrigger>
          <TabsTrigger value="history">Run History</TabsTrigger>
        </TabsList>

        <TabsContent value="config" className="space-y-4">
          <form onSubmit={handleSubmit(onSubmit)}>
            <Card>
              <CardHeader>
                <CardTitle>General Settings</CardTitle>
                <CardDescription>
                  Configure the basic settings for this evaluator.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="name">Name</Label>
                    <Input id="name" {...register("name")} />
                    {errors.name && (
                      <p className="text-sm text-destructive">
                        {errors.name.message}
                      </p>
                    )}
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="scoreName">Score Name</Label>
                    <Input id="scoreName" {...register("scoreName")} />
                    {errors.scoreName && (
                      <p className="text-sm text-destructive">
                        {errors.scoreName.message}
                      </p>
                    )}
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="description">Description</Label>
                  <Textarea
                    id="description"
                    rows={2}
                    {...register("description")}
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Score Data Type</Label>
                    <Select
                      value={scoreDataType}
                      onValueChange={(value) =>
                        setValue("scoreDataType", value as any, { shouldDirty: true })
                      }
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="NUMERIC">Numeric (0-1)</SelectItem>
                        <SelectItem value="BOOLEAN">Boolean (true/false)</SelectItem>
                        <SelectItem value="CATEGORICAL">Categorical</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label>Status</Label>
                    <div className="flex items-center gap-3 h-10">
                      <Switch
                        checked={status === "ACTIVE"}
                        onCheckedChange={(checked) =>
                          setValue("status", checked ? "ACTIVE" : "INACTIVE", {
                            shouldDirty: true,
                          })
                        }
                      />
                      <span className="text-sm">
                        {status === "ACTIVE" ? "Active" : "Inactive"}
                      </span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            {evaluator.type === "LLM_AS_JUDGE" && (
              <Card className="mt-4">
                <CardHeader>
                  <CardTitle>LLM Configuration</CardTitle>
                  <CardDescription>
                    Configure the LLM model and prompt for evaluation.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <Label>Model</Label>
                    <Select
                      value={watch("model")}
                      onValueChange={(value) =>
                        setValue("model", value, { shouldDirty: true })
                      }
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="gpt-4">GPT-4</SelectItem>
                        <SelectItem value="gpt-4-turbo">GPT-4 Turbo</SelectItem>
                        <SelectItem value="gpt-3.5-turbo">GPT-3.5 Turbo</SelectItem>
                        <SelectItem value="claude-3-opus">Claude 3 Opus</SelectItem>
                        <SelectItem value="claude-3-sonnet">Claude 3 Sonnet</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="prompt">Evaluation Prompt</Label>
                    <Textarea
                      id="prompt"
                      rows={10}
                      className="font-mono text-sm"
                      placeholder="You are an evaluator. Given the following input and output, evaluate..."
                      {...register("prompt")}
                    />
                    <p className="text-xs text-muted-foreground">
                      Use {`{{input}}`}, {`{{output}}`}, and {`{{expected}}`} as
                      placeholders.
                    </p>
                  </div>
                </CardContent>
              </Card>
            )}

            {evaluator.type === "CODE" && (
              <Card className="mt-4">
                <CardHeader>
                  <CardTitle>Code Configuration</CardTitle>
                  <CardDescription>
                    Write custom JavaScript code for evaluation.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="code">Evaluation Code</Label>
                    <Textarea
                      id="code"
                      rows={15}
                      className="font-mono text-sm"
                      placeholder={`// Return a score between 0 and 1
function evaluate(input, output, expected) {
  // Your evaluation logic here
  return output === expected ? 1 : 0;
}`}
                      {...register("code")}
                    />
                    <p className="text-xs text-muted-foreground">
                      The function receives input, output, and expected values.
                      Return a score between 0 and 1.
                    </p>
                  </div>
                </CardContent>
              </Card>
            )}

            <div className="flex justify-end gap-2 mt-4">
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
        </TabsContent>

        <TabsContent value="stats" className="space-y-4">
          {stats ? (
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <Card>
                <CardHeader className="pb-2">
                  <CardDescription>Total Evaluations</CardDescription>
                </CardHeader>
                <CardContent>
                  <span className="text-2xl font-bold">{stats.totalCount}</span>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className="pb-2">
                  <CardDescription>Average Score</CardDescription>
                </CardHeader>
                <CardContent>
                  <span className="text-2xl font-bold">
                    {stats.avgScore?.toFixed(2) || "-"}
                  </span>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className="pb-2">
                  <CardDescription>Last 24h</CardDescription>
                </CardHeader>
                <CardContent>
                  <span className="text-2xl font-bold">{stats.last24hCount}</span>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className="pb-2">
                  <CardDescription>Pass Rate</CardDescription>
                </CardHeader>
                <CardContent>
                  <span className="text-2xl font-bold">
                    {stats.passRate !== undefined
                      ? `${(stats.passRate * 100).toFixed(0)}%`
                      : "-"}
                  </span>
                </CardContent>
              </Card>
            </div>
          ) : (
            <div className="flex items-center justify-center py-12 text-muted-foreground">
              No statistics available yet
            </div>
          )}

          {stats?.scoreDistribution && (
            <Card>
              <CardHeader>
                <CardTitle>Score Distribution</CardTitle>
              </CardHeader>
              <CardContent>
                <ScoreDistributionChart data={stats.scoreDistribution} />
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="history">
          <EvaluatorRunHistory evaluatorId={evaluator.id} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
