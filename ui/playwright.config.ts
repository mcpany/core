/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineConfig, devices } from '@playwright/test';
import os from 'os';

// Use TEST_PORT env var if set, otherwise default to 9111
const PORT = process.env.TEST_PORT || 9111;
const BASE_URL = process.env.PLAYWRIGHT_BASE_URL || `http://localhost:${PORT}`;

// Command to seed DB, start Backend, and start Frontend
// Note: We move up to root to run go commands.
const SEED_CMD = 'cd .. && go run server/cmd/mcpctl/main.go seed';
const START_BACKEND_CMD = 'cd .. && go run server/cmd/server/main.go run';
const START_FRONTEND_CMD = `npx next dev -p ${PORT}`;

// Composite command: Seed -> Start Backend (bg) -> Wait -> Start Frontend
// We use a trap to kill the backend when this script exits?
// Playwright kills the process tree, so background jobs started by shell should die if they are in the same group.
const WEB_SERVER_COMMAND = `${SEED_CMD} && (${START_BACKEND_CMD} &) && sleep 5 && ${START_FRONTEND_CMD}`;

export default defineConfig({
  testDir: './tests',
  testMatch: ['**/*.spec.ts'], // Changed to match all specs
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1, // Run serially to avoid state collisions in shared backend
  outputDir: 'test-results/artifacts',
  reporter: [['line'], ['json', { outputFile: 'test-results/test-results.json' }]],
  timeout: 120000,
  expect: {
    timeout: 30000,
  },
  use: {
    baseURL: BASE_URL,
    trace: 'on-first-retry',
    colorScheme: 'dark',
    actionTimeout: 30000,
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
        command: WEB_SERVER_COMMAND,
        url: BASE_URL,
        reuseExistingServer: false,
        env: {
          BACKEND_URL: process.env.BACKEND_URL || 'http://localhost:50050',
          MCPANY_API_KEY: process.env.MCPANY_API_KEY || 'test-token',
        },
      },
});
