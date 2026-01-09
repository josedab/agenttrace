"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api";
import { PageHeader } from "@/components/layout/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiKeyList } from "@/components/settings/api-key-list";
import { CreateApiKeyDialog } from "@/components/settings/create-api-key-dialog";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function ApiKeysSettingsPage() {
  const { data: apiKeys, isLoading, error } = useQuery({
    queryKey: ["api-keys"],
    queryFn: () => api.apiKeys.list(),
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="API Keys"
        description="Manage API keys for SDK and programmatic access."
        actions={<CreateApiKeyDialog />}
      />

      {/* Usage instructions */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Start</CardTitle>
          <CardDescription>
            Use API keys to authenticate SDK requests.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <p className="text-sm font-medium mb-2">Python</p>
              <pre className="bg-muted p-3 rounded-md text-sm overflow-x-auto">
{`from agenttrace import AgentTrace

client = AgentTrace(
    api_key="your-api-key",
    host="https://api.agenttrace.io"
)`}
              </pre>
            </div>
            <div>
              <p className="text-sm font-medium mb-2">TypeScript</p>
              <pre className="bg-muted p-3 rounded-md text-sm overflow-x-auto">
{`import { AgentTrace } from '@agenttrace/sdk';

const client = new AgentTrace({
  apiKey: 'your-api-key',
  host: 'https://api.agenttrace.io'
});`}
              </pre>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* API Keys list */}
      {isLoading ? (
        <Card>
          <CardHeader>
            <CardTitle>Your API Keys</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {[...Array(3)].map((_, i) => (
                <div key={i} className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="space-y-2">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-3 w-48" />
                  </div>
                  <Skeleton className="h-8 w-20" />
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      ) : error ? (
        <Card>
          <CardContent className="py-12">
            <p className="text-destructive text-center">
              Failed to load API keys
            </p>
          </CardContent>
        </Card>
      ) : apiKeys && apiKeys.length > 0 ? (
        <ApiKeyList apiKeys={apiKeys} />
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <h3 className="text-lg font-semibold">No API keys</h3>
            <p className="text-sm text-muted-foreground mt-1 mb-4">
              Create an API key to start using the SDK.
            </p>
            <CreateApiKeyDialog />
          </CardContent>
        </Card>
      )}
    </div>
  );
}
