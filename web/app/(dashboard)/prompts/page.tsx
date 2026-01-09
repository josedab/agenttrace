import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { PromptList } from "@/components/prompts/prompt-list";
import { PromptListSkeleton } from "@/components/prompts/prompt-list-skeleton";
import { CreatePromptDialog } from "@/components/prompts/create-prompt-dialog";

export const metadata = {
  title: "Prompts | AgentTrace",
  description: "Manage and version your AI prompts",
};

export default function PromptsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Prompts"
        description="Manage and version your AI prompts"
      >
        <CreatePromptDialog />
      </PageHeader>
      <Suspense fallback={<PromptListSkeleton />}>
        <PromptList />
      </Suspense>
    </div>
  );
}
