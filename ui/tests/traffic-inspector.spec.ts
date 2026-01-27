/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Traffic Inspector captures JSON-RPC traffic', async ({ page }) => {
  // 1. Go to Inspector
  await page.goto('/inspector');

  // 2. Verify UI elements
  await expect(page.getByRole('heading', { name: 'Traffic Inspector' })).toBeVisible();
  await expect(page.getByText('Live')).toBeVisible();

  // 3. Trigger a JSON-RPC request to the server
  // This should be captured by the LoggingMiddleware and broadcasted to the Inspector via WebSocket.
  await page.evaluate(async () => {
      await fetch('/', {
          method: 'POST',
          headers: {
              'Content-Type': 'application/json'
          },
          body: JSON.stringify({
              jsonrpc: '2.0',
              id: 'inspector-test-1',
              method: 'tools/list',
              params: {}
          })
      });
  });

  // 4. Verify the traffic appears in the Inspector list
  // It might take a moment for the log to arrive
  await expect(page.getByText('tools/list')).toBeVisible({ timeout: 10000 });

  // 5. Click it to see details
  await page.getByText('tools/list').first().click();

  // 6. Verify details
  await expect(page.getByText('inspector-test-1')).toBeVisible();
});
