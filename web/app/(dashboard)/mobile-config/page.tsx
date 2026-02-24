import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Mobile App | AgentTrace",
  description: "Configure push notifications and mobile companion app settings",
};

export default function MobileConfigPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Mobile App" description="Configure push notifications and mobile companion app settings" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <MobileConfigContent />
      </Suspense>
    </div>
  );
}

function MobileConfigContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Mobile App</p>
      <p className="text-sm mt-2">Configure push notifications, device registration, and mobile companion app settings.</p>
    </div>
  );
}
