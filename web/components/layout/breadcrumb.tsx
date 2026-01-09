"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronRight, Home } from "lucide-react";

import { cn } from "@/lib/utils";

const routeLabels: Record<string, string> = {
  dashboard: "Dashboard",
  traces: "Traces",
  sessions: "Sessions",
  prompts: "Prompts",
  datasets: "Datasets",
  evals: "Evaluators",
  analytics: "Analytics",
  settings: "Settings",
  profile: "Profile",
  project: "Project",
  team: "Team",
  "api-keys": "API Keys",
  cost: "Cost",
  latency: "Latency",
  usage: "Usage",
  playground: "Playground",
  versions: "Versions",
  runs: "Runs",
  queues: "Queues",
};

export function Breadcrumb() {
  const pathname = usePathname();
  const segments = pathname.split("/").filter(Boolean);

  // Build breadcrumb items
  const items = segments.map((segment, index) => {
    const href = "/" + segments.slice(0, index + 1).join("/");
    const label = routeLabels[segment] || formatSegment(segment);
    const isLast = index === segments.length - 1;

    return {
      href,
      label,
      isLast,
    };
  });

  return (
    <nav className="flex items-center space-x-1 text-sm">
      <Link
        href="/dashboard"
        className="text-muted-foreground hover:text-foreground transition-colors"
      >
        <Home className="h-4 w-4" />
      </Link>

      {items.map((item, index) => (
        <React.Fragment key={item.href}>
          <ChevronRight className="h-4 w-4 text-muted-foreground" />
          {item.isLast ? (
            <span className="font-medium text-foreground">{item.label}</span>
          ) : (
            <Link
              href={item.href}
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              {item.label}
            </Link>
          )}
        </React.Fragment>
      ))}
    </nav>
  );
}

function formatSegment(segment: string): string {
  // Check if it looks like an ID (uuid or similar)
  if (segment.match(/^[a-f0-9-]{8,}$/i)) {
    return segment.slice(0, 8) + "...";
  }

  // Convert kebab-case or snake_case to Title Case
  return segment
    .replace(/[-_]/g, " ")
    .replace(/\b\w/g, (char) => char.toUpperCase());
}
