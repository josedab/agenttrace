"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { Sparkles, Send, Loader2, ChevronDown, HelpCircle } from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface NLQueryResult {
  query: string;
  interpretedAs: string;
  filter: Record<string, unknown>;
  traces: {
    data: Array<{
      id: string;
      name: string;
      startTime: string;
      durationMs: number;
      totalCost: number;
      level: string;
    }>;
    totalCount: number;
    hasMore: boolean;
  };
  suggestions: string[];
  executionTimeMs: number;
}

interface NaturalQueryProps {
  apiKey: string;
  baseUrl: string;
  onResults?: (results: NLQueryResult) => void;
  className?: string;
}

const EXAMPLE_QUERIES = [
  "Show me traces with errors from the last 24 hours",
  "Find expensive traces that cost more than $0.10",
  "Show slow traces taking more than 5 seconds",
  "Traces tagged with 'production' from this week",
  "Failed agent runs from yesterday",
];

export function NaturalQuery({
  apiKey,
  baseUrl,
  onResults,
  className,
}: NaturalQueryProps) {
  const [query, setQuery] = React.useState("");
  const [isLoading, setIsLoading] = React.useState(false);
  const [result, setResult] = React.useState<NLQueryResult | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [isExamplesOpen, setIsExamplesOpen] = React.useState(false);

  const executeQuery = async (queryText: string) => {
    if (!queryText.trim()) return;

    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch(`${baseUrl}/v1/traces/query/natural`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify({ query: queryText, limit: 50 }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || "Query failed");
      }

      const data: NLQueryResult = await response.json();
      setResult(data);
      onResults?.(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    executeQuery(query);
  };

  const handleExampleClick = (example: string) => {
    setQuery(example);
    executeQuery(example);
    setIsExamplesOpen(false);
  };

  const handleSuggestionClick = (suggestion: string) => {
    setQuery(suggestion);
    executeQuery(suggestion);
  };

  return (
    <div className={cn("space-y-4", className)}>
      {/* Query Input */}
      <form onSubmit={handleSubmit} className="space-y-2">
        <div className="flex items-center gap-2">
          <div className="relative flex-1">
            <Sparkles className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-purple-500" />
            <Input
              placeholder="Ask in natural language, e.g., 'Show me failed traces from yesterday'"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="pl-10 pr-4"
              disabled={isLoading}
            />
          </div>
          <Button type="submit" disabled={isLoading || !query.trim()}>
            {isLoading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Send className="h-4 w-4" />
            )}
          </Button>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => setIsExamplesOpen(!isExamplesOpen)}
                >
                  <HelpCircle className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>View example queries</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>

        {/* Examples Collapsible */}
        <Collapsible open={isExamplesOpen} onOpenChange={setIsExamplesOpen}>
          <CollapsibleContent className="space-y-2">
            <div className="rounded-lg border bg-muted/50 p-3">
              <p className="mb-2 text-sm font-medium text-muted-foreground">
                Example queries:
              </p>
              <div className="flex flex-wrap gap-2">
                {EXAMPLE_QUERIES.map((example, index) => (
                  <button
                    key={index}
                    type="button"
                    onClick={() => handleExampleClick(example)}
                    className="text-sm text-primary hover:underline"
                  >
                    "{example}"
                  </button>
                ))}
              </div>
            </div>
          </CollapsibleContent>
        </Collapsible>
      </form>

      {/* Error Display */}
      {error && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-3">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      {/* Results Display */}
      {result && (
        <div className="space-y-3">
          {/* Interpretation */}
          <div className="rounded-lg border bg-muted/50 p-3">
            <div className="flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-purple-500" />
              <span className="text-sm font-medium">Interpreted as:</span>
            </div>
            <p className="mt-1 text-sm text-muted-foreground">
              {result.interpretedAs}
            </p>
            <div className="mt-2 flex items-center gap-2 text-xs text-muted-foreground">
              <Badge variant="secondary">{result.traces.totalCount} traces</Badge>
              <span>•</span>
              <span>{result.executionTimeMs}ms</span>
            </div>
          </div>

          {/* Applied Filters Preview */}
          {Object.keys(result.filter).length > 0 && (
            <Collapsible>
              <CollapsibleTrigger className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
                <ChevronDown className="h-3 w-3" />
                View applied filters
              </CollapsibleTrigger>
              <CollapsibleContent>
                <pre className="mt-2 overflow-x-auto rounded-lg bg-muted p-3 text-xs">
                  {JSON.stringify(result.filter, null, 2)}
                </pre>
              </CollapsibleContent>
            </Collapsible>
          )}

          {/* Suggestions */}
          {result.suggestions && result.suggestions.length > 0 && (
            <div className="space-y-2">
              <p className="text-sm font-medium text-muted-foreground">
                Related queries:
              </p>
              <div className="flex flex-wrap gap-2">
                {result.suggestions.map((suggestion, index) => (
                  <Button
                    key={index}
                    variant="outline"
                    size="sm"
                    onClick={() => handleSuggestionClick(suggestion)}
                    className="text-xs"
                  >
                    {suggestion}
                  </Button>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// Separate component for inline natural query in trace list
export function NaturalQueryInline({
  onApplyFilter,
}: {
  onApplyFilter: (filter: Record<string, unknown>) => void;
}) {
  const [query, setQuery] = React.useState("");
  const [isLoading, setIsLoading] = React.useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    setIsLoading(true);
    // Implementation would call the API and pass filter to parent
    // This is a simplified version
    setIsLoading(false);
  };

  return (
    <form onSubmit={handleSubmit} className="flex items-center gap-2">
      <div className="relative flex-1">
        <Sparkles className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-purple-500" />
        <Input
          placeholder="Ask in natural language..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="h-8 pl-9 text-sm"
          disabled={isLoading}
        />
      </div>
      <Button
        type="submit"
        size="sm"
        disabled={isLoading || !query.trim()}
        className="h-8"
      >
        {isLoading ? (
          <Loader2 className="h-3 w-3 animate-spin" />
        ) : (
          <Send className="h-3 w-3" />
        )}
      </Button>
    </form>
  );
}
