import * as React from 'react';

import { cn } from '@/lib/utils';

interface PageHeaderProps {
  title: string;
  description?: string;
  children?: React.ReactNode;
  actions?: React.ReactNode;
  icon?: React.ElementType;
  className?: string;
}

export function PageHeader({
  title,
  description,
  children,
  actions,
  icon: Icon,
  className,
}: PageHeaderProps) {
  const headerActions = actions ?? children;

  return (
    <div
      className={cn(
        'flex flex-col gap-4 md:flex-row md:items-center md:justify-between',
        className
      )}
    >
      <div>
        <div className="flex items-center gap-2">
          {Icon && <Icon className="h-6 w-6 text-muted-foreground" />}
          <h1 className="text-2xl font-bold tracking-tight">{title}</h1>
        </div>
        {description && <p className="text-muted-foreground">{description}</p>}
      </div>
      {headerActions && <div className="flex items-center gap-2">{headerActions}</div>}
    </div>
  );
}
