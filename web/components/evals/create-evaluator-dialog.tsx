"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { Plus, Loader2 } from "lucide-react";

import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { EvaluatorTemplatePicker } from "./evaluator-template-picker";

const createEvaluatorSchema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string().optional(),
  type: z.enum(["LLM_AS_JUDGE", "CODE", "HUMAN"]),
  scoreName: z.string().min(1, "Score name is required"),
  template: z.string().optional(),
});

type CreateEvaluatorFormData = z.infer<typeof createEvaluatorSchema>;

export function CreateEvaluatorDialog() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [open, setOpen] = React.useState(false);
  const [step, setStep] = React.useState<"type" | "details">("type");

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    formState: { errors },
  } = useForm<CreateEvaluatorFormData>({
    resolver: zodResolver(createEvaluatorSchema),
    defaultValues: {
      type: "LLM_AS_JUDGE",
    },
  });

  const selectedType = watch("type");
  const selectedTemplate = watch("template");

  const createMutation = useMutation({
    mutationFn: (data: CreateEvaluatorFormData) =>
      api.evaluators.create({
        name: data.name,
        description: data.description,
        type: data.type,
        scoreName: data.scoreName,
        template: data.template,
      }),
    onSuccess: (evaluator) => {
      toast.success("Evaluator created successfully");
      queryClient.invalidateQueries({ queryKey: ["evaluators"] });
      setOpen(false);
      reset();
      setStep("type");
      router.push(`/evals/${evaluator.id}`);
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to create evaluator");
    },
  });

  const onSubmit = (data: CreateEvaluatorFormData) => {
    createMutation.mutate(data);
  };

  const handleTemplateSelect = (templateId: string, templateName: string, scoreName: string) => {
    setValue("template", templateId);
    setValue("name", templateName);
    setValue("scoreName", scoreName);
    setStep("details");
  };

  const handleOpenChange = (isOpen: boolean) => {
    setOpen(isOpen);
    if (!isOpen) {
      reset();
      setStep("type");
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="h-4 w-4 mr-2" />
          Create Evaluator
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Create Evaluator</DialogTitle>
          <DialogDescription>
            {step === "type"
              ? "Choose an evaluator type or start from a template."
              : "Configure your evaluator settings."}
          </DialogDescription>
        </DialogHeader>

        {step === "type" ? (
          <div className="py-4">
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Evaluator Type</Label>
                <Select
                  value={selectedType}
                  onValueChange={(value) =>
                    setValue("type", value as "LLM_AS_JUDGE" | "CODE" | "HUMAN")
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="LLM_AS_JUDGE">
                      LLM-as-Judge - Use an LLM to evaluate outputs
                    </SelectItem>
                    <SelectItem value="CODE">
                      Code - Use custom code for evaluation
                    </SelectItem>
                    <SelectItem value="HUMAN">
                      Human - Manual annotation queue
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {selectedType === "LLM_AS_JUDGE" && (
                <EvaluatorTemplatePicker onSelect={handleTemplateSelect} />
              )}

              <div className="flex justify-end gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setOpen(false)}
                >
                  Cancel
                </Button>
                <Button
                  type="button"
                  onClick={() => setStep("details")}
                >
                  Continue
                </Button>
              </div>
            </div>
          </div>
        ) : (
          <form onSubmit={handleSubmit(onSubmit)}>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  placeholder="My Evaluator"
                  {...register("name")}
                />
                {errors.name && (
                  <p className="text-sm text-destructive">{errors.name.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description (optional)</Label>
                <Textarea
                  id="description"
                  placeholder="Describe what this evaluator does..."
                  rows={2}
                  {...register("description")}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="scoreName">Score Name</Label>
                <Input
                  id="scoreName"
                  placeholder="accuracy"
                  {...register("scoreName")}
                />
                {errors.scoreName && (
                  <p className="text-sm text-destructive">
                    {errors.scoreName.message}
                  </p>
                )}
                <p className="text-xs text-muted-foreground">
                  The name for scores produced by this evaluator (e.g., accuracy,
                  relevance, hallucination)
                </p>
              </div>
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setStep("type")}
              >
                Back
              </Button>
              <Button type="submit" disabled={createMutation.isPending}>
                {createMutation.isPending && (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Create Evaluator
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
