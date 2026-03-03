import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Data Warehouse Sync | AgentTrace",
  description: "Sync traces with Snowflake, BigQuery, Databricks, and S3",
};

export default function WarehouseSyncPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Data Warehouse Sync"
        description="Native bidirectional sync with Snowflake, BigQuery, Databricks, and S3/Parquet"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <WarehouseContent />
      </Suspense>
    </div>
  );
}

function WarehouseContent() {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="Connections" value="0" description="Active connections" />
        <StatCard title="Records Synced" value="0" description="Total synced" />
        <StatCard title="Last Sync" value="—" description="Most recent" />
        <StatCard title="Data Volume" value="0 MB" description="Total transferred" />
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-4">Supported Warehouses</h3>
        <div className="grid gap-4 md:grid-cols-4">
          {[
            { name: "Snowflake", desc: "Cloud data warehouse" },
            { name: "BigQuery", desc: "Google Cloud analytics" },
            { name: "Databricks", desc: "Lakehouse platform" },
            { name: "S3/Parquet", desc: "Object storage export" },
          ].map((wh) => (
            <div key={wh.name} className="rounded-md border p-4 text-center">
              <p className="font-medium">{wh.name}</p>
              <p className="text-xs text-muted-foreground">{wh.desc}</p>
            </div>
          ))}
        </div>
      </div>

      <div className="rounded-lg border bg-card p-6">
        <h3 className="text-lg font-semibold mb-2">Connections</h3>
        <div className="text-center text-sm text-muted-foreground py-8">
          No warehouse connections configured. Add one to start syncing trace data.
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
