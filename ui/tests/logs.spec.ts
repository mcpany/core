/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Logs Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/logs');
  });

  test('should display logs title', async ({ page }) => {
    // Verify page loaded
    await expect(page).toHaveTitle(/MCPAny/);
    // Use more specific selector with longer timeout
    await expect(page.getByRole('heading', { name: 'Live Logs' })).toBeVisible({ timeout: 30000 });
  });

  test('should display log entries', async ({ page }) => {
    // Wait for at least one log entry to appear (logs are generated every 800ms)
    await page.waitForTimeout(2000);
    const logs = page.locator('.group');
    const count = await logs.count();
    expect(count).toBeGreaterThan(0);
  });






  test('should pause and resume logs', async ({ page }) => {
    // Wait for initial logs
    await page.waitForTimeout(2000);

    // Check initial state (Resume button hidden or Pause visible)
    // Actually the button toggles text. Default is "Pause" (meaning click to pause).
    const pauseButton = page.getByRole('button', { name: 'Pause' });
    await expect(pauseButton).toBeVisible();
    await pauseButton.click();

    // Verify it changed to "Resume"
    await expect(page.getByRole('button', { name: 'Resume' })).toBeVisible();

    // Click Resume
    await page.getByRole('button', { name: 'Resume' }).click();
    await expect(page.getByRole('button', { name: 'Pause' })).toBeVisible();
  });


  test('should clear logs', async ({ page }) => {
      await page.waitForTimeout(4000);
      const clearButton = page.getByRole('button', { name: 'Clear' });
      await clearButton.click();

      await clearButton.click();

      const logArea = page.getByTestId('log-rows-container');
      const logRows = logArea.locator('.group');

      // Wait for logs to be cleared (count should drop)
      await expect(async () => {
        const count = await logRows.count();
        expect(count).toBeLessThan(3);
      }).toPass({ timeout: 2000 });
  });

  test('should display and expand JSON logs (Real Data)', async ({ page, request }) => {
    const jsonMessage = { foo: "bar", nested: { val: 123 } };
    const seedLog = {
      level: "INFO",
      message: JSON.stringify(jsonMessage),
      source: "e2e-json-test"
    };

    // Seed log
    await request.post('/api/v1/debug/seed_logs', { data: [seedLog] });

    // Wait for the log to appear (live update)
    const row = page.locator('.group', { hasText: JSON.stringify(jsonMessage) });
    await expect(row).toBeVisible({ timeout: 15000 });

    // Check for the Expand JSON button
    const expandButton = row.getByLabel("Expand JSON");
    await expect(expandButton).toBeVisible();

    // Click it
    await expandButton.click();

    // Verify it expanded (Collapse button visible)
    await expect(row.getByLabel("Collapse JSON")).toBeVisible();
  });

  test('should display persistent logs after reload', async ({ page, request }) => {
    // Seed a log via API
    const uniqueMsg = "Persistent Log Test " + Date.now();
    const seedLog = {
      level: "WARN",
      message: uniqueMsg,
      source: "e2e-seed"
    };

    await request.post('/api/v1/debug/seed_logs', {
      data: [seedLog]
    });

    // Wait for it to appear live
    await expect(page.getByText(uniqueMsg)).toBeVisible({ timeout: 10000 });

    // Reload the page
    await page.reload();

    // Check if log is still there (fetched from backend persistence)
    await expect(page.getByText(uniqueMsg)).toBeVisible({ timeout: 10000 });
  });
});
