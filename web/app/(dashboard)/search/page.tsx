import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Semantic Search | AgentTrace",
  description: "Search traces using natural language",
};

export default function SearchPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Semantic Search"
        description="Search traces using natural language"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <SearchContent />
      </Suspense>
    </div>
  );
}

function SearchContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Semantic Search</p>
      <p className="text-sm mt-2">
        Search through traces using natural language queries with AI-powered semantic understanding.
      </p>
    </div>
  );
}
