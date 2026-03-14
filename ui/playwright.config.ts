import { defineConfig, devices } from '@playwright/test';

const PORT = process.env.TEST_PORT || 9111;
const BASE_URL = process.env.PLAYWRIGHT_BASE_URL || `http://localhost:${PORT}`;
const NEXT_DEV_COMMAND = process.env.NEXT_DEV_COMMAND || `npx next dev -p ${PORT}`;
const TEST_MATCH = process.env.PLAYWRIGHT_TEST_MATCH;

export default defineConfig({
  testDir: './tests',
  testMatch: TEST_MATCH ? new RegExp(TEST_MATCH) : ['**/*.spec.ts'],
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
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
      command: NEXT_DEV_COMMAND,
      url: BASE_URL,
      reuseExistingServer: true,
      timeout: 300000,
      env: {
        BACKEND_URL: process.env.BACKEND_URL || 'http://localhost:50050',
        MCPANY_API_KEY: process.env.MCPANY_API_KEY || 'test-token',
      },
    },
});
