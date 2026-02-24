import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Marketplace | AgentTrace",
  description: "Browse and install agent configuration packages",
};

export default function MarketplacePage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Marketplace"
        description="Browse and install agent configuration packages"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <MarketplaceContent />
      </Suspense>
    </div>
  );
}

function MarketplaceContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Marketplace</p>
      <p className="text-sm mt-2">
        Browse, install, and publish agent configuration packages shared by the community.
      </p>
    </div>
  );
}
