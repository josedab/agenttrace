"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  usePlaygroundTemplates,
  useExecutePlayground,
  type PlaygroundResult,
  type PlaygroundTemplate,
} from "@/hooks/use-eval-playground";
import { Play, Share2, FileCode, CheckCircle, XCircle } from "lucide-react";

export function EvalPlayground() {
  const [code, setCode] = useState(
    'function evaluate(trace) {\n  // Access trace properties: trace.input, trace.output, trace.totalCost, trace.durationMs\n  const hasOutput = trace.output && trace.output.length > 0;\n  return {\n    score: hasOutput ? 1.0 : 0.0,\n    label: hasOutput ? "pass" : "fail",\n    reasoning: hasOutput ? "Trace has output" : "No output found"\n  };\n}'
  );
  const [language, setLanguage] = useState<"javascript" | "python">("javascript");
  const [traceIds, setTraceIds] = useState("");
  const [results, setResults] = useState<PlaygroundResult[]>([]);

  const { data: templatesData } = usePlaygroundTemplates();
  const executeMutation = useExecutePlayground();
  const templates = templatesData?.templates ?? [];

  const handleRun = async () => {
    const ids = traceIds
      .split(",")
      .map((id) => id.trim())
      .filter(Boolean);
    if (ids.length === 0) return;

    try {
      const response = await executeMutation.mutateAsync({
        code,
        language,
        traceIds: ids,
      });
      setResults(response?.results ?? []);
    } catch {
      // Error handled by mutation state
    }
  };

  const loadTemplate = (template: PlaygroundTemplate) => {
    setCode(template.code);
    setLanguage(template.language);
  };

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Editor Panel */}
        <div className="lg:col-span-2 space-y-4">
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <FileCode className="h-5 w-5" />
                  Evaluator Code
                </CardTitle>
                <div className="flex gap-2">
                  <Badge
                    variant={language === "javascript" ? "default" : "outline"}
                    className="cursor-pointer"
                    onClick={() => setLanguage("javascript")}
                  >
                    JavaScript
                  </Badge>
                  <Badge
                    variant={language === "python" ? "default" : "outline"}
                    className="cursor-pointer"
                    onClick={() => setLanguage("python")}
                  >
                    Python
                  </Badge>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <textarea
                value={code}
                onChange={(e) => setCode(e.target.value)}
                className="w-full h-64 font-mono text-sm bg-gray-900 text-green-400 p-4 rounded-lg border resize-none focus:outline-none focus:ring-2 focus:ring-primary"
                spellCheck={false}
              />
              <div className="flex gap-2 mt-3">
                <Input
                  placeholder="Enter trace IDs (comma-separated)..."
                  value={traceIds}
                  onChange={(e) => setTraceIds(e.target.value)}
                  className="flex-1"
                />
                <Button
                  onClick={handleRun}
                  disabled={executeMutation.isPending || !traceIds}
                >
                  <Play className="h-4 w-4 mr-1" />
                  {executeMutation.isPending ? "Running..." : "Run"}
                </Button>
                <Button variant="outline" size="icon">
                  <Share2 className="h-4 w-4" />
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Results */}
          {results.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">
                  Results ({results.length} traces evaluated)
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  {results.map((result, i) => (
                    <div
                      key={i}
                      className="flex items-center justify-between p-3 rounded-lg border"
                    >
                      <div className="flex items-center gap-3">
                        {result.error ? (
                          <XCircle className="h-5 w-5 text-red-500" />
                        ) : result.label === "pass" ? (
                          <CheckCircle className="h-5 w-5 text-green-500" />
                        ) : (
                          <XCircle className="h-5 w-5 text-yellow-500" />
                        )}
                        <div>
                          <span className="font-mono text-sm">{result.traceId}</span>
                          {result.reasoning && (
                            <p className="text-xs text-muted-foreground mt-0.5">
                              {result.reasoning}
                            </p>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-3">
                        {result.score !== null && result.score !== undefined && (
                          <span className="font-bold text-lg">
                            {(result.score * 100).toFixed(0)}%
                          </span>
                        )}
                        <Badge variant={result.error ? "destructive" : "outline"}>
                          {result.error ? "error" : result.label}
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                          {result.durationMs}ms
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {executeMutation.isError && (
            <Card className="border-red-200">
              <CardContent className="pt-4">
                <p className="text-sm text-red-600">
                  Execution error: {(executeMutation.error as Error).message}
                </p>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Templates Panel */}
        <div>
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Templates</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {templates.map((template: PlaygroundTemplate) => (
                  <div
                    key={template.id}
                    className="p-3 rounded-lg border cursor-pointer hover:bg-accent transition-colors"
                    onClick={() => loadTemplate(template)}
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-sm">{template.name}</span>
                      <Badge variant="outline" className="text-xs">
                        {template.language}
                      </Badge>
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">
                      {template.description}
                    </p>
                    <Badge variant="secondary" className="text-xs mt-1">
                      {template.category}
                    </Badge>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
