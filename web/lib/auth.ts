import NextAuth from 'next-auth';
import type { NextAuthConfig } from 'next-auth';
import 'next-auth/jwt';
import CredentialsProvider from 'next-auth/providers/credentials';
import GoogleProvider from 'next-auth/providers/google';
import GitHubProvider from 'next-auth/providers/github';

const API_URL =
  process.env.API_INTERNAL_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const OAUTH_CALLBACK_SECRET = process.env.OAUTH_CALLBACK_SECRET;

interface BackendAuthResponse {
  accessToken: string;
  refreshToken: string;
  expiresAt: string;
  user: {
    id: string;
    email: string;
    name: string | null;
    image: string | null;
  };
}

const providers: NextAuthConfig['providers'] = [
  CredentialsProvider({
    name: 'credentials',
    credentials: {
      email: { label: 'Email', type: 'email' },
      password: { label: 'Password', type: 'password' },
    },
    async authorize(credentials) {
      if (!credentials?.email || !credentials?.password) {
        return null;
      }

      try {
        const response = await fetch(`${API_URL}/api/auth/login`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            email: credentials.email,
            password: credentials.password,
          }),
        });

        if (!response.ok) {
          return null;
        }

        const data = (await response.json()) as BackendAuthResponse;

        return {
          id: data.user.id,
          email: data.user.email,
          name: data.user.name,
          image: data.user.image,
          accessToken: data.accessToken,
          refreshToken: data.refreshToken,
          accessTokenExpires: new Date(data.expiresAt).getTime(),
        };
      } catch {
        return null;
      }
    },
  }),
];

if (process.env.GOOGLE_CLIENT_ID && process.env.GOOGLE_CLIENT_SECRET && OAUTH_CALLBACK_SECRET) {
  providers.push(
    GoogleProvider({
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
    })
  );
}

if (process.env.GITHUB_CLIENT_ID && process.env.GITHUB_CLIENT_SECRET && OAUTH_CALLBACK_SECRET) {
  providers.push(
    GitHubProvider({
      clientId: process.env.GITHUB_CLIENT_ID!,
      clientSecret: process.env.GITHUB_CLIENT_SECRET!,
    })
  );
}

export const authConfig: NextAuthConfig = {
  providers,
  basePath: '/auth',
  trustHost: process.env.AUTH_TRUST_HOST === 'true',
  pages: {
    signIn: '/sign-in',
    signOut: '/sign-out',
    error: '/error',
    newUser: '/onboarding',
  },
  callbacks: {
    async jwt({ token, user, account, profile }) {
      if (user) {
        token.id = user.id;
        token.accessToken = (user as { accessToken?: string }).accessToken;
        token.refreshToken = (user as { refreshToken?: string }).refreshToken;
        token.accessTokenExpires = (user as { accessTokenExpires?: number }).accessTokenExpires;
      }

      // Handle OAuth providers
      if (account?.provider === 'google' || account?.provider === 'github') {
        try {
          const response = await fetch(`${API_URL}/api/auth/callback/${account.provider}`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'X-AgentTrace-OAuth-Secret': OAUTH_CALLBACK_SECRET ?? '',
            },
            body: JSON.stringify({
              providerAccountId: account.providerAccountId,
              email: profile?.email ?? user?.email,
              name: profile?.name ?? user?.name,
              image:
                (profile as { picture?: string; avatar_url?: string } | undefined)?.picture ??
                (profile as { avatar_url?: string } | undefined)?.avatar_url ??
                user?.image,
              accessToken: account.access_token,
            }),
          });

          if (!response.ok) {
            throw new Error(`OAuth exchange failed with status ${response.status}`);
          }

          const data = (await response.json()) as BackendAuthResponse;
          token.accessToken = data.accessToken;
          token.refreshToken = data.refreshToken;
          token.accessTokenExpires = new Date(data.expiresAt).getTime();
          token.id = data.user.id;
        } catch (error) {
          console.error('OAuth callback error:', error);
          token.accessToken = undefined;
          token.refreshToken = undefined;
          token.accessTokenExpires = undefined;
          token.error = 'OAuthAccountError';
        }
      }

      if (token.accessTokenExpires && Date.now() < token.accessTokenExpires - 60_000) {
        return token;
      }

      if (!token.refreshToken) {
        return token;
      }

      try {
        const response = await fetch(`${API_URL}/api/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refreshToken: token.refreshToken }),
        });
        if (!response.ok) {
          throw new Error(`refresh failed with status ${response.status}`);
        }

        const refreshed = (await response.json()) as Omit<BackendAuthResponse, 'user'>;
        token.accessToken = refreshed.accessToken;
        token.refreshToken = refreshed.refreshToken;
        token.accessTokenExpires = new Date(refreshed.expiresAt).getTime();
        token.error = undefined;
      } catch {
        token.accessToken = undefined;
        token.refreshToken = undefined;
        token.accessTokenExpires = undefined;
        token.error = 'RefreshAccessTokenError';
      }

      return token;
    },
    async session({ session, token }) {
      if (token) {
        session.user.id = token.id as string;
        (session as { accessToken?: string }).accessToken = token.accessToken as string;
        session.error = token.error;
      }
      return session;
    },
    async redirect({ url, baseUrl }) {
      // Allows relative callback URLs
      if (url.startsWith('/')) return `${baseUrl}${url}`;
      // Allows callback URLs on the same origin
      else if (new URL(url).origin === baseUrl) return url;
      return baseUrl;
    },
  },
  session: {
    strategy: 'jwt',
    maxAge: 30 * 24 * 60 * 60, // 30 days
  },
  debug: process.env.NODE_ENV === 'development',
};

export const { handlers, auth, signIn, signOut } = NextAuth(authConfig);

// Types
declare module 'next-auth' {
  interface Session {
    user: {
      id: string;
      email: string;
      name: string | null;
      image: string | null;
    };
    accessToken?: string;
    error?: 'RefreshAccessTokenError' | 'OAuthAccountError';
  }

  interface User {
    id: string;
    email: string;
    name: string | null;
    image: string | null;
    accessToken?: string;
    refreshToken?: string;
    accessTokenExpires?: number;
  }
}

declare module 'next-auth/jwt' {
  interface JWT {
    id: string;
    accessToken?: string;
    refreshToken?: string;
    accessTokenExpires?: number;
    error?: 'RefreshAccessTokenError' | 'OAuthAccountError';
  }
}
