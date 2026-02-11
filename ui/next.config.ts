/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import type {NextConfig} from 'next';
import path from 'path';

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
  transpilePackages: ['@bufbuild/protobuf', 'long', 'browser-headers', '@improbable-eng/grpc-web'],
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
  webpack: (config) => {
    // Explicitly add alias for @proto to resolve generated source directory
    const srcProto = path.join(__dirname, 'src/proto');
    // Local dev fallback (if we run locally without Docker/src-gen)
    const rootProto = path.join(__dirname, '../proto');

    // In Docker, we generate to src/proto. Locally, we might use ../proto.
    // Check if src/proto/config exists (it should in Docker)
    const fs = require('fs');
    const protoPath = fs.existsSync(path.join(srcProto, 'config')) ? srcProto : rootProto;

    config.resolve.alias = {
      ...config.resolve.alias,
      '@proto': protoPath,
      '@google': path.join(protoPath, 'google'),
    };

    // Note: We intentionally do NOT alias libraries like 'long' or '@bufbuild/protobuf' here.
    // Since we are now generating code into src/proto (inside the project root),
    // standard Node.js module resolution should work correctly for finding dependencies
    // in node_modules without explicit aliases.
    // Explicit aliases can sometimes cause issues with ESM/CJS interop if not careful.

    config.resolve.symlinks = false;
    return config;
  },
};

export default nextConfig;
