import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Test Suites | AgentTrace",
  description: "Trace-powered test generation and regression testing",
};

export default function TestSuitesPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Trace-Powered Test Generation"
        description="Record production traces and automatically generate reproducible test suites with assertion-based regression tests"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <TestSuitesContent />
      </Suspense>
    </div>
  );
}

function TestSuitesContent() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="Test Suites" value="0" description="Total suites" />
        <StatCard title="Test Cases" value="0" description="Generated cases" />
        <StatCard title="Pass Rate" value="—" description="Latest run" />
        <StatCard title="Snapshots" value="0" description="Golden snapshots" />
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Generate Tests from Traces</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Select production traces to automatically generate test cases with input/output pairs and assertions.
        </p>
        <div className="grid gap-4 md:grid-cols-2">
          <div className="rounded-md border p-4">
            <h4 className="font-medium mb-2">Supported Frameworks</h4>
            <div className="flex gap-2 flex-wrap">
              {["pytest", "jest", "custom"].map((fw) => (
                <span key={fw} className="px-2 py-1 rounded-full bg-primary/10 text-primary text-xs font-medium">{fw}</span>
              ))}
            </div>
          </div>
          <div className="rounded-md border p-4">
            <h4 className="font-medium mb-2">Assertion Types</h4>
            <div className="flex gap-2 flex-wrap">
              {["exact_match", "contains", "json_path", "regex", "similarity"].map((t) => (
                <span key={t} className="px-2 py-1 rounded-full bg-secondary text-secondary-foreground text-xs font-medium">{t}</span>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-2">Test Suites</h3>
        <div className="text-center text-sm text-muted-foreground py-8">
          No test suites yet. Generate tests from your production traces to get started.
        </div>
      </div>
    </div>
  );
}

function StatCard({ title, value, description }: { title: string; value: string; description: string }) {
  return (
    <div className="rounded-lg border bg-card p-4">
      <p className="text-sm text-muted-foreground">{title}</p>
      <p className="text-2xl font-bold">{value}</p>
      <p className="text-xs text-muted-foreground">{description}</p>
    </div>
  );
}
