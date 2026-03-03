import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Edge Devices | AgentTrace",
  description: "Monitor AI agents on mobile, IoT, and edge environments",
};

export default function EdgePage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Edge & Mobile Monitoring"
        description="Lightweight SDK for monitoring AI agents on mobile devices, IoT, and edge environments"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <EdgeContent />
      </Suspense>
    </div>
  );
}

function EdgeContent() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="Devices" value="0" description="Registered devices" />
        <StatCard title="Online" value="0" description="Currently online" />
        <StatCard title="Events" value="0" description="Total ingested" />
        <StatCard title="Bandwidth Saved" value="0 KB" description="Via batching" />
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Supported Platforms</h3>
        <div className="grid gap-4 md:grid-cols-5">
          {[
            { name: "iOS", desc: "Swift SDK" },
            { name: "Android", desc: "Kotlin SDK" },
            { name: "WASM", desc: "WebAssembly" },
            { name: "IoT", desc: "Lightweight C" },
            { name: "Desktop", desc: "Cross-platform" },
          ].map((p) => (
            <div key={p.name} className="rounded-md border p-3 text-center">
              <p className="font-medium">{p.name}</p>
              <p className="text-xs text-muted-foreground">{p.desc}</p>
            </div>
          ))}
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-2">SDK Features</h3>
          <div className="space-y-2 text-sm">
            <div>• Offline trace buffering with automatic sync</div>
            <div>• Bandwidth-optimized batch uploads</div>
            <div>• Privacy-preserving local aggregation</div>
            <div>• Compressed event payloads (&lt;50KB SDK)</div>
            <div>• Automatic device fingerprinting</div>
          </div>
        </div>
        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-2">Registered Devices</h3>
          <div className="text-center text-sm text-muted-foreground py-6">
            No edge devices registered. Install the SDK to get started.
          </div>
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
