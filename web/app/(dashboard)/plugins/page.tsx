import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Plugins | AgentTrace",
  description: "Install and manage plugins for custom evaluators, processors, and widgets",
};

export default function PluginsPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Plugins" description="Install and manage plugins for custom evaluators, processors, and widgets" />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <PluginsContent />
      </Suspense>
    </div>
  );
}

function PluginsContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Plugins</p>
      <p className="text-sm mt-2">Install and manage plugins for custom evaluators, data processors, and dashboard widgets.</p>
    </div>
  );
}
