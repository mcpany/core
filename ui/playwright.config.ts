/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineConfig, devices } from '@playwright/test';
import os from 'os';

// Use TEST_PORT env var if set, otherwise default to 9111
const PORT = process.env.TEST_PORT || 9111;
const BASE_URL = process.env.PLAYWRIGHT_BASE_URL || `http://localhost:${PORT}`;

// Determine Backend URL
// In CI (Docker Compose), the server is available at http://server:50050
// Locally, it's at http://127.0.0.1:50050 (default) or whatever BACKEND_URL is set to.
const isCI = !!process.env.CI;
const defaultBackendUrl = isCI ? 'http://server:50050' : 'http://127.0.0.1:50050';
const BACKEND_URL = process.env.BACKEND_URL || defaultBackendUrl;

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
        // Pass BACKEND_URL explicitly to Next.js dev server
        command: `BACKEND_URL=${BACKEND_URL} npx next dev -p ${PORT}`,
        url: BASE_URL,
        reuseExistingServer: false,
        env: {
          BACKEND_URL: BACKEND_URL,
          MCPANY_API_KEY: process.env.MCPANY_API_KEY || 'test-token',
        },
      },
});
