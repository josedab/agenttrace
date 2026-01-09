"use client";

import * as React from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  ChevronLeft,
  ChevronRight,
  ThumbsUp,
  ThumbsDown,
  SkipForward,
  Loader2,
  AlertCircle,
} from "lucide-react";

import { api } from "@/lib/api";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";

interface AnnotationQueue {
  id: string;
  name: string;
  scoreName: string;
  scoreDataType: "NUMERIC" | "BOOLEAN" | "CATEGORICAL";
  categories?: string[];
  pendingCount: number;
  completedCount: number;
  totalCount: number;
}

interface AnnotationItem {
  id: string;
  traceId: string;
  input: any;
  output: any;
  expectedOutput?: any;
  metadata?: Record<string, any>;
}

interface AnnotationInterfaceProps {
  queue: AnnotationQueue;
}

export function AnnotationInterface({ queue }: AnnotationInterfaceProps) {
  const queryClient = useQueryClient();
  const [currentIndex, setCurrentIndex] = React.useState(0);
  const [comment, setComment] = React.useState("");
  const [numericScore, setNumericScore] = React.useState<number>(0.5);

  const { data: items, isLoading, error } = useQuery({
    queryKey: ["annotation-items", queue.id],
    queryFn: () => api.annotationQueues.getItems(queue.id),
  });

  const submitMutation = useMutation({
    mutationFn: ({
      itemId,
      score,
      comment,
    }: {
      itemId: string;
      score: number | boolean | string;
      comment?: string;
    }) =>
      api.annotationQueues.submitScore(queue.id, itemId, {
        score,
        comment,
      }),
    onSuccess: () => {
      toast.success("Score submitted");
      queryClient.invalidateQueries({ queryKey: ["annotation-items", queue.id] });
      queryClient.invalidateQueries({ queryKey: ["annotation-queue", queue.id] });
      setComment("");
      // Move to next item
      if (items && currentIndex < items.length - 1) {
        setCurrentIndex(currentIndex + 1);
      }
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to submit score");
    },
  });

  const skipMutation = useMutation({
    mutationFn: (itemId: string) => api.annotationQueues.skipItem(queue.id, itemId),
    onSuccess: () => {
      if (items && currentIndex < items.length - 1) {
        setCurrentIndex(currentIndex + 1);
      }
    },
  });

  if (isLoading) {
    return <AnnotationInterfaceSkeleton />;
  }

  if (error || !items) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <AlertCircle className="h-12 w-12 text-destructive mb-4" />
        <p className="text-destructive">Failed to load annotation items</p>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <ThumbsUp className="h-12 w-12 text-green-500 mb-4" />
          <h3 className="text-lg font-semibold">All caught up!</h3>
          <p className="text-sm text-muted-foreground mt-1">
            No more items pending review.
          </p>
        </CardContent>
      </Card>
    );
  }

  const currentItem = items[currentIndex];
  const progress = ((queue.completedCount) / queue.totalCount) * 100;

  const handleSubmitBoolean = (value: boolean) => {
    submitMutation.mutate({
      itemId: currentItem.id,
      score: value,
      comment: comment || undefined,
    });
  };

  const handleSubmitNumeric = () => {
    submitMutation.mutate({
      itemId: currentItem.id,
      score: numericScore,
      comment: comment || undefined,
    });
  };

  const handleSubmitCategorical = (category: string) => {
    submitMutation.mutate({
      itemId: currentItem.id,
      score: category,
      comment: comment || undefined,
    });
  };

  return (
    <div className="space-y-6">
      {/* Progress bar */}
      <Card>
        <CardContent className="pt-6">
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Progress</span>
              <span>
                {queue.completedCount} / {queue.totalCount} reviewed
              </span>
            </div>
            <Progress value={progress} />
          </div>
        </CardContent>
      </Card>

      {/* Navigation */}
      <div className="flex items-center justify-between">
        <Button
          variant="outline"
          size="sm"
          onClick={() => setCurrentIndex(Math.max(0, currentIndex - 1))}
          disabled={currentIndex === 0}
        >
          <ChevronLeft className="h-4 w-4 mr-1" />
          Previous
        </Button>
        <span className="text-sm text-muted-foreground">
          {currentIndex + 1} of {items.length}
        </span>
        <Button
          variant="outline"
          size="sm"
          onClick={() =>
            setCurrentIndex(Math.min(items.length - 1, currentIndex + 1))
          }
          disabled={currentIndex === items.length - 1}
        >
          Next
          <ChevronRight className="h-4 w-4 ml-1" />
        </Button>
      </div>

      {/* Content */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Input */}
        <Card>
          <CardHeader>
            <CardTitle>Input</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="text-sm bg-muted p-4 rounded-lg overflow-auto max-h-64 whitespace-pre-wrap">
              {typeof currentItem.input === "string"
                ? currentItem.input
                : JSON.stringify(currentItem.input, null, 2)}
            </pre>
          </CardContent>
        </Card>

        {/* Output */}
        <Card>
          <CardHeader>
            <CardTitle>Output</CardTitle>
            {currentItem.expectedOutput && (
              <CardDescription>
                Expected: {typeof currentItem.expectedOutput === "string"
                  ? currentItem.expectedOutput
                  : JSON.stringify(currentItem.expectedOutput)}
              </CardDescription>
            )}
          </CardHeader>
          <CardContent>
            <pre className="text-sm bg-muted p-4 rounded-lg overflow-auto max-h-64 whitespace-pre-wrap">
              {typeof currentItem.output === "string"
                ? currentItem.output
                : JSON.stringify(currentItem.output, null, 2)}
            </pre>
          </CardContent>
        </Card>
      </div>

      {/* Scoring */}
      <Card>
        <CardHeader>
          <CardTitle>Score: {queue.scoreName}</CardTitle>
          <CardDescription>
            Provide your evaluation for this item
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {queue.scoreDataType === "BOOLEAN" && (
            <div className="flex items-center gap-4">
              <Button
                size="lg"
                variant="outline"
                className="flex-1 h-16"
                onClick={() => handleSubmitBoolean(true)}
                disabled={submitMutation.isPending}
              >
                <ThumbsUp className="h-6 w-6 mr-2 text-green-500" />
                Pass
              </Button>
              <Button
                size="lg"
                variant="outline"
                className="flex-1 h-16"
                onClick={() => handleSubmitBoolean(false)}
                disabled={submitMutation.isPending}
              >
                <ThumbsDown className="h-6 w-6 mr-2 text-red-500" />
                Fail
              </Button>
            </div>
          )}

          {queue.scoreDataType === "NUMERIC" && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Score (0 - 1)</Label>
                <div className="flex items-center gap-4">
                  <input
                    type="range"
                    min="0"
                    max="1"
                    step="0.1"
                    value={numericScore}
                    onChange={(e) => setNumericScore(parseFloat(e.target.value))}
                    className="flex-1"
                  />
                  <Badge
                    variant={
                      numericScore >= 0.7
                        ? "default"
                        : numericScore >= 0.4
                        ? "secondary"
                        : "destructive"
                    }
                    className="w-16 justify-center"
                  >
                    {numericScore.toFixed(1)}
                  </Badge>
                </div>
              </div>
              <Button
                onClick={handleSubmitNumeric}
                disabled={submitMutation.isPending}
              >
                {submitMutation.isPending && (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Submit Score
              </Button>
            </div>
          )}

          {queue.scoreDataType === "CATEGORICAL" && queue.categories && (
            <div className="flex flex-wrap gap-2">
              {queue.categories.map((category) => (
                <Button
                  key={category}
                  variant="outline"
                  onClick={() => handleSubmitCategorical(category)}
                  disabled={submitMutation.isPending}
                >
                  {category}
                </Button>
              ))}
            </div>
          )}

          <Separator />

          {/* Comment */}
          <div className="space-y-2">
            <Label htmlFor="comment">Comment (optional)</Label>
            <Textarea
              id="comment"
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              placeholder="Add any notes about this evaluation..."
              rows={2}
            />
          </div>

          {/* Skip button */}
          <div className="flex justify-end">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => skipMutation.mutate(currentItem.id)}
              disabled={skipMutation.isPending}
            >
              <SkipForward className="h-4 w-4 mr-1" />
              Skip
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export function AnnotationInterfaceSkeleton() {
  return (
    <div className="space-y-6">
      <Card>
        <CardContent className="pt-6">
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Progress</span>
              <span>Loading...</span>
            </div>
            <Progress value={0} />
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Input</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="bg-muted p-4 rounded-lg h-48 animate-pulse" />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Output</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="bg-muted p-4 rounded-lg h-48 animate-pulse" />
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Score</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-24 animate-pulse bg-muted rounded-lg" />
        </CardContent>
      </Card>
    </div>
  );
}
