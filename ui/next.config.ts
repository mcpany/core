/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import type {NextConfig} from 'next';
import path from 'path';
import fs from 'fs';

const nextConfig: NextConfig = {
  output: 'standalone',
  /* config options here */
  typescript: {
    ignoreBuildErrors: true,
  },
  eslint: {
    ignoreDuringBuilds: true,
  },
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'placehold.co',
        port: '',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 'images.unsplash.com',
        port: '',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 'picsum.photos',
        port: '',
        pathname: '/**',
      },
    ],
  },
  experimental: {
    // turbopack: {
    //   resolveAlias: {
    //     'canvas': './empty-module.ts',
    //   },
    //   rules: {
    //     '*.node': {
    //       loaders: ['node-loader'],
    //       as: '*.js',
    //     },
    //   },
    // },
  },
  async headers() {
    const isDev = process.env.NODE_ENV !== 'production';
    const csp = [
      "default-src 'self'",
      `script-src 'self' 'unsafe-inline'${isDev ? " 'unsafe-eval'" : ""} https://cdn.jsdelivr.net`, // Added cdn.jsdelivr.net
      "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net", // Added cdn.jsdelivr.net
      "img-src 'self' data: https://placehold.co https://images.unsplash.com https://picsum.photos",
      "font-src 'self' data: https://fonts.gstatic.com",
      "connect-src 'self' https://cdn.jsdelivr.net", // Added cdn.jsdelivr.net. Restricted http/https wildcards for security.
      "worker-src 'self' blob:", // Added worker-src
      "object-src 'none'",
      "base-uri 'self'",
      "form-action 'self'",
      "frame-ancestors 'none'",
      isDev ? "" : "upgrade-insecure-requests"
    ].filter(Boolean).join("; ");

    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'X-DNS-Prefetch-Control',
            value: 'on'
          },
          {
            key: 'Strict-Transport-Security',
            value: 'max-age=63072000; includeSubDomains; preload'
          },
          {
            key: 'X-XSS-Protection',
            value: '1; mode=block'
          },
          {
            key: 'X-Frame-Options',
            value: 'SAMEORIGIN'
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff'
          },
          {
            key: 'Referrer-Policy',
            value: 'origin-when-cross-origin'
          },
          {
            key: 'Permissions-Policy',
            value: 'geolocation=(), camera=(), microphone=(), payment=(), usb=(), vr=()'
          },
          {
            key: 'Content-Security-Policy',
            value: csp
          }
        ]
      }
    ];
  },
  async redirects() {
    return [
      {
        source: '/topology',
        destination: '/network',
        permanent: true,
      },
    ];
  },
  webpack: (config, { isServer }) => {

    // Explicitly add alias for @proto to resolve external directory
    // In Docker, we copy proto to /proto (system root). Locally, it maps to ../proto.
    // Use process.cwd() for reliability across environments.
    const cwd = process.cwd();
    const localProto = path.join(cwd, 'proto'); // e.g. /app/proto (if copied there)
    const siblingProto = path.join(cwd, '../proto'); // e.g. /proto (in Docker) or ../proto (local)

    let protoPath = siblingProto;
    if (fs.existsSync(localProto)) {
        protoPath = localProto;
    } else if (fs.existsSync('/proto')) {
        // Fallback to system root /proto if explicitly present (Docker specific)
        protoPath = '/proto';
    }

    if (isServer) {
        console.log(`[Webpack] Resolving @proto to: ${protoPath}`);
    }

    config.resolve.alias = {
      ...config.resolve.alias,
      '@proto': protoPath,
      '@google': path.join(protoPath, 'google'),
    };
    // Important: Disable symlink resolution to prevent Webpack from resolving symlinks to their real path (which is outside the project)
    config.resolve.symlinks = false;
    return config;
  },
  // rewrites moved to middleware.ts for runtime/dynamic proxy support
  // async rewrites() { ... }
};

export default nextConfig;
