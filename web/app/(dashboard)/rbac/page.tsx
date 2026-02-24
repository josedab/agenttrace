import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Access Control | AgentTrace",
  description: "Manage roles, permissions, and SSO configuration",
};

export default function RBACPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Access Control"
        description="Manage roles, permissions, and SSO configuration"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <RBACContent />
      </Suspense>
    </div>
  );
}

function RBACContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Access Control</p>
      <p className="text-sm mt-2">
        Configure role-based access control, manage user permissions, and set up SSO integration.
      </p>
    </div>
  );
}
