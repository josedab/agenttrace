"use client";

import * as React from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { Plus, Loader2 } from "lucide-react";

import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
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

const addItemSchema = z.object({
  input: z.string().min(1, "Input is required"),
  expectedOutput: z.string().optional(),
  metadata: z.string().optional(),
});

type AddItemFormData = z.infer<typeof addItemSchema>;

interface AddItemDialogProps {
  datasetId: string;
}

export function AddItemDialog({ datasetId }: AddItemDialogProps) {
  const queryClient = useQueryClient();
  const [open, setOpen] = React.useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<AddItemFormData>({
    resolver: zodResolver(addItemSchema),
  });

  const addMutation = useMutation({
    mutationFn: (data: AddItemFormData) => {
      let input: any;
      let expectedOutput: any;
      let metadata: any;

      try {
        input = JSON.parse(data.input);
      } catch {
        input = data.input;
      }

      if (data.expectedOutput) {
        try {
          expectedOutput = JSON.parse(data.expectedOutput);
        } catch {
          expectedOutput = data.expectedOutput;
        }
      }

      if (data.metadata) {
        try {
          metadata = JSON.parse(data.metadata);
        } catch {
          throw new Error("Metadata must be valid JSON");
        }
      }

      return api.datasets.addItem(datasetId, {
        input,
        expectedOutput,
        metadata,
      });
    },
    onSuccess: () => {
      toast.success("Item added successfully");
      queryClient.invalidateQueries({ queryKey: ["dataset-items", datasetId] });
      queryClient.invalidateQueries({ queryKey: ["dataset", datasetId] });
      setOpen(false);
      reset();
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to add item");
    },
  });

  const onSubmit = (data: AddItemFormData) => {
    addMutation.mutate(data);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">
          <Plus className="h-4 w-4 mr-2" />
          Add Item
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Add Dataset Item</DialogTitle>
          <DialogDescription>
            Add a new test case to the dataset.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)}>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="input">Input</Label>
              <Textarea
                id="input"
                placeholder='{"query": "What is the capital of France?"}'
                rows={4}
                className="font-mono text-sm"
                {...register("input")}
              />
              {errors.input && (
                <p className="text-sm text-destructive">{errors.input.message}</p>
              )}
              <p className="text-xs text-muted-foreground">
                Enter JSON or plain text
              </p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="expectedOutput">Expected Output (optional)</Label>
              <Textarea
                id="expectedOutput"
                placeholder='"Paris"'
                rows={3}
                className="font-mono text-sm"
                {...register("expectedOutput")}
              />
              <p className="text-xs text-muted-foreground">
                The expected result for evaluation
              </p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="metadata">Metadata (optional)</Label>
              <Textarea
                id="metadata"
                placeholder='{"category": "geography", "difficulty": "easy"}'
                rows={2}
                className="font-mono text-sm"
                {...register("metadata")}
              />
              {errors.metadata && (
                <p className="text-sm text-destructive">{errors.metadata.message}</p>
              )}
              <p className="text-xs text-muted-foreground">
                Additional metadata as JSON
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
            <Button type="submit" disabled={addMutation.isPending}>
              {addMutation.isPending && (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              )}
              Add Item
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
