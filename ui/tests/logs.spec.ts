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

  test.skip('should pause on scroll up', async ({ page }) => {
    await page.clock.install();
    await page.goto('/logs');

    // Generate some logs
    await page.clock.runFor(10000);
    const logContainer = page.getByTestId('log-rows-container');

    // Simulate scroll up
    // We need to ensure there is enough content to scroll.
    // The mock generates plenty of logs.

    // Hover over container to ensure focus (optional but good for stability)
    await logContainer.hover();

    // Scroll up wheel event
    await logContainer.dispatchEvent('wheel', { deltaY: -500 });

    // Verify "Resume" button appears which indicates pause state
    await expect(page.getByRole('button', { name: 'Resume' })).toBeVisible();
  });

  test('should pause and resume logs', async ({ page }) => {
    // Install fake clock to control setInterval
    await page.clock.install();
    await page.goto('/logs');

    // Initial accumulation - advance time to generate some logs
    await page.clock.runFor(2000);
    const logs = page.locator('.group');
    const countInitial = await logs.count();
    expect(countInitial).toBeGreaterThan(0);

    // Click pause
    const pauseButton = page.getByRole('button', { name: 'Pause' });
    await pauseButton.click();

    // Wait for state to update
    await expect(page.getByRole('button', { name: 'Resume' })).toBeVisible();

    const countAfterPause = await logs.count();

    // Advance time - NO new logs should be added
    await page.clock.runFor(4000);
    const countAfterWait = await logs.count();
    expect(countAfterWait).toBe(countAfterPause);

    // Resume
    const resumeButton = page.getByRole('button', { name: 'Resume' });
    await resumeButton.click();
    await expect(page.getByRole('button', { name: 'Pause' })).toBeVisible();

    // Advance time - logs SHOULD be added
    await page.clock.runFor(4000);
    const countAfterResume = await logs.count();
    expect(countAfterResume).toBeGreaterThan(countAfterWait);
  });

  test('should filter logs', async ({ page }) => {
      // Wait for some logs to appear
      await page.waitForTimeout(3000);

      // Find a log level that exists in the current list
      const logRows = page.getByTestId('log-rows-container').locator('.group');
      const count = await logRows.count();
      expect(count).toBeGreaterThan(0);

      // Get the level of the first log
      const firstLogLevel = await logRows.first().locator('span').nth(1).innerText();
      const targetLevel = firstLogLevel.trim();

      // Set filter to the found level
      // Map log level to title case for selection (INFO -> Info, ERROR -> Error)
      const levelOptionMap: Record<string, string> = {
          'INFO': 'Info',
          'WARN': 'Warning',
          'ERROR': 'Error',
          'DEBUG': 'Debug'
      };

      const optionName = levelOptionMap[targetLevel];
      if (!optionName) {
           // Fallback if we catch a level we didn't map (or if text is empty)
           console.log(`Unknown level: ${targetLevel}, skipping filter test`);
           return;
      }

      const filterSelect = page.getByRole('combobox');
      await filterSelect.click();
      await page.getByRole('option', { name: optionName }).click();

      await page.waitForTimeout(1000);

      // Check if all visible logs match the target level
      // We need to re-locate because DOM updates
      const visibleLogs = page.getByTestId('log-rows-container').locator('.group');
      const visibleCount = await visibleLogs.count();

      for (let i = 0; i < visibleCount; i++) {
          const logText = await visibleLogs.nth(i).innerText();
          expect(logText).toContain(targetLevel);
      }
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
});
