'use client';

import * as React from 'react';
import {
  FileCode2,
  FilePlus2,
  FileMinus2,
  FileEdit,
  Search,
  FolderTree,
  BarChart3,
  Languages,
} from 'lucide-react';

import { cn } from '@/lib/utils';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useCodeImpact } from '@/hooks/use-code-impact';

interface CodeImpactMapProps {
  traceId?: string;
}

const operationConfig = {
  created: {
    color: 'bg-green-500/10 text-green-700 border-green-500/20',
    icon: FilePlus2,
    label: 'Created',
  },
  modified: {
    color: 'bg-yellow-500/10 text-yellow-700 border-yellow-500/20',
    icon: FileEdit,
    label: 'Modified',
  },
  deleted: {
    color: 'bg-red-500/10 text-red-700 border-red-500/20',
    icon: FileMinus2,
    label: 'Deleted',
  },
};

const complexityConfig = {
  low: 'bg-green-500/10 text-green-700',
  medium: 'bg-yellow-500/10 text-yellow-700',
  high: 'bg-red-500/10 text-red-700',
};

export function CodeImpactMap({ traceId: initialTraceId }: CodeImpactMapProps) {
  const [traceId, setTraceId] = React.useState(initialTraceId ?? '');
  const [submittedId, setSubmittedId] = React.useState(initialTraceId ?? '');

  const { data: impact, isLoading } = useCodeImpact(submittedId || null);
  const files = impact?.files ?? [];
  const summary = impact?.summary;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setSubmittedId(traceId);
  };

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit} className="flex items-center gap-2">
        <div className="relative max-w-md flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Enter trace ID..."
            value={traceId}
            onChange={(e) => setTraceId(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button type="submit" disabled={!traceId}>
          Analyze
        </Button>
      </form>

      {summary && (
        <div className="grid gap-4 md:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Files
              </CardTitle>
              <FolderTree className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{summary.totalFiles}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Lines Added
              </CardTitle>
              <BarChart3 className="h-4 w-4 text-green-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-600">+{summary.totalLinesAdded}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Lines Removed
              </CardTitle>
              <BarChart3 className="h-4 w-4 text-red-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-red-600">-{summary.totalLinesRemoved}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Languages</CardTitle>
              <Languages className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{summary.languages.length}</div>
              <p className="truncate text-xs text-muted-foreground">
                {summary.languages.join(', ')}
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {isLoading && (
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-16 animate-pulse rounded-lg bg-muted" />
          ))}
        </div>
      )}

      {files.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Impacted Files</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {files.map((file) => {
              const config = operationConfig[file.operation];
              const OpIcon = config.icon;
              return (
                <div
                  key={file.path}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div className="flex min-w-0 items-center gap-3">
                    <OpIcon className={cn('h-4 w-4 shrink-0', config.color.split(' ')[1])} />
                    <div className="min-w-0">
                      <p className="truncate font-mono text-sm">{file.path}</p>
                      <div className="mt-1 flex items-center gap-2">
                        <Badge variant="outline" className={cn('text-xs', config.color)}>
                          {config.label}
                        </Badge>
                        <Badge variant="outline" className="text-xs">
                          {file.language}
                        </Badge>
                        <Badge
                          variant="outline"
                          className={cn('text-xs', complexityConfig[file.complexity])}
                        >
                          {file.complexity} complexity
                        </Badge>
                      </div>
                    </div>
                  </div>
                  <div className="flex shrink-0 items-center gap-3 text-sm">
                    <span className="font-mono text-green-600">+{file.linesAdded}</span>
                    <span className="font-mono text-red-600">-{file.linesRemoved}</span>
                  </div>
                </div>
              );
            })}
          </CardContent>
        </Card>
      )}

      {submittedId && !isLoading && files.length === 0 && (
        <div className="py-12 text-center text-muted-foreground">
          <FileCode2 className="mx-auto mb-4 h-12 w-12 opacity-50" />
          <p className="text-lg font-medium">No code impact data found</p>
          <p className="mt-2 text-sm">Enter a valid trace ID to view the code impact map</p>
        </div>
      )}
    </div>
  );
}
