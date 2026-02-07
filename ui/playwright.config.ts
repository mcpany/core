/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineConfig, devices } from '@playwright/test';
import os from 'os';

// Use TEST_PORT env var if set, otherwise default to 9111
const PORT = process.env.TEST_PORT || 9111;
const BASE_URL = process.env.PLAYWRIGHT_BASE_URL || `http://localhost:${PORT}`;

export default defineConfig({
  testDir: './tests',
  testMatch: ['**/*.spec.ts'], // Changed to match all specs
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 4 : 4, // Limit workers to prevent server overload
  outputDir: 'test-results/artifacts',
  reporter: [['line'], ['json', { outputFile: 'test-results/test-results.json' }]],
  timeout: 120000,
  expect: {
    timeout: 15000,
  },
  use: {
    baseURL: BASE_URL,
    trace: 'on-first-retry',
    colorScheme: 'dark',
    actionTimeout: 15000,
    extraHTTPHeaders: {
      'X-API-Key': process.env.MCPANY_API_KEY || 'test-token',
    },
  },
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        launchOptions: {
          args: ['--disable-dev-shm-usage', '--no-sandbox', '--disable-setuid-sandbox', '--disable-gpu'],
        },
      },
    },
  ],
  webServer: process.env.SKIP_WEBSERVER
    ? undefined
    : {
        command: `BACKEND_URL=${process.env.BACKEND_URL || 'http://localhost:50050'} npx next dev -p ${PORT}`,
        url: BASE_URL,
        reuseExistingServer: false,
        timeout: 120000,
        env: {
          BACKEND_URL: process.env.BACKEND_URL || 'http://localhost:50050',
          MCPANY_API_KEY: process.env.MCPANY_API_KEY || 'test-token',
        },
      },
});
