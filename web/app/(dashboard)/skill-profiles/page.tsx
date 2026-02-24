import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Agent Skill Profiles | AgentTrace",
  description: "Analyze agent capabilities across code generation, refactoring, testing, and more",
};

export default function SkillProfilesPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Agent Skill Profiles"
        description="Analyze agent capabilities across code generation, refactoring, testing, and more"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <SkillProfilesContent />
      </Suspense>
    </div>
  );
}

function SkillProfilesContent() {
  return (
    <div className="text-center py-12 text-muted-foreground">
      <p className="text-lg font-medium">Agent Skill Profiles</p>
      <p className="text-sm mt-2">
        View and compare agent capabilities across code generation, refactoring, testing, debugging, and documentation dimensions.
      </p>
    </div>
  );
}
