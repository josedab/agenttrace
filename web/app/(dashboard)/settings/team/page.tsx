"use client";

import * as React from "react";
import { useQuery } from "@tanstack/react-query";

import { api } from "@/lib/api";
import { PageHeader } from "@/components/layout/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { TeamMemberList } from "@/components/settings/team-member-list";
import { TeamInviteDialog } from "@/components/settings/team-invite-dialog";

export default function TeamSettingsPage() {
  const { data: members, isLoading, error } = useQuery({
    queryKey: ["team-members"],
    queryFn: () => api.team.listMembers(),
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Team"
        description="Manage team members and their permissions."
        actions={<TeamInviteDialog />}
      />

      {isLoading ? (
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="flex items-center gap-4 p-4 border rounded-lg">
              <Skeleton className="h-10 w-10 rounded-full" />
              <div className="space-y-2">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-3 w-48" />
              </div>
            </div>
          ))}
        </div>
      ) : error ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <p className="text-destructive">Failed to load team members</p>
        </div>
      ) : members && members.length > 0 ? (
        <TeamMemberList members={members} />
      ) : (
        <div className="flex flex-col items-center justify-center py-12 text-center border rounded-lg bg-muted/20">
          <h3 className="text-lg font-semibold">No team members</h3>
          <p className="text-sm text-muted-foreground mt-1 mb-4">
            Invite team members to collaborate on this project.
          </p>
          <TeamInviteDialog />
        </div>
      )}
    </div>
  );
}
