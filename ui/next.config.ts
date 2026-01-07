/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import type {NextConfig} from 'next';

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
  async rewrites() {
    console.error("DEBUG: BACKEND_URL =", process.env.BACKEND_URL);
    // Default to server:50050 in CI (Docker), localhost:50050 otherwise
    const defaultBackend = process.env.CI ? 'http://server:50050' : 'http://localhost:50050';
    return [
      {
        source: '/api/v1/:path*',
        destination: `${process.env.BACKEND_URL || defaultBackend}/api/v1/:path*`,
      },
      {
        source: '/mcpany.api.v1.RegistrationService/:path*',
        destination: `${process.env.BACKEND_URL || 'http://localhost:8080'}/mcpany.api.v1.RegistrationService/:path*`,
      },
    ];
  },
};

export default nextConfig;
