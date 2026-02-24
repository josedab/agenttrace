import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Annotations | AgentTrace",
  description: "Collaboratively annotate traces with your team in real-time",
};

export default function AnnotationsPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Annotations" description="Collaboratively annotate traces with your team in real-time" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <AnnotationsContent />
      </Suspense>
    </div>
  );
}

function AnnotationsContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Annotations</p>
      <p className="text-sm mt-2">Collaboratively annotate traces with your team in real-time</p>
    </div>
  );
}
