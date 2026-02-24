"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Layers,
  MessageSquare,
  Database,
  FlaskConical,
  BarChart3,
  Settings,
  ChevronLeft,
  ChevronRight,
  Workflow,
  GitBranch,
  Radio,
  GitCompareArrows,
  ShieldAlert,
  PiggyBank,
  UserCheck,
  Shield,
  Users,
  Search,
  GraduationCap,
  Lock,
  Webhook,
  Store,
  FileCheck,
  Network,
  SearchCheck,
  TrendingUp,
  Code,
  Wand2,
  Server,
  Eye,
  Puzzle,
  Brain,
  Share2,
  Flame,
  LineChart,
  ArrowRightLeft,
  Leaf,
  FlaskRound,
  Target,
  Gauge,
  Globe,
  CheckCircle,
  Receipt,
  ClipboardCheck,
  Image,
  Blocks,
  Link,
  Bot,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

const navigation = [
  {
    name: "Dashboard",
    href: "/dashboard",
    icon: LayoutDashboard,
  },
  {
    name: "Traces",
    href: "/traces",
    icon: Layers,
  },
  {
    name: "Sessions",
    href: "/sessions",
    icon: Workflow,
  },
  {
    name: "Checkpoints",
    href: "/checkpoints",
    icon: GitBranch,
  },
  {
    name: "Prompts",
    href: "/prompts",
    icon: MessageSquare,
  },
  {
    name: "Datasets",
    href: "/datasets",
    icon: Database,
  },
  {
    name: "Evaluators",
    href: "/evals",
    icon: FlaskConical,
  },
  {
    name: "Analytics",
    href: "/analytics",
    icon: BarChart3,
  },
  {
    name: "Streaming",
    href: "/streaming",
    icon: Radio,
  },
  {
    name: "Diff Intelligence",
    href: "/diff-intelligence",
    icon: GitCompareArrows,
  },
  {
    name: "Anomaly Detection",
    href: "/anomaly",
    icon: ShieldAlert,
  },
  {
    name: "Cost Optimizer",
    href: "/cost-optimizer",
    icon: PiggyBank,
  },
  {
    name: "Skill Profiles",
    href: "/skill-profiles",
    icon: UserCheck,
  },
  {
    name: "Prompt Lab",
    href: "/prompt-lab",
    icon: FlaskConical,
  },
  {
    name: "Sandbox",
    href: "/sandbox",
    icon: Shield,
  },
  {
    name: "Team",
    href: "/team",
    icon: Users,
  },
  {
    name: "Search",
    href: "/search",
    icon: Search,
  },
  {
    name: "Training",
    href: "/training",
    icon: GraduationCap,
  },
  {
    name: "Access Control",
    href: "/rbac",
    icon: Lock,
  },
  {
    name: "Webhooks",
    href: "/webhook-rules",
    icon: Webhook,
  },
  {
    name: "Marketplace",
    href: "/marketplace",
    icon: Store,
  },
  {
    name: "Compliance",
    href: "/compliance-reports",
    icon: FileCheck,
  },
  {
    name: "Orchestration",
    href: "/orchestration",
    icon: Network,
  },
  {
    name: "Root Cause",
    href: "/rca",
    icon: SearchCheck,
  },
  {
    name: "Agent Versions",
    href: "/agent-versions",
    icon: GitBranch,
  },
  {
    name: "Predictions",
    href: "/predictions",
    icon: TrendingUp,
  },
  {
    name: "Embed",
    href: "/embed",
    icon: Code,
  },
  {
    name: "Agent Builder",
    href: "/agent-builder",
    icon: Wand2,
  },
  {
    name: "Fleet",
    href: "/fleet",
    icon: Server,
  },
  {
    name: "Privacy",
    href: "/privacy",
    icon: Eye,
  },
  {
    name: "Plugins",
    href: "/plugins",
    icon: Puzzle,
  },
  {
    name: "Memory",
    href: "/memory",
    icon: Brain,
  },
  {
    name: "Distributed",
    href: "/distributed-traces",
    icon: Share2,
  },
  {
    name: "Cache",
    href: "/prompt-cache",
    icon: Database,
  },
  {
    name: "Chaos Testing",
    href: "/chaos",
    icon: Flame,
  },
  {
    name: "Metrics",
    href: "/custom-metrics",
    icon: LineChart,
  },
  {
    name: "Handoffs",
    href: "/handoffs",
    icon: ArrowRightLeft,
  },
  {
    name: "Carbon",
    href: "/carbon",
    icon: Leaf,
  },
  {
    name: "Synthetic Data",
    href: "/synthetic-data",
    icon: FlaskRound,
  },
  {
    name: "SLOs",
    href: "/slos",
    icon: Target,
  },
  {
    name: "Autonomy",
    href: "/autonomy",
    icon: Gauge,
  },
  {
    name: "Cross-Org",
    href: "/cross-org",
    icon: Globe,
  },
  {
    name: "Intents",
    href: "/intents",
    icon: CheckCircle,
  },
  {
    name: "Cost ROI",
    href: "/cost-attribution",
    icon: Receipt,
  },
  {
    name: "Knowledge Graph",
    href: "/knowledge-graph",
    icon: Share2,
  },
  {
    name: "Compliance",
    href: "/compliance-monitor",
    icon: ClipboardCheck,
  },
  {
    name: "Multi-Modal",
    href: "/multimodal",
    icon: Image,
  },
  {
    name: "Patterns",
    href: "/collab-patterns",
    icon: Blocks,
  },
  {
    name: "Federation",
    href: "/federated",
    icon: Link,
  },
  {
    name: "Copilot",
    href: "/copilot",
    icon: Bot,
  },
  {
    name: "Settings",
    href: "/settings",
    icon: Settings,
  },
];

export function Sidebar() {
  const pathname = usePathname();
  const [collapsed, setCollapsed] = React.useState(false);

  return (
    <TooltipProvider delayDuration={0}>
      <div
        className={cn(
          "flex h-full flex-col border-r bg-card transition-all duration-300",
          collapsed ? "w-16" : "w-64"
        )}
      >
        {/* Logo */}
        <div className="flex h-16 items-center border-b px-4">
          <Link href="/dashboard" className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-primary flex items-center justify-center flex-shrink-0">
              <span className="text-lg font-bold text-primary-foreground">A</span>
            </div>
            {!collapsed && (
              <span className="font-semibold text-lg">AgentTrace</span>
            )}
          </Link>
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-1 px-2 py-4">
          {navigation.map((item) => {
            const isActive =
              pathname === item.href || pathname.startsWith(`${item.href}/`);

            const linkContent = (
              <Link
                href={item.href}
                className={cn(
                  "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground",
                  collapsed && "justify-center px-2"
                )}
              >
                <item.icon className="h-5 w-5 flex-shrink-0" />
                {!collapsed && <span>{item.name}</span>}
              </Link>
            );

            if (collapsed) {
              return (
                <Tooltip key={item.name}>
                  <TooltipTrigger asChild>{linkContent}</TooltipTrigger>
                  <TooltipContent side="right">{item.name}</TooltipContent>
                </Tooltip>
              );
            }

            return <div key={item.name}>{linkContent}</div>;
          })}
        </nav>

        {/* Collapse button */}
        <div className="border-t p-2">
          <Button
            variant="ghost"
            size="sm"
            className={cn("w-full", collapsed && "px-2")}
            onClick={() => setCollapsed(!collapsed)}
          >
            {collapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <>
                <ChevronLeft className="h-4 w-4 mr-2" />
                <span>Collapse</span>
              </>
            )}
          </Button>
        </div>
      </div>
    </TooltipProvider>
  );
}
