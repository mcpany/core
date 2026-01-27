/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/tests/setup.ts'],
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@proto/api/v1/registration': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
      '@proto/config/v1/upstream_service': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
      '@proto/config/v1/tool': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
      '@proto/config/v1/resource': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
      '@proto/config/v1/prompt': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
      '@proto/admin/v1/admin': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
    },
  },
})
