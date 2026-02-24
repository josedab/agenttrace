import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Compliance Monitor | AgentTrace",
  description: "Continuous compliance monitoring with real-time scoring",
};

export default function ComplianceMonitorPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Compliance Monitor"
        description="Continuous compliance monitoring with real-time scoring"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <ComplianceMonitorContent />
      </Suspense>
    </div>
  );
}

function ComplianceMonitorContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Compliance Monitor Dashboard</p>
      <p className="text-sm mt-2">Define policies and rules, evaluate traces for compliance violations</p>
      <p className="text-sm mt-1">Real-time scoring across SOC2, HIPAA, GDPR, and custom frameworks</p>
    </div>
  );
}
