'use client';

import * as React from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { SessionProvider, signOut, useSession } from 'next-auth/react';

import { API_URL, getApiProjectId, setApiAccessToken, setApiProjectId } from '@/lib/api';

export function AuthProvider({ children }: { children: React.ReactNode }) {
  return (
    <SessionProvider basePath="/auth">
      <ApiAccessTokenBridge>{children}</ApiAccessTokenBridge>
    </SessionProvider>
  );
}

function ApiAccessTokenBridge({ children }: { children: React.ReactNode }) {
  const { data: session, status } = useSession();
  const [projectReady, setProjectReady] = React.useState(false);
  const queryClient = useQueryClient();

  React.useLayoutEffect(() => {
    setApiAccessToken(session?.accessToken);
    return () => setApiAccessToken(undefined);
  }, [session?.accessToken]);

  React.useEffect(() => {
    let cancelled = false;

    if (status === 'loading') return;
    if (status === 'authenticated' && (!session?.accessToken || session.error)) {
      queryClient.clear();
      setApiAccessToken(undefined);
      setApiProjectId(undefined);
      window.localStorage.removeItem('agenttrace.activeProjectId');
      setProjectReady(false);
      void signOut({
        callbackUrl: '/sign-in?error=SessionExpired',
      });
      return;
    }
    if (!session?.accessToken) {
      queryClient.clear();
      setApiProjectId(undefined);
      setProjectReady(true);
      return;
    }

    setProjectReady(false);
    const headers = new Headers({
      Authorization: ['Bearer', session.accessToken].join(' '),
    });

    void fetch(`${API_URL}/api/v1/projects`, { headers })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(`project request failed with status ${response.status}`);
        }

        const payload = (await response.json()) as {
          data?: Array<{ id: string }>;
        };
        const projects = payload.data ?? [];
        const storedProjectId = window.localStorage.getItem('agenttrace.activeProjectId');
        const selected = projects.find((project) => project.id === storedProjectId) ?? projects[0];

        if (!cancelled) {
          if (getApiProjectId() !== selected?.id) {
            queryClient.clear();
          }
          setApiProjectId(selected?.id);
          if (selected) {
            window.localStorage.setItem('agenttrace.activeProjectId', selected.id);
          }
        }
      })
      .catch(() => {
        if (!cancelled) setApiProjectId(undefined);
      })
      .finally(() => {
        if (!cancelled) setProjectReady(true);
      });

    return () => {
      cancelled = true;
    };
  }, [queryClient, session?.accessToken, session?.error, status]);

  if (status === 'loading' || (status === 'authenticated' && !projectReady)) {
    return null;
  }

  return children;
}
