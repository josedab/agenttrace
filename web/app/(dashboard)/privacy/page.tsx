import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Privacy & PII | AgentTrace",
  description: "PII detection, data redaction, and data residency controls",
};

export default function PrivacyPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Privacy & PII" description="PII detection, data redaction, and data residency controls" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <PrivacyContent />
      </Suspense>
    </div>
  );
}

function PrivacyContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Privacy & PII</p>
      <p className="text-sm mt-2">PII detection, automatic data redaction, and data residency controls for compliance.</p>
    </div>
  );
}
