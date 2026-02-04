/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineConfig, devices } from '@playwright/test';
import os from 'os';

// Use TEST_PORT env var if set, otherwise default to 9111
const PORT = process.env.TEST_PORT || 9111;
const BASE_URL = process.env.PLAYWRIGHT_BASE_URL || `http://localhost:${PORT}`;

// Helper to determine if we should start the backend
const shouldStartBackend = process.env.SKIP_BACKEND_STARTUP !== 'true';

// Construct the webServer command
// If local: start backend, wait for health, then start frontend
// If CI/Docker (SKIP_BACKEND_STARTUP=true): just start frontend (backend assumed running)
// Note: We avoid leading spaces in the command to prevent execution issues
let webServerCommand = `BACKEND_URL=${process.env.BACKEND_URL || 'http://localhost:50050'} npx next dev -p ${PORT}`;

if (shouldStartBackend) {
  webServerCommand = `../build/bin/server run --mcp-listen-address 0.0.0.0:50050 --api-key test-token > server.log 2>&1 & count=0; while ! curl -s http://localhost:50050/health > /dev/null; do if [ $count -ge 30 ]; then echo "Timeout waiting for backend"; exit 1; fi; echo "Waiting for backend..."; sleep 1; count=$((count+1)); done; ${webServerCommand}`;
}

console.log(`Using webServer command: ${webServerCommand}`);

export default defineConfig({
  testDir: './tests',
  testMatch: ['**/*.spec.ts'],
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 4 : 4,
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
        command: webServerCommand,
        url: BASE_URL,
        reuseExistingServer: false,
        stdout: 'pipe',
        stderr: 'pipe',
        env: {
          BACKEND_URL: process.env.BACKEND_URL || 'http://localhost:50050',
          MCPANY_API_KEY: process.env.MCPANY_API_KEY || 'test-token',
        },
      },
});
