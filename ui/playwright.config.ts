/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  testMatch: ['**/*.spec.ts'], // Changed to match all specs
  // testIgnore: '**/generate_docs_screenshots.spec.ts',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  outputDir: 'test-results/artifacts',
  reporter: [['line'], ['html', { outputFolder: 'playwright-report/html' }], ['json', { outputFile: 'test-results/test-results.json' }]],
  timeout: 60000,
  expect: {
    timeout: 15000,
  },
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:9002',
    extraHTTPHeaders: {
      'X-API-Key': process.env.NEXT_PUBLIC_MCPANY_API_KEY || 'test-token',
    },
    trace: 'on-first-retry',
    colorScheme: 'dark',
    actionTimeout: 15000,
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
  webServer: {
    command: `BACKEND_URL=${process.env.BACKEND_URL || 'http://localhost:50050'} npm run dev`,
    url: 'http://localhost:9002',
    reuseExistingServer: true,
    env: {
      BACKEND_URL: process.env.BACKEND_URL || 'http://localhost:50050',
    },
  },
});
