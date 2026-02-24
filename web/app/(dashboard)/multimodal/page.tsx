import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Multi-Modal Traces | AgentTrace",
  description: "View image, audio, and video attachments in traces",
};

export default function MultimodalPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Multi-Modal Traces"
        description="View image, audio, and video attachments in traces"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <MultimodalContent />
      </Suspense>
    </div>
  );
}

function MultimodalContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Multi-Modal Traces Dashboard</p>
      <p className="text-sm mt-2">Browse and preview image, audio, and video attachments linked to traces</p>
      <p className="text-sm mt-1">Supports inline rendering, transcription previews, and attachment summaries</p>
    </div>
  );
}
