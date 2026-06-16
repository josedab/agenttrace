import type { NextConfig } from 'next';
import { compatibilityRedirects } from './lib/product-capabilities';

const publicApiUrl = process.env.NEXT_PUBLIC_API_URL || '';
const connectSources = new Set(["'self'"]);

try {
  const apiUrl = new URL(publicApiUrl);
  connectSources.add(apiUrl.origin);
  connectSources.add(`${apiUrl.protocol === 'https:' ? 'wss:' : 'ws:'}//${apiUrl.host}`);
} catch {
  // The Docker build validates NEXT_PUBLIC_API_URL; local development falls back above.
}

const nextConfig: NextConfig = {
  reactStrictMode: true,

  // Enable standalone output for Docker
  output: 'standalone',

  // Image optimization
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'avatars.githubusercontent.com',
      },
      {
        protocol: 'https',
        hostname: 'lh3.googleusercontent.com',
      },
    ],
  },

  // Experimental features
  experimental: {
    serverActions: {
      bodySizeLimit: '2mb',
    },
  },

  // Environment variables
  env: {
    NEXT_PUBLIC_API_URL: publicApiUrl,
  },

  // Redirects
  async redirects() {
    return [
      {
        source: '/',
        destination: '/traces',
        permanent: false,
      },
      ...compatibilityRedirects,
    ];
  },

  // Security headers
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'X-Frame-Options', value: 'DENY' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
          {
            key: 'Content-Security-Policy',
            value: `default-src 'self'; script-src 'self' 'unsafe-eval' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src ${Array.from(connectSources).join(' ')}; frame-ancestors 'none'`,
          },
          {
            key: 'Strict-Transport-Security',
            value: 'max-age=63072000; includeSubDomains; preload',
          },
        ],
      },
    ];
  },

  // Webpack configuration for monaco editor
  webpack: (config, { isServer }) => {
    if (!isServer) {
      config.resolve.fallback = {
        ...config.resolve.fallback,
        fs: false,
        path: false,
      };
    }
    return config;
  },
};

export default nextConfig;
