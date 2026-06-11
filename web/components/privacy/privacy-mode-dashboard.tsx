'use client';

import { useQuery } from '@tanstack/react-query';
import {
  CheckCircle2,
  CloudOff,
  ExternalLink,
  LockKeyhole,
  RefreshCw,
  ShieldCheck,
  XCircle,
} from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { privacyModeApi } from '@/lib/privacy-mode';

const capabilityLabels: Record<string, string> = {
  webhookDelivery: 'Slack, Discord, and webhooks',
  githubReporting: 'GitHub workflow reporting',
  remoteImport: 'Remote migration sources',
  remoteExport: 'Remote export destinations',
  externalModelProviders: 'External model providers',
  otelExport: 'OpenTelemetry export and destinations',
  federation: 'Federated peers and queries',
  warehouseSync: 'Data warehouse connections and syncs',
  sentry: 'Sentry telemetry',
  localTraceStorage: 'Local trace storage',
  redactedShareLinks: 'Redacted share links',
};

export function PrivacyModeDashboard() {
  const query = useQuery({
    queryKey: ['privacy-capabilities'],
    queryFn: privacyModeApi.getCapabilities,
  });

  if (query.isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-40 w-full" />
        <Skeleton className="h-80 w-full" />
      </div>
    );
  }
  if (query.isError || !query.data) {
    return (
      <Card className="border-destructive/30">
        <CardContent className="flex min-h-52 flex-col items-center justify-center gap-3 text-center">
          <XCircle className="h-8 w-8 text-destructive" />
          <p className="font-medium">Privacy capabilities could not be loaded.</p>
          <Button variant="outline" onClick={() => query.refetch()}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  const capabilities = query.data;
  return (
    <div className="space-y-6">
      <Card className={capabilities.noEgress ? 'border-emerald-500/40' : undefined}>
        <CardContent className="grid gap-6 p-6 md:grid-cols-[auto_minmax(0,1fr)_auto] md:items-center">
          <div className="grid h-12 w-12 place-items-center rounded-xl border bg-muted/40">
            {capabilities.noEgress ? (
              <CloudOff className="h-6 w-6 text-emerald-600" />
            ) : (
              <ExternalLink className="h-6 w-6 text-muted-foreground" />
            )}
          </div>
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-xl font-semibold">
                {capabilities.noEgress ? 'Local/private mode' : 'Standard connected mode'}
              </h2>
              <Badge variant="outline">{capabilities.mode.replace('_', ' ')}</Badge>
            </div>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
              {capabilities.noEgress
                ? 'Outbound integrations are blocked at runtime. Startup validation also rejects conflicting external providers.'
                : 'Outbound capabilities are available when individually configured. AgentTrace does not claim no-egress in this mode.'}
            </p>
          </div>
          <div className="flex items-center gap-2 text-sm">
            <ShieldCheck className="h-4 w-4 text-emerald-600" />
            Deterministic redaction {capabilities.redactionEnabled ? 'on' : 'off'}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Effective capabilities</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-2">
          {Object.entries(capabilities.capabilities).map(([key, capability]) => (
            <div key={key} className="flex items-start gap-3 rounded-lg border p-4">
              {capability.available ? (
                <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" />
              ) : (
                <LockKeyhole className="mt-0.5 h-4 w-4 shrink-0 text-amber-600" />
              )}
              <div>
                <p className="text-sm font-medium">{capabilityLabels[key] || key}</p>
                <p className="mt-1 text-xs text-muted-foreground">
                  {capability.available ? 'Available' : capability.reason || 'Unavailable'}
                </p>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <div className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
        Configure with <code>PRIVACY_NO_EGRESS=true</code> and{' '}
        <code>PRIVACY_REDACTION_ENABLED=true</code>. AgentTrace refuses startup when enabled
        external providers conflict with no-egress mode.
      </div>
    </div>
  );
}
