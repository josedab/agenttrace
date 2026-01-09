# ADR-010: Next.js BFF with Server Components

## Status

Accepted

## Context

AgentTrace requires a web frontend for:
- Viewing traces, observations, and analytics
- Managing prompts, datasets, and experiments
- Configuring projects, exporters, and integrations
- User authentication and organization management

The frontend needs to:
1. Communicate with the Go backend API
2. Handle user authentication (OAuth, SSO)
3. Provide excellent developer experience for the engineering team
4. Support server-side rendering for performance and SEO
5. Enable real-time updates for trace streaming

### Alternatives Considered

1. **SPA (React + Vite) calling Go API directly**
   - Pros: Simple architecture, frontend independence
   - Cons: CORS complexity, auth token management in browser, no SSR

2. **Go serving HTML templates**
   - Pros: Single codebase, simple deployment
   - Cons: Poor developer experience, limited interactivity, Go template limitations

3. **Next.js Pages Router**
   - Pros: Proven, good ecosystem
   - Cons: Less efficient than Server Components, older patterns

4. **Next.js App Router with Server Components** (chosen)
   - Pros: Modern React patterns, reduced client JS, built-in BFF capabilities
   - Cons: Learning curve, ecosystem still maturing

## Decision

We use **Next.js 15 with App Router** as both the frontend framework and Backend-for-Frontend (BFF):

### Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Browser                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                    React (Client Components)                    │ │
│  │  - Interactive UI                                               │ │
│  │  - Real-time updates                                            │ │
│  │  - Client state (Zustand)                                       │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Next.js Server                                  │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                    Server Components                            │ │
│  │  - Data fetching                                                │ │
│  │  - Auth validation                                              │ │
│  │  - Initial render                                               │ │
│  └────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                      API Routes (BFF)                           │ │
│  │  - /api/auth/* (NextAuth)                                       │ │
│  │  - /api/graphql (proxy)                                         │ │
│  │  - /api/export/* (file downloads)                               │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Go Backend API                                │
│  - /api/public/* (REST)                                             │
│  - /graphql                                                          │
└─────────────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
web/
├── app/                      # App Router
│   ├── layout.tsx           # Root layout (providers, nav)
│   ├── page.tsx             # Home page
│   ├── (auth)/              # Auth group
│   │   ├── login/
│   │   └── signup/
│   ├── (dashboard)/         # Dashboard group
│   │   ├── traces/
│   │   │   ├── page.tsx     # Server Component - list
│   │   │   └── [id]/
│   │   │       └── page.tsx # Server Component - detail
│   │   ├── prompts/
│   │   ├── datasets/
│   │   └── settings/
│   └── api/                 # BFF routes
│       ├── auth/
│       │   └── [...nextauth]/
│       ├── graphql/
│       │   └── route.ts     # GraphQL proxy
│       └── export/
│           └── route.ts
├── components/
│   ├── ui/                  # Radix UI primitives
│   ├── traces/              # Trace components
│   └── shared/              # Shared components
├── lib/
│   ├── api.ts               # Backend API client
│   ├── graphql.ts           # GraphQL client
│   └── auth.ts              # Auth utilities
└── hooks/
    └── use-traces.ts        # React Query hooks
```

### Server Components for Data Fetching

```tsx
// app/(dashboard)/traces/page.tsx
import { getTraces } from '@/lib/api';
import { TraceList } from '@/components/traces/trace-list';

export default async function TracesPage({
  searchParams,
}: {
  searchParams: { page?: string; filter?: string };
}) {
  // Data fetched on server - no client JS needed
  const traces = await getTraces({
    page: parseInt(searchParams.page || '1'),
    filter: searchParams.filter,
  });

  return (
    <div>
      <h1>Traces</h1>
      <TraceList traces={traces} />
    </div>
  );
}
```

### Client Components for Interactivity

```tsx
// components/traces/trace-list.tsx
'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';

export function TraceList({ initialTraces }: { initialTraces: Trace[] }) {
  const [filter, setFilter] = useState('');

  // Client-side refresh with React Query
  const { data: traces } = useQuery({
    queryKey: ['traces', filter],
    queryFn: () => fetchTraces(filter),
    initialData: initialTraces,
  });

  return (
    <>
      <FilterInput value={filter} onChange={setFilter} />
      {traces.map(trace => <TraceRow key={trace.id} trace={trace} />)}
    </>
  );
}
```

### BFF API Routes

```typescript
// app/api/graphql/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth';

export async function POST(request: NextRequest) {
  // 1. Validate session
  const session = await getServerSession();
  if (!session) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  // 2. Forward to Go backend with auth
  const body = await request.json();
  const response = await fetch(`${process.env.API_URL}/graphql`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${session.accessToken}`,
    },
    body: JSON.stringify(body),
  });

  // 3. Return response
  const data = await response.json();
  return NextResponse.json(data);
}
```

## Consequences

### Positive

- **Reduced client JS**: Server Components render on server, less JavaScript sent to browser
- **Built-in BFF**: API routes handle auth, proxying, and data transformation
- **Excellent DX**: Hot reload, TypeScript, great error messages
- **Performance**: Streaming SSR, automatic code splitting
- **Auth at edge**: NextAuth handles OAuth flows, session management
- **Unified codebase**: Frontend and BFF in same repo

### Negative

- **Learning curve**: Server Components require new mental model
- **Ecosystem maturity**: Some libraries not yet compatible with App Router
- **Build complexity**: Next.js builds can be slow for large apps
- **Debugging**: Server Component errors can be harder to debug
- **Lock-in**: Tightly coupled to Next.js/Vercel patterns

### Neutral

- Middleware handles auth checks before page load
- React Query for client-side data fetching
- Zustand for client state management
- Radix UI + Tailwind for components

## State Management Strategy

| State Type | Solution | Example |
|------------|----------|---------|
| Server state | React Query | Traces, prompts, datasets |
| Client state | Zustand | UI preferences, filters |
| Form state | React Hook Form | Create/edit forms |
| URL state | Next.js searchParams | Pagination, filters |

## Authentication Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Browser │────►│ NextAuth │────►│  OAuth   │────►│ Callback │
│          │     │ /api/auth│     │ Provider │     │          │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
                                                         │
                                                         ▼
                                                  ┌──────────────┐
                                                  │ Session JWT  │
                                                  │ in HttpOnly  │
                                                  │ cookie       │
                                                  └──────────────┘
```

## References

- [Next.js App Router](https://nextjs.org/docs/app)
- [React Server Components](https://react.dev/reference/rsc/server-components)
- [NextAuth.js](https://next-auth.js.org/)
- [TanStack Query](https://tanstack.com/query/latest)
