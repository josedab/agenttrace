import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Embedding & White-Label | AgentTrace",
  description: "Configure embeddable dashboards and white-label settings",
};

export default function EmbedPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Embedding & White-Label" description="Configure embeddable dashboards and white-label settings" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <EmbedContent />
      </Suspense>
    </div>
  );
}

function EmbedContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Embedding & White-Label</p>
      <p className="text-sm mt-2">Configure embeddable dashboards, custom branding, and white-label settings for your clients.</p>
    </div>
  );
}
