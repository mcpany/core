/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Inspector Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock the WebSocket connection to inject Traffic logs
    await page.routeWebSocket(/\/api\/v1\/ws\/logs/, ws => {
      ws.onConnect(connection => {
        // Send a Traffic log
        const trafficPayload = {
          method: "tools/list",
          timestamp: new Date().toISOString(),
          duration: "10ms",
          request: { params: {} },
          result: { tools: [] },
          status: "success"
        };
        const logEntry = {
          id: "traffic-e2e-1",
          timestamp: new Date().toISOString(),
          level: "INFO",
          message: JSON.stringify(trafficPayload),
          source: "INSPECTOR"
        };
        connection.send(JSON.stringify(logEntry));
      });
    });

    await page.goto('/inspector');
  });

  test('should display inspector title', async ({ page }) => {
    await expect(page).toHaveTitle(/MCPAny/);
    await expect(page.getByRole('heading', { name: 'Inspector' })).toBeVisible({ timeout: 10000 });
  });

  test('should display traffic events', async ({ page }) => {
    // Wait for the event to appear
    await expect(page.getByText('tools/list')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('OK')).toBeVisible();
  });

  test('should show details when event selected', async ({ page }) => {
    // Click the event
    await page.getByText('tools/list').click();

    // Verify details pane
    await expect(page.getByText('Request Params')).toBeVisible();
    await expect(page.getByText('Response Result')).toBeVisible();

    // Check request JSON
    await expect(page.locator('.row-span-1').first()).toContainText('params');

    // Check result JSON
    await expect(page.locator('.row-span-1').last()).toContainText('tools');
  });
});
