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
    exclude: ['**/node_modules/**', '**/dist/**', '**/*.spec.ts'],
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@proto': path.resolve(__dirname, '../proto'),
      '@bufbuild/protobuf/wire': path.resolve(__dirname, './node_modules/@bufbuild/protobuf/dist/esm/wire/index.js'),
      '@improbable-eng/grpc-web': path.resolve(__dirname, './node_modules/@improbable-eng/grpc-web'),
      'browser-headers': path.resolve(__dirname, './node_modules/browser-headers'),
      'long': path.resolve(__dirname, './node_modules/long'),
    },
  },
})
