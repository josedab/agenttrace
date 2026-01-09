import { Suspense } from "react";
import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PromptEditor } from "@/components/prompts/prompt-editor";
import { PromptEditorSkeleton } from "@/components/prompts/prompt-editor-skeleton";

export const metadata = {
  title: "Edit Prompt | AgentTrace",
  description: "Edit and manage prompt versions",
};

interface PromptDetailPageProps {
  params: {
    promptName: string;
  };
}

export default function PromptDetailPage({ params }: PromptDetailPageProps) {
  const { promptName } = params;
  const decodedName = decodeURIComponent(promptName);

  if (!promptName) {
    notFound();
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/prompts">
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Prompts
          </Link>
        </Button>
      </div>
      <Suspense fallback={<PromptEditorSkeleton />}>
        <PromptEditor promptName={decodedName} />
      </Suspense>
    </div>
  );
}
