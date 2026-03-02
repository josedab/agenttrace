import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Marketplace | AgentTrace",
  description: "Community marketplace for evaluator templates, agent blueprints, and plugins",
};

export default function MarketplacePage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Marketplace"
        description="Community-driven marketplace for evaluator templates, agent blueprints, dashboard widgets, and integration plugins"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <MarketplaceContent />
      </Suspense>
    </div>
  );
}

function MarketplaceContent() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="Packages" value="5" description="Available packages" />
        <StatCard title="Downloads" value="10,053" description="Total installs" />
        <StatCard title="Publishers" value="4" description="Community publishers" />
        <StatCard title="Avg Rating" value="4.7" description="Across all packages" />
      </div>

      {/* Starter Kits */}
      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Starter Kits</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Curated collections for common AI patterns — get started in minutes.
        </p>
        <div className="grid gap-4 md:grid-cols-4">
          {[
            { name: "RAG Agent", installs: 342, pattern: "rag" },
            { name: "Coding Agent", installs: 567, pattern: "coding_agent" },
            { name: "Chatbot", installs: 891, pattern: "chatbot" },
            { name: "Data Pipeline", installs: 234, pattern: "data_pipeline" },
          ].map((kit) => (
            <div key={kit.name} className="rounded-md border p-4">
              <p className="font-medium">{kit.name}</p>
              <p className="text-xs text-muted-foreground">{kit.installs} installs</p>
            </div>
          ))}
        </div>
      </div>

      {/* Categories */}
      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Categories</h3>
        <div className="grid gap-3 md:grid-cols-5">
          {[
            { name: "Prompts", desc: "Prompt templates" },
            { name: "Guardrails", desc: "Safety rules" },
            { name: "Evaluators", desc: "Quality checks" },
            { name: "Benchmarks", desc: "Test suites" },
            { name: "Bundles", desc: "All-in-one" },
          ].map((cat) => (
            <div key={cat.name} className="rounded-md border p-3 text-center hover:border-primary cursor-pointer transition-colors">
              <p className="font-medium text-sm">{cat.name}</p>
              <p className="text-xs text-muted-foreground">{cat.desc}</p>
            </div>
          ))}
        </div>
      </div>

      {/* Featured Packages */}
      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Featured Packages</h3>
        <div className="grid gap-4 md:grid-cols-2">
          {[
            { name: "RAG Quality Evaluator", author: "eval-labs", downloads: 3201, rating: 4.9, type: "evaluator" },
            { name: "Agent Benchmark Suite", author: "bench-team", downloads: 2456, rating: 4.8, type: "benchmark" },
            { name: "Safety Guardrail Suite", author: "agenttrace-team", downloads: 1842, rating: 4.7, type: "guardrail" },
            { name: "Production Agent Bundle", author: "agenttrace-team", downloads: 1567, rating: 4.6, type: "bundle" },
          ].map((pkg) => (
            <div key={pkg.name} className="rounded-md border p-4 flex justify-between items-start">
              <div>
                <p className="font-medium">{pkg.name}</p>
                <p className="text-xs text-muted-foreground">by {pkg.author}</p>
                <div className="flex gap-3 mt-2 text-xs text-muted-foreground">
                  <span>{pkg.downloads.toLocaleString()} downloads</span>
                  <span>★ {pkg.rating}</span>
                </div>
              </div>
              <span className="px-2 py-1 rounded-full bg-primary/10 text-primary text-xs">{pkg.type}</span>
            </div>
          ))}
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
