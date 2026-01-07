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
    // We use a symlink in Docker to make it appear inside the project
    config.resolve.alias = {
      ...config.resolve.alias,
      '@proto': path.resolve(__dirname, 'proto'),
      '@google': path.resolve(__dirname, 'proto/google'),
    };
    // Important: Disable symlink resolution to prevent Webpack from resolving symlinks to their real path (which is outside the project)
    config.resolve.symlinks = false;
    return config;
  },
  async rewrites() {
    console.log("DEBUG: BACKEND_URL =", process.env.BACKEND_URL);
    return [
      {
        source: '/api/v1/:path*',
        destination: `${process.env.BACKEND_URL || 'http://localhost:8080'}/api/v1/:path*`,
      },
      {
        source: '/mcpany.api.v1.RegistrationService/:path*',
        destination: `${process.env.BACKEND_URL || 'http://localhost:8080'}/mcpany.api.v1.RegistrationService/:path*`,
      },
    ];
  },
};

export default nextConfig;
