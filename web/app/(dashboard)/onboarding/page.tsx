import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Cloud Onboarding | AgentTrace",
  description: "Guided onboarding, quickstart generation, and usage metering",
};

export default function OnboardingPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Cloud Onboarding"
        description="Get started quickly with guided setup, quickstart configs, and usage monitoring"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <OnboardingDashboardContent />
      </Suspense>
    </div>
  );
}

function OnboardingDashboardContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Cloud Onboarding Dashboard</p>
      <p className="text-sm mt-2">Follow step-by-step onboarding and generate framework-specific quickstart configs</p>
      <p className="text-sm mt-1">Includes usage metering, quota management, and billing period tracking</p>
    </div>
  );
}
