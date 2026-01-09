import { Suspense } from "react";
import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PromptPlayground } from "@/components/prompts/prompt-playground";
import { Skeleton } from "@/components/ui/skeleton";

export const metadata = {
  title: "Prompt Playground | AgentTrace",
  description: "Test and iterate on your prompts",
};

interface PromptPlaygroundPageProps {
  params: {
    promptName: string;
  };
}

export default function PromptPlaygroundPage({ params }: PromptPlaygroundPageProps) {
  const { promptName } = params;
  const decodedName = decodeURIComponent(promptName);

  if (!promptName) {
    notFound();
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" asChild>
          <Link href={`/prompts/${encodeURIComponent(decodedName)}`}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Prompt
          </Link>
        </Button>
      </div>
      <Suspense fallback={<PlaygroundSkeleton />}>
        <PromptPlayground promptName={decodedName} />
      </Suspense>
    </div>
  );
}

function PlaygroundSkeleton() {
  return (
    <div className="grid lg:grid-cols-2 gap-6">
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64 w-full" />
      </div>
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64 w-full" />
      </div>
    </div>
  );
}
