"use client";

import * as React from "react";
import { Shield } from "lucide-react";

import { PageHeader } from "@/components/layout/page-header";
import { SSOConfigurationForm } from "@/components/sso/sso-configuration-form";
import { SSOConfigurationList } from "@/components/sso/sso-configuration-list";

export default function SSOSettingsPage() {
  // In a real app, this would come from auth context
  const organizationId = "org-1";

  return (
    <div className="space-y-6">
      <PageHeader
        title="Single Sign-On (SSO)"
        description="Configure SAML or OIDC authentication for your organization."
        icon={Shield}
      />

      <div className="grid gap-6">
        <SSOConfigurationList organizationId={organizationId} />
        <SSOConfigurationForm organizationId={organizationId} />
      </div>
    </div>
  );
}
