"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Plus, X, Loader2 } from "lucide-react";

import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";

interface PromptLabelsProps {
  promptName: string;
}

const COMMON_LABELS = ["production", "staging", "latest", "dev", "test"];

export function PromptLabels({ promptName }: PromptLabelsProps) {
  const queryClient = useQueryClient();
  const [newLabel, setNewLabel] = React.useState("");
  const [selectedVersion, setSelectedVersion] = React.useState<string>("");

  const { data: prompt } = useQuery({
    queryKey: ["prompt", promptName],
    queryFn: () => api.prompts.getByName(promptName),
  });

  const setLabelMutation = useMutation({
    mutationFn: ({ label, version }: { label: string; version: number }) =>
      api.prompts.setLabel(promptName, label, version),
    onSuccess: () => {
      toast.success("Label updated");
      queryClient.invalidateQueries({ queryKey: ["prompt", promptName] });
      setNewLabel("");
      setSelectedVersion("");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to set label");
    },
  });

  const removeLabelMutation = useMutation({
    mutationFn: (label: string) =>
      api.prompts.removeLabel(promptName, label),
    onSuccess: () => {
      toast.success("Label removed");
      queryClient.invalidateQueries({ queryKey: ["prompt", promptName] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to remove label");
    },
  });

  const labels = prompt?.labels || {};
  const versions = prompt?.versions || [];
  const existingLabels = Object.keys(labels);
  const availableLabels = COMMON_LABELS.filter(
    (l) => !existingLabels.includes(l)
  );

  const handleAddLabel = () => {
    if (newLabel && selectedVersion) {
      setLabelMutation.mutate({
        label: newLabel,
        version: parseInt(selectedVersion),
      });
    }
  };

  return (
    <div className="space-y-4">
      {/* Existing labels */}
      {existingLabels.length > 0 ? (
        <div className="space-y-2">
          {existingLabels.map((label) => (
            <div
              key={label}
              className="flex items-center justify-between p-2 bg-muted rounded"
            >
              <div className="flex items-center gap-2">
                <Badge variant="secondary">{label}</Badge>
                <span className="text-sm text-muted-foreground">
                  → v{labels[label]}
                </span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0"
                onClick={() => removeLabelMutation.mutate(label)}
                disabled={removeLabelMutation.isPending}
              >
                {removeLabelMutation.isPending ? (
                  <Loader2 className="h-3 w-3 animate-spin" />
                ) : (
                  <X className="h-3 w-3" />
                )}
              </Button>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">No labels assigned</p>
      )}

      {/* Add new label */}
      <div className="space-y-2 pt-2 border-t">
        <div className="flex gap-2">
          <Select value={newLabel} onValueChange={setNewLabel}>
            <SelectTrigger className="flex-1">
              <SelectValue placeholder="Label" />
            </SelectTrigger>
            <SelectContent>
              {availableLabels.map((label) => (
                <SelectItem key={label} value={label}>
                  {label}
                </SelectItem>
              ))}
              <SelectItem value="_custom">Custom...</SelectItem>
            </SelectContent>
          </Select>
          <Select value={selectedVersion} onValueChange={setSelectedVersion}>
            <SelectTrigger className="w-24">
              <SelectValue placeholder="Ver" />
            </SelectTrigger>
            <SelectContent>
              {versions.map((v) => (
                <SelectItem key={v.version} value={v.version.toString()}>
                  v{v.version}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {newLabel === "_custom" && (
          <Input
            placeholder="Custom label name"
            value=""
            onChange={(e) => setNewLabel(e.target.value)}
          />
        )}

        <Button
          size="sm"
          className="w-full"
          onClick={handleAddLabel}
          disabled={!newLabel || !selectedVersion || setLabelMutation.isPending}
        >
          {setLabelMutation.isPending ? (
            <Loader2 className="h-4 w-4 mr-2 animate-spin" />
          ) : (
            <Plus className="h-4 w-4 mr-2" />
          )}
          Add Label
        </Button>
      </div>
    </div>
  );
}
