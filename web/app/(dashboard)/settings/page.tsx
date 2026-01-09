"use client";

import * as React from "react";
import Link from "next/link";
import { User, Building2, Users, Key, Bell, Palette, Shield, FileText } from "lucide-react";

import { PageHeader } from "@/components/layout/page-header";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

const settingsSections = [
  {
    title: "Profile",
    description: "Manage your personal account settings",
    href: "/settings/profile",
    icon: User,
  },
  {
    title: "Project",
    description: "Configure project settings and preferences",
    href: "/settings/project",
    icon: Building2,
  },
  {
    title: "Team",
    description: "Manage team members and roles",
    href: "/settings/team",
    icon: Users,
  },
  {
    title: "API Keys",
    description: "Manage API keys for SDK access",
    href: "/settings/api-keys",
    icon: Key,
  },
  {
    title: "Single Sign-On",
    description: "Configure SAML or OIDC authentication",
    href: "/settings/sso",
    icon: Shield,
  },
  {
    title: "Audit Logs",
    description: "Track security and administrative events",
    href: "/settings/audit",
    icon: FileText,
  },
  {
    title: "Notifications",
    description: "Configure notification preferences",
    href: "/settings/notifications",
    icon: Bell,
  },
  {
    title: "Appearance",
    description: "Customize the look and feel",
    href: "/settings/appearance",
    icon: Palette,
  },
];

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Settings"
        description="Manage your account and project settings."
      />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {settingsSections.map((section) => {
          const Icon = section.icon;
          return (
            <Link key={section.href} href={section.href}>
              <Card className="hover:border-primary/50 transition-colors cursor-pointer h-full">
                <CardHeader>
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-muted rounded-md">
                      <Icon className="h-5 w-5" />
                    </div>
                    <CardTitle className="text-base">{section.title}</CardTitle>
                  </div>
                </CardHeader>
                <CardContent>
                  <CardDescription>{section.description}</CardDescription>
                </CardContent>
              </Card>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
