/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import type {NextConfig} from 'next';
import path from 'path';
import fs from 'fs';

const isBazelBuild = Boolean(process.env.BAZEL_WORKSPACE || process.env.JS_BINARY__CHDIR);

const nextConfig: NextConfig = {
  output: isBazelBuild ? undefined : 'standalone',
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
  webpack: (config) => {

    // Explicitly add alias for @proto to resolve external directory
    // In Docker, we copy proto to ./proto. Locally, it maps to ../proto.
    const localProto = path.join(__dirname, 'proto');
    const rootProto = path.join(__dirname, '../proto');
    const protoPath = fs.existsSync(localProto) ? localProto : rootProto;

    // In Bazel-managed node_modules, 'ms' (a dependency of 'debug') is nested
    // under .aspect_rules_js/debug@.../node_modules/ms but not at the top level.
    // With symlinks=false, webpack can't find it via normal traversal, so we
    // provide an explicit alias for it.
    const bazelMsPath = path.join(__dirname, 'node_modules/.aspect_rules_js/ms@2.1.3/node_modules/ms/index.js');
    const msAlias = fs.existsSync(bazelMsPath) ? { 'ms': bazelMsPath } : {};

    config.resolve.alias = {
      ...config.resolve.alias,
      '@proto': protoPath,
      '@google': path.join(protoPath, 'google'),
      // highlight.js v11 removed deprecated language aliases used by react-syntax-highlighter v16
      'highlight.js/lib/languages/c-like': false,
      'highlight.js/lib/languages/htmlbars': false,
      'highlight.js/lib/languages/sql_more': false,
      ...msAlias,
    };
    // Important: Disable symlink resolution to prevent Webpack from resolving symlinks to their real path (which is outside the project)
    config.resolve.symlinks = false;

    // Ignore fsevents missing error in webpack build
    config.externals = [...(config.externals || []), 'fsevents'];

    return config;
  },
  // WebSockets need to be proxied via next.config.ts because NextResponse.rewrite in middleware
  // doesn't support WebSocket upgrade headers properly in all environments.
  async rewrites() {
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';
    // Remove protocol (http/https) to construct the wss/ws url
    const wsUrl = backendUrl.replace(/^http/, 'ws');
    return [
      {
        source: '/api/v1/ws/:path*',
        destination: `${wsUrl}/api/v1/ws/:path*`,
      },
    ];
  },
};

export default nextConfig;
