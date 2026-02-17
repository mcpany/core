/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import * as http from 'http';
import * as os from 'os';

// Real E2E test without mocking the MCP backend.
// We start a local "Upstream" server to test connectivity against.

test.describe('Credential Tester (Real)', () => {
  let server: http.Server;
  let port: number;

  test.beforeAll(async () => {
    // Start a local HTTP server to act as the upstream service
    server = http.createServer((req, res) => {
      // Check for auth header
      const auth = req.headers['authorization'];
      if (auth === 'Bearer secret-token-123') {
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ message: 'Authenticated!', user: 'admin' }));
      } else {
        res.writeHead(401, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ error: 'Unauthorized' }));
      }
    });

    await new Promise<void>((resolve) => {
      server.listen(0, '0.0.0.0', () => {
        port = (server.address() as any).port;
        console.log(`Test upstream server listening on port ${port}`);
        resolve();
      });
    });
  });

  test.afterAll(async () => {
    server.close();
  });

  test('should verify credential against a real local server', async ({ page }) => {
    // 1. Navigate to Credentials Page
    await page.goto('/credentials');

    // 2. Click New Credential
    await page.getByRole('button', { name: 'New Credential' }).click();

    // 3. Fill Form
    await page.getByPlaceholder('My Credential').fill('Real Test Credential');

    // Select Bearer Token type
    await page.getByRole('combobox', { name: 'Type' }).click();
    await page.getByRole('option', { name: 'Bearer Token' }).click();

    // Fill Token
    await page.getByPlaceholder('...bearer token...').fill('secret-token-123');

    // 4. Use Credential Tester
    // Use os.hostname() to ensure the backend container can address the test runner container in CI environments
    const hostname = process.env.CI ? os.hostname() : 'localhost';
    const testUrl = `http://${hostname}:${port}/test`;
    await page.getByPlaceholder('https://api.example.com/user').fill(testUrl);

    // Click Test
    await page.getByRole('button', { name: 'Test', exact: true }).click();

    // 5. Verify Result
    // Expect "Connection Successful"
    await expect(page.getByText('Connection Successful')).toBeVisible({ timeout: 10000 });

    // Expect response body to contain our message
    await expect(page.getByText('Authenticated!')).toBeVisible();
  });
});
