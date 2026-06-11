'use client';

import * as React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  BarChart3,
  ChevronLeft,
  ChevronRight,
  FlaskConical,
  Layers3,
  MessageSquareText,
  Settings,
  Users,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import {
  isCapabilityPath,
  primaryNavigationCapabilities,
  type ProductCapability,
} from '@/lib/product-capabilities';
import { cn } from '@/lib/utils';

const capabilityIcons = {
  trace: Layers3,
  eval: FlaskConical,
  prompt: MessageSquareText,
  cost: BarChart3,
  collaboration: Users,
} satisfies Record<ProductCapability['icon'], React.ComponentType<{ className?: string }>>;

function CapabilityLink({
  capability,
  collapsed,
  pathname,
}: {
  capability: ProductCapability;
  collapsed: boolean;
  pathname: string;
}) {
  const Icon = capabilityIcons[capability.icon];
  const isActive = isCapabilityPath(capability, pathname);

  const link = (
    <Link
      href={capability.canonicalPath}
      aria-current={isActive ? 'page' : undefined}
      className={cn(
        'group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors',
        isActive
          ? 'bg-foreground text-background shadow-sm'
          : 'text-muted-foreground hover:bg-muted hover:text-foreground',
        collapsed && 'justify-center px-2'
      )}
    >
      <Icon className="h-5 w-5 shrink-0" />
      {collapsed ? null : <span>{capability.shortName}</span>}
    </Link>
  );

  if (!collapsed) {
    return link;
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>{link}</TooltipTrigger>
      <TooltipContent side="right">
        <p className="font-medium">{capability.name}</p>
        <p className="max-w-56 text-xs text-muted-foreground">{capability.description}</p>
      </TooltipContent>
    </Tooltip>
  );
}

export function Sidebar() {
  const pathname = usePathname();
  const [collapsed, setCollapsed] = React.useState(false);

  return (
    <TooltipProvider delayDuration={100}>
      <aside
        className={cn(
          'flex h-full flex-col border-r bg-card transition-[width] duration-200',
          collapsed ? 'w-16' : 'w-64'
        )}
      >
        <div className="flex h-16 items-center border-b px-4">
          <Link href="/traces" className="flex min-w-0 items-center gap-2.5">
            <div className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-foreground text-sm font-semibold text-background">
              AT
            </div>
            {collapsed ? null : (
              <div className="min-w-0">
                <p className="truncate text-sm font-semibold tracking-tight">AgentTrace</p>
                <p className="truncate text-[11px] text-muted-foreground">Outcome observability</p>
              </div>
            )}
          </Link>
        </div>

        <nav aria-label="Primary" className="flex-1 space-y-1 overflow-y-auto px-2 py-4">
          {primaryNavigationCapabilities.map((capability) => (
            <CapabilityLink
              key={capability.id}
              capability={capability}
              collapsed={collapsed}
              pathname={pathname}
            />
          ))}
        </nav>

        <div className="space-y-1 border-t p-2">
          <Link
            href="/settings"
            aria-current={pathname.startsWith('/settings') ? 'page' : undefined}
            className={cn(
              'flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors',
              pathname.startsWith('/settings')
                ? 'bg-foreground text-background'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground',
              collapsed && 'justify-center px-2'
            )}
          >
            <Settings className="h-5 w-5 shrink-0" />
            {collapsed ? null : <span>Settings</span>}
          </Link>
          <Button
            variant="ghost"
            size="sm"
            className={cn('w-full text-muted-foreground', collapsed && 'px-2')}
            onClick={() => setCollapsed((value) => !value)}
            aria-label={collapsed ? 'Expand navigation' : 'Collapse navigation'}
          >
            {collapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <>
                <ChevronLeft className="mr-2 h-4 w-4" />
                Collapse
              </>
            )}
          </Button>
        </div>
      </aside>
    </TooltipProvider>
  );
}
