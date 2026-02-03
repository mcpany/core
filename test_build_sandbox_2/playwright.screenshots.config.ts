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
  workers: 4,
  // Use the webServer config from the base config (npm run dev)
  webServer: config.webServer,
});
