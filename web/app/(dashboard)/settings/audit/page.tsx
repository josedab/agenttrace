"use client";

import * as React from "react";
import { FileText } from "lucide-react";

import { PageHeader } from "@/components/layout/page-header";
import { AuditLogList } from "@/components/audit/audit-log-list";
import { AuditLogFilters } from "@/components/audit/audit-log-filters";
import { AuditLogSummary } from "@/components/audit/audit-log-summary";
import { AuditExportPanel } from "@/components/audit/audit-export-panel";

export default function AuditLogsPage() {
  // In a real app, this would come from auth context
  const organizationId = "org-1";

  const [filters, setFilters] = React.useState({
    userId: undefined as string | undefined,
    action: undefined as string | undefined,
    resourceType: undefined as string | undefined,
    startDate: undefined as string | undefined,
    endDate: undefined as string | undefined,
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Audit Logs"
        description="Track and review all security and administrative events in your organization."
        icon={FileText}
      />

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <AuditLogFilters filters={filters} onFiltersChange={setFilters} />
          <AuditLogList organizationId={organizationId} filters={filters} />
        </div>
        <div className="space-y-6">
          <AuditLogSummary organizationId={organizationId} />
          <AuditExportPanel organizationId={organizationId} />
        </div>
      </div>
    </div>
  );
}
