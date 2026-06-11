'use client';

import type * as React from 'react';
import { Shield } from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import type { EvalHubVisibility } from '@/lib/eval-hub';

export function EvalHubErrorState({ label, onRetry }: { label: string; onRetry: () => void }) {
  return (
    <Card className="border-destructive/30">
      <CardContent className="flex min-h-44 flex-col items-center justify-center gap-3 text-center">
        <p className="text-sm text-destructive">{label}</p>
        <Button variant="outline" onClick={onRetry}>
          Retry
        </Button>
      </CardContent>
    </Card>
  );
}

export function EvalHubEmptyState({
  icon: Icon,
  title,
  description,
  action,
}: {
  icon: React.ComponentType<{ className?: string }>;
  title: string;
  description: string;
  action?: React.ReactNode;
}) {
  return (
    <Card>
      <CardContent className="flex min-h-56 flex-col items-center justify-center text-center">
        <Icon className="h-9 w-9 text-muted-foreground" />
        <p className="mt-4 font-medium">{title}</p>
        <p className="mt-1 max-w-md text-sm text-muted-foreground">{description}</p>
        {action ? <div className="mt-4">{action}</div> : null}
      </CardContent>
    </Card>
  );
}

export function EvalHubGridSkeleton() {
  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      {Array.from({ length: 6 }).map((_, index) => (
        <Skeleton key={index} className="h-48 w-full" />
      ))}
    </div>
  );
}

export function EvalHubVisibilityBadge({ visibility }: { visibility: EvalHubVisibility }) {
  return (
    <Badge variant="outline">
      {visibility !== 'public' ? <Shield className="mr-1 h-3 w-3" /> : null}
      {visibility}
    </Badge>
  );
}
