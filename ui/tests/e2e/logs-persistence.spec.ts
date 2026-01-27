/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Logs Persistence', () => {
  test('should persist logs after reload', async ({ page }) => {
    const uniqueMessage = `Persistent Log ${Date.now()}`;

    // Mock the WebSocket connection
    await page.routeWebSocket(/\/api\/v1\/ws\/logs/, ws => {
         // Send a unique log immediately on connection
         const logEntry = {
            id: `persist-test-${Date.now()}`,
            timestamp: new Date().toISOString(),
            level: "INFO",
            message: uniqueMessage,
            source: "persistence-test"
          };
          // usage based on existing logs.spec.ts
          ws.send(JSON.stringify(logEntry));
    });

    // 1. Visit logs page
    await page.goto('/logs');

    // 2. Verify log is visible
    await expect(page.getByText(uniqueMessage)).toBeVisible({ timeout: 10000 });

    // 3. Reload the page
    // We want to ensure no NEW logs are sent, so we can verify persistence.
    // We can unroute or route to a handler that does nothing.
    await page.unrouteAll();

    await page.routeWebSocket(/\/api\/v1\/ws\/logs/, _ws => {
       // Do nothing -> no logs sent
    });

    await page.reload();

    // 4. Verify log is still visible (from persistence)
    await expect(page.getByText(uniqueMessage)).toBeVisible({ timeout: 10000 });
  });

  test('should clear persistence when logs are cleared', async ({ page }) => {
    const uniqueMessage = `Clear Test ${Date.now()}`;

    await page.routeWebSocket(/\/api\/v1\/ws\/logs/, ws => {
          const logEntry = {
            id: `clear-test-${Date.now()}`,
            timestamp: new Date().toISOString(),
            level: "INFO",
            message: uniqueMessage,
            source: "persistence-test"
          };
          ws.send(JSON.stringify(logEntry));
    });

    await page.goto('/logs');
    await expect(page.getByText(uniqueMessage)).toBeVisible();

    // Click Clear
    await page.getByRole('button', { name: 'Clear' }).click();
    await expect(page.getByText(uniqueMessage)).not.toBeVisible();

    // Reload and ensure it's still gone
    // Ensure no new logs are sent
    await page.unrouteAll();
    await page.routeWebSocket(/\/api\/v1\/ws\/logs/, _ws => {
       // Do nothing
    });

    await page.reload();
    await expect(page.getByText(uniqueMessage)).not.toBeVisible();
  });
});
