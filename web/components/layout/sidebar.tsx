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
  ChevronDown,
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
  Link as LinkIcon,
  Bot,
  PlayCircle,
  ShieldCheck,
  GitFork,
  TestTube2,
  Trophy,
  Sparkles,
  GitGraph,
  FileCode,
  ScanSearch,
  Rocket,
  Bug,
  Zap,
  BellRing,
  ClipboardList,
  MessagesSquare,
  Satellite,
  Scale,
  Cpu,
  BarChartHorizontal,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface NavItem {
  name: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface NavSection {
  title: string;
  items: NavItem[];
}

const sections: NavSection[] = [
  {
    title: "Core",
    items: [
      { name: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
      { name: "Traces", href: "/traces", icon: Layers },
      { name: "Sessions", href: "/sessions", icon: Workflow },
      { name: "Checkpoints", href: "/checkpoints", icon: GitBranch },
      { name: "Prompts", href: "/prompts", icon: MessageSquare },
      { name: "Datasets", href: "/datasets", icon: Database },
      { name: "Evaluators", href: "/evals", icon: FlaskConical },
      { name: "Analytics", href: "/analytics", icon: BarChart3 },
      { name: "Streaming", href: "/streaming", icon: Radio },
      { name: "Search", href: "/search", icon: Search },
      { name: "Semantic Search", href: "/semantic-search", icon: Sparkles },
    ],
  },
  {
    title: "Analysis",
    items: [
      { name: "Diff Intelligence", href: "/diff-intelligence", icon: GitCompareArrows },
      { name: "Anomaly Detection", href: "/anomaly", icon: ShieldAlert },
      { name: "Root Cause", href: "/rca", icon: SearchCheck },
      { name: "Predictions", href: "/predictions", icon: TrendingUp },
      { name: "AI Debugger", href: "/ai-debugger", icon: Bug },
      { name: "Regression", href: "/regression", icon: ClipboardList },
      { name: "Regression AI", href: "/regression-detection", icon: Scale },
      { name: "Code Quality", href: "/code-quality", icon: FileCode },
      { name: "Agent Compare", href: "/agent-comparison", icon: BarChartHorizontal },
      { name: "Benchmarks", href: "/agent-benchmarks", icon: Trophy },
      { name: "Metrics", href: "/custom-metrics", icon: LineChart },
      { name: "SLOs", href: "/slos", icon: Target },
      { name: "Scorecards", href: "/skill-profiles", icon: UserCheck },
    ],
  },
  {
    title: "Cost",
    items: [
      { name: "Cost Optimizer", href: "/cost-optimizer", icon: PiggyBank },
      { name: "Cost ROI", href: "/cost-attribution", icon: Receipt },
      { name: "Cost Guardrails", href: "/cost-guardrails", icon: ShieldCheck },
      { name: "Cost Alerts", href: "/cost-alerts", icon: BellRing },
      { name: "Carbon", href: "/carbon", icon: Leaf },
    ],
  },
  {
    title: "Compliance",
    items: [
      { name: "Compliance Reports", href: "/compliance-reports", icon: FileCheck },
      { name: "Compliance Monitor", href: "/compliance-monitor", icon: ClipboardCheck },
      { name: "Privacy", href: "/privacy", icon: Eye },
      { name: "Access Control", href: "/rbac", icon: Lock },
      { name: "Security", href: "/security", icon: Shield },
    ],
  },
  {
    title: "Agents",
    items: [
      { name: "Agent Builder", href: "/agent-builder", icon: Wand2 },
      { name: "Agent Versions", href: "/agent-versions", icon: GitBranch },
      { name: "Multi-Agent", href: "/multi-agent", icon: GitFork },
      { name: "Autonomy", href: "/autonomy", icon: Gauge },
      { name: "Handoffs", href: "/handoffs", icon: ArrowRightLeft },
      { name: "Memory", href: "/memory", icon: Brain },
      { name: "Fleet", href: "/fleet", icon: Server },
      { name: "Copilot", href: "/copilot", icon: Bot },
      { name: "Knowledge Graph", href: "/knowledge-graph", icon: GitGraph },
    ],
  },
  {
    title: "Prompts & Data",
    items: [
      { name: "Prompt Lab", href: "/prompt-lab", icon: FlaskConical },
      { name: "Prompt CI", href: "/prompt-ci", icon: TestTube2 },
      { name: "Prompt Opt", href: "/prompt-optimization", icon: Zap },
      { name: "Cache", href: "/prompt-cache", icon: Database },
      { name: "Synthetic Data", href: "/synthetic-data", icon: FlaskRound },
      { name: "Training", href: "/training", icon: GraduationCap },
    ],
  },
  {
    title: "Collaboration",
    items: [
      { name: "Team", href: "/team", icon: Users },
      { name: "Collaboration", href: "/collab", icon: MessagesSquare },
      { name: "Patterns", href: "/collab-patterns", icon: Blocks },
      { name: "Cross-Org", href: "/cross-org", icon: Globe },
    ],
  },
  {
    title: "Infrastructure",
    items: [
      { name: "Distributed", href: "/distributed-traces", icon: Share2 },
      { name: "Orchestration", href: "/orchestration", icon: Network },
      { name: "Workflows", href: "/workflows", icon: Workflow },
      { name: "OpenTelemetry", href: "/otel", icon: Satellite },
      { name: "Webhooks", href: "/webhook-rules", icon: Webhook },
      { name: "Plugins", href: "/plugins", icon: Puzzle },
      { name: "Federation", href: "/federated", icon: LinkIcon },
      { name: "Sandbox", href: "/sandbox", icon: Cpu },
      { name: "Chaos Testing", href: "/chaos", icon: Flame },
    ],
  },
  {
    title: "Advanced",
    items: [
      { name: "Replay", href: "/replay", icon: PlayCircle },
      { name: "Multi-Modal", href: "/multimodal", icon: Image },
      { name: "IDE Traces", href: "/ide-traces", icon: FileCode },
      { name: "Intents", href: "/intents", icon: CheckCircle },
      { name: "Embed", href: "/embed", icon: Code },
      { name: "Marketplace", href: "/marketplace", icon: Store },
      { name: "Discovery", href: "/discovery", icon: ScanSearch },
      { name: "Onboarding", href: "/onboarding", icon: Rocket },
    ],
  },
  {
    title: "Settings",
    items: [
      { name: "Settings", href: "/settings", icon: Settings },
    ],
  },
];

function SectionGroup({
  section,
  collapsed,
  pathname,
}: {
  section: NavSection;
  collapsed: boolean;
  pathname: string;
}) {
  const hasActiveItem = section.items.some(
    (item) => pathname === item.href || pathname.startsWith(`${item.href}/`)
  );
  const [open, setOpen] = React.useState(hasActiveItem);

  return (
    <div>
      {!collapsed && (
        <button
          onClick={() => setOpen(!open)}
          className="flex w-full items-center justify-between px-3 py-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground hover:text-foreground transition-colors"
        >
          <span>{section.title}</span>
          <ChevronDown
            className={cn(
              "h-3 w-3 transition-transform",
              !open && "-rotate-90"
            )}
          />
        </button>
      )}
      {(collapsed || open) && (
        <div className="space-y-0.5">
          {section.items.map((item) => {
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
                <Tooltip key={item.href}>
                  <TooltipTrigger asChild>{linkContent}</TooltipTrigger>
                  <TooltipContent side="right">{item.name}</TooltipContent>
                </Tooltip>
              );
            }

            return <div key={item.href}>{linkContent}</div>;
          })}
        </div>
      )}
    </div>
  );
}

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
        <nav className="flex-1 space-y-3 overflow-y-auto px-2 py-4">
          {sections.map((section) => (
            <SectionGroup
              key={section.title}
              section={section}
              collapsed={collapsed}
              pathname={pathname}
            />
          ))}
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
