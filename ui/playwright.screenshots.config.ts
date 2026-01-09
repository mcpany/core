/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineConfig, devices } from '@playwright/test';
import baseConfig from './playwright.config';

export default defineConfig({
  ...baseConfig,
  testMatch: ['**/generate_docs_screenshots.spec.ts'],
  testIgnore: [], // Explicitly un-ignore for this specific config
  webServer: {
    command: `npm run build && npm run start -- -p 9002`,
    url: 'http://localhost:9002',
    reuseExistingServer: false, // Ensure we build fresh for screenshots
    timeout: 120000, // Build might take time
    env: {
      BACKEND_URL: process.env.BACKEND_URL || 'http://localhost:50050',
    },
  },
});
