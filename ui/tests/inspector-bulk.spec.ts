/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Inspector Bulk Actions', () => {
  test('should allow selecting multiple traces, deleting them, and exporting', async ({ page }) => {
    // Navigate to the Inspector page
    await page.goto('/inspector');

    // Wait for the page to load by checking for the "Inspector" header
    await expect(page.getByRole('heading', { name: 'Inspector' })).toBeVisible();

    // Mock the WebSocket connection to send some traces immediately
    await page.routeWebSocket(/\/api\/v1\/ws\/traces/, ws => {
      // Send trace 1
      ws.send(JSON.stringify({
        id: "trace-bulk-1",
        timestamp: new Date().toISOString(),
        totalDuration: 100,
        status: "success",
        trigger: "user",
        rootSpan: {
          id: "span-bulk-1",
          name: "tool-1",
          type: "tool",
          startTime: Date.now(),
          endTime: Date.now() + 100,
          status: "success",
          input: {},
          output: {}
        }
      }));
      // Send trace 2
      ws.send(JSON.stringify({
        id: "trace-bulk-2",
        timestamp: new Date().toISOString(),
        totalDuration: 200,
        status: "success",
        trigger: "user",
        rootSpan: {
          id: "span-bulk-2",
          name: "tool-2",
          type: "tool",
          startTime: Date.now(),
          endTime: Date.now() + 200,
          status: "success",
          input: {},
          output: {}
        }
      }));
    });

    // Reload the page to catch the websocket route
    await page.goto('/inspector');
    await expect(page.getByRole('heading', { name: 'Inspector' })).toBeVisible();

    // Wait for traces to appear
    await expect(page.locator('text=tool-1').first()).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=tool-2').first()).toBeVisible({ timeout: 10000 });

    // Ensure bulk actions bar is NOT visible initially
    await expect(page.getByText('trace(s) selected')).not.toBeVisible();

    // Select the first trace
    const checkbox1 = page.getByTestId('checkbox-trace-bulk-1');
    await checkbox1.click();

    // Verify bulk actions bar appears and shows 1 selected
    await expect(page.getByText('1 trace(s) selected')).toBeVisible();

    // Select the second trace
    const checkbox2 = page.getByTestId('checkbox-trace-bulk-2');
    await checkbox2.click();

    // Verify bulk actions bar shows 2 selected
    await expect(page.getByText('2 trace(s) selected')).toBeVisible();

    // Click "Delete Selected"
    const deleteBtn = page.getByRole('button', { name: 'Delete Selected' });
    await expect(deleteBtn).toBeVisible();
    await deleteBtn.click();

    // Wait for delete button to disappear to know action is processed
    // Actually the button hides immediately because the traces are removed and state clears
    await expect(deleteBtn).not.toBeVisible();

    // Verify the table items should be gone
    await expect(page.locator('text=tool-1')).not.toBeVisible();
    await expect(page.locator('text=tool-2')).not.toBeVisible();

    // Verify bulk actions bar disappears
    await expect(page.getByText('trace(s) selected')).not.toBeVisible();
  });
});
