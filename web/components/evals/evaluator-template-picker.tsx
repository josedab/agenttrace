"use client";

import * as React from "react";
import { Check, Target, FileCheck, AlertTriangle, MessageSquare, Sparkles } from "lucide-react";

import { cn } from "@/lib/utils";
import { Label } from "@/components/ui/label";

interface EvaluatorTemplate {
  id: string;
  name: string;
  description: string;
  scoreName: string;
  icon: React.ComponentType<{ className?: string }>;
}

const templates: EvaluatorTemplate[] = [
  {
    id: "relevance",
    name: "Relevance",
    description: "Evaluate if the output is relevant to the input query",
    scoreName: "relevance",
    icon: Target,
  },
  {
    id: "factuality",
    name: "Factuality",
    description: "Check if the output contains factual information",
    scoreName: "factuality",
    icon: FileCheck,
  },
  {
    id: "hallucination",
    name: "Hallucination Detection",
    description: "Detect hallucinations or made-up information",
    scoreName: "hallucination",
    icon: AlertTriangle,
  },
  {
    id: "helpfulness",
    name: "Helpfulness",
    description: "Evaluate how helpful the response is to the user",
    scoreName: "helpfulness",
    icon: MessageSquare,
  },
  {
    id: "coherence",
    name: "Coherence",
    description: "Assess the logical flow and consistency of the output",
    scoreName: "coherence",
    icon: Sparkles,
  },
];

interface EvaluatorTemplatePickerProps {
  onSelect: (templateId: string, name: string, scoreName: string) => void;
}

export function EvaluatorTemplatePicker({ onSelect }: EvaluatorTemplatePickerProps) {
  const [selectedId, setSelectedId] = React.useState<string | null>(null);

  const handleSelect = (template: EvaluatorTemplate) => {
    setSelectedId(template.id);
    onSelect(template.id, template.name, template.scoreName);
  };

  return (
    <div className="space-y-3">
      <Label>Templates (optional)</Label>
      <div className="grid grid-cols-1 gap-2">
        {templates.map((template) => {
          const Icon = template.icon;
          const isSelected = selectedId === template.id;
          return (
            <button
              key={template.id}
              type="button"
              onClick={() => handleSelect(template)}
              className={cn(
                "flex items-start gap-3 p-3 text-left rounded-lg border transition-colors",
                isSelected
                  ? "border-primary bg-primary/5"
                  : "border-border hover:border-muted-foreground/50 hover:bg-muted/50"
              )}
            >
              <div
                className={cn(
                  "p-2 rounded-md",
                  isSelected ? "bg-primary/10" : "bg-muted"
                )}
              >
                <Icon
                  className={cn(
                    "h-4 w-4",
                    isSelected ? "text-primary" : "text-muted-foreground"
                  )}
                />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-sm">{template.name}</span>
                  {isSelected && (
                    <Check className="h-4 w-4 text-primary" />
                  )}
                </div>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {template.description}
                </p>
              </div>
            </button>
          );
        })}
      </div>
      <p className="text-xs text-muted-foreground">
        Select a template for pre-configured prompts, or skip to create a custom evaluator.
      </p>
    </div>
  );
}
