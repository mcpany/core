/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import * as http from 'http';

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
      server.listen(0, () => {
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
    // Enter the local server URL
    // Ideally we would use localhost, but inside docker/CI, "localhost" refers to the container.
    // If the test runner and the browser are in the same network (or same host), localhost works.
    // In Playwright, browser connects to the test runner's localhost?
    // If we use `server.listen(0)`, it binds to 0.0.0.0 or ::?
    // Usually standard node http.createServer defaults to :: or 0.0.0.0.
    // Accessing via localhost:<port> should work if browser is local.
    // BUT if Playwright is running in a container and browser in another, or standard docker network:
    // If the backend (MCP server) executes the test (via `auth-test` endpoint),
    // it needs to be able to reach the test runner's process.
    // Wait, the MCP Server is running separately (via `make run` or similar in the environment).
    // The Test Runner starts an HTTP server.
    // The MCP Server (backend) needs to reach this HTTP server.
    // If they are in the same sandbox environment (localhost), it should work.
    // Assuming they share `localhost`.

    const testUrl = `http://localhost:${port}/test`;
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
