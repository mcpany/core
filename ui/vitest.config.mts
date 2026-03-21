/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/tests/setup.ts'],
    exclude: ['**/node_modules/**', '**/dist/**', '**/*.spec.ts'],
    pool: 'vmThreads',
    // Force react-router-dom (and its deps) to be bundled through Vite's ESM
    // pipeline so that the "module" field (ESM build) is used instead of the
    // CJS "main" wrapper that re-exports the UMD bundle.  The UMD build has
    // its own copy of React, which breaks React context propagation in tests.
    server: {
      deps: {
        inline: ['react-router-dom', 'react-router', '@remix-run/router'],
      },
    },
  },
  resolve: {
    // Ensure only one copy of React (and its ecosystem) is used across test
    // files, preventing "useContext returned null" errors from mismatched
    // React instances when mixing CJS/ESM builds of react-router-dom.
    dedupe: ['react', 'react-dom', 'react-router-dom', 'react-router', '@remix-run/router'],
    // Prefer ESM builds ("module" field) so that react-router-dom uses the
    // same React instance as the rest of the test code.  Without this Vite
    // falls back to the CJS entrypoint which re-exports the UMD build that
    // carries its own React reference, causing Router context to be
    // invisible to hooks.
    conditions: ['import', 'browser', 'module'],
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@proto/api/v1/registration': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
      '@proto/config/v1/upstream_service': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
      '@proto/config/v1/tool': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
      '@proto/config/v1/resource': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
      '@proto/config/v1/prompt': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
      '@proto/config/v1/call': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
      '@proto/admin/v1/admin': path.resolve(__dirname, './src/mocks/proto/mock-proto.ts'),
    },
  },
})
