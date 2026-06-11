'use client';

import * as React from 'react';
import { FileText } from 'lucide-react';

import { PageHeader } from '@/components/layout/page-header';
import { AuditLogList } from '@/components/audit/audit-log-list';
import { AuditLogFilters } from '@/components/audit/audit-log-filters';
import { AuditLogSummary } from '@/components/audit/audit-log-summary';
import { AuditExportPanel } from '@/components/audit/audit-export-panel';

export default function AuditLogsPage() {
  // In a real app, this would come from auth context
  const organizationId = 'org-1';

  const [filters, setFilters] = React.useState<{
    userId?: string;
    action?: string;
    resourceType?: string;
    startDate?: string;
    endDate?: string;
  }>({
    userId: undefined,
    action: undefined,
    resourceType: undefined,
    startDate: undefined,
    endDate: undefined,
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Audit Logs"
        description="Track and review all security and administrative events in your organization."
        icon={FileText}
      />

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
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
