"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { Play, Loader2 } from "lucide-react";

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

const runExperimentSchema = z.object({
  name: z.string().min(1, "Name is required"),
  evaluatorId: z.string().optional(),
  metadata: z.string().optional(),
});

type RunExperimentFormData = z.infer<typeof runExperimentSchema>;

interface RunExperimentDialogProps {
  datasetId: string;
  datasetName: string;
}

export function RunExperimentDialog({ datasetId, datasetName }: RunExperimentDialogProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [open, setOpen] = React.useState(false);

  const { data: evaluators } = useQuery({
    queryKey: ["evaluators"],
    queryFn: () => api.evaluators.list(),
    enabled: open,
  });

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<RunExperimentFormData>({
    resolver: zodResolver(runExperimentSchema),
    defaultValues: {
      name: `${datasetName} Run ${new Date().toLocaleDateString()}`,
    },
  });

  const runMutation = useMutation({
    mutationFn: (data: RunExperimentFormData) => {
      let metadata: any;
      if (data.metadata) {
        try {
          metadata = JSON.parse(data.metadata);
        } catch {
          throw new Error("Metadata must be valid JSON");
        }
      }

      return api.datasets.createRun(datasetId, {
        name: data.name,
        evaluatorId: data.evaluatorId,
        metadata,
      });
    },
    onSuccess: (run) => {
      toast.success("Experiment started");
      queryClient.invalidateQueries({ queryKey: ["dataset-runs", datasetId] });
      setOpen(false);
      reset();
      router.push(`/datasets/${datasetId}/runs/${run.id}`);
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to start experiment");
    },
  });

  const onSubmit = (data: RunExperimentFormData) => {
    runMutation.mutate(data);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Play className="h-4 w-4 mr-2" />
          Run Experiment
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Run Experiment</DialogTitle>
          <DialogDescription>
            Run all dataset items through your pipeline and optionally evaluate results.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)}>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">Run Name</Label>
              <Input
                id="name"
                placeholder="My Experiment Run"
                {...register("name")}
              />
              {errors.name && (
                <p className="text-sm text-destructive">{errors.name.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="evaluator">Evaluator (optional)</Label>
              <Select
                onValueChange={(value) => setValue("evaluatorId", value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select an evaluator" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">No evaluator</SelectItem>
                  {evaluators?.map((evaluator) => (
                    <SelectItem key={evaluator.id} value={evaluator.id}>
                      {evaluator.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Automatically score results with an evaluator
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="metadata">Metadata (optional)</Label>
              <Textarea
                id="metadata"
                placeholder='{"model": "gpt-4", "temperature": 0.7}'
                rows={3}
                className="font-mono text-sm"
                {...register("metadata")}
              />
              {errors.metadata && (
                <p className="text-sm text-destructive">{errors.metadata.message}</p>
              )}
              <p className="text-xs text-muted-foreground">
                Additional metadata to attach to this run
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={runMutation.isPending}>
              {runMutation.isPending && (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              )}
              Start Experiment
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
