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
            key: 'Content-Security-Policy',
            // Note: 'unsafe-eval' is required for Next.js in some environments.
            // In a strict production environment, this should ideally be removed.
            value: "default-src 'self'; script-src 'self' 'unsafe-eval' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https://placehold.co https://images.unsplash.com https://picsum.photos; font-src 'self' data:; connect-src 'self' http://localhost:8080; object-src 'none'; base-uri 'self';"
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

    config.resolve.alias = {
      ...config.resolve.alias,
      '@proto': protoPath,
      '@google': path.join(protoPath, 'google'),
    };
    // Important: Disable symlink resolution to prevent Webpack from resolving symlinks to their real path (which is outside the project)
    config.resolve.symlinks = false;
    return config;
  },
  async rewrites() {
    const backendUrl = process.env.BACKEND_URL || 'http://mcpany:50050';
    console.log("DEBUG: BACKEND_URL env:", process.env.BACKEND_URL);
    console.log("DEBUG: Using backend URL:", backendUrl);
    return [
      {
        source: '/doctor',
        destination: `${backendUrl}/doctor`,
      },
      {
        source: '/api/v1/:path*',
        destination: `${backendUrl}/api/v1/:path*`,
      },
      {
        source: '/mcpany.api.v1.RegistrationService/:path*',
        destination: `${process.env.BACKEND_URL || 'http://localhost:8080'}/mcpany.api.v1.RegistrationService/:path*`,
      },
      {
        source: '/auth/:path*',
        destination: `${process.env.BACKEND_URL || 'http://localhost:8080'}/auth/:path*`,
      },
    ];
  },
};

export default nextConfig;
