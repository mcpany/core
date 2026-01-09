/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import config from './playwright.config';
import { defineConfig } from '@playwright/test';

export default defineConfig({
  ...config,
  testIgnore: undefined, // Override ignore
  testMatch: '**/generate_docs_screenshots.spec.ts',
  webServer: {
    ...config.webServer,
    command: `PORT=9002 BACKEND_URL=${process.env.BACKEND_URL || 'http://localhost:50050'} node core/ui/server.js`,
    cwd: '.next/standalone',
  },
});
