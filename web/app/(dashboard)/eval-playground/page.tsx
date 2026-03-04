import { Metadata } from "next";
import { EvalPlayground } from "@/components/eval-playground/eval-playground";

export const metadata: Metadata = {
  title: "Evaluation Playground | AgentTrace",
  description: "Write and test evaluator functions against live trace data",
};

export default function EvalPlaygroundPage() {
  return (
    <div className="container mx-auto py-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Evaluation Playground</h1>
        <p className="text-muted-foreground mt-1">
          Write evaluator functions, test them against traces, and share with your team
        </p>
      </div>
      <EvalPlayground />
    </div>
  );
}
