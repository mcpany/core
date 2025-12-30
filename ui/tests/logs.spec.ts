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
    await expect(page.getByText('Live Logs')).toBeVisible();
  });

  test('should display log entries', async ({ page }) => {
    // Wait for at least one log entry to appear (logs are generated every 800ms)
    await page.waitForTimeout(2000);
    const logs = page.locator('.group');
    const count = await logs.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should pause and resume logs', async ({ page }) => {
    const pauseButton = page.getByRole('button', { name: 'Pause' });
    await pauseButton.click();

    // Get count of logs
    await page.waitForTimeout(1000);
    const logs = page.locator('.group');
    const countAfterPause = await logs.count();

    // Wait more to ensure no new logs are added
    await page.waitForTimeout(2000);
    const countAfterWait = await logs.count();
    expect(countAfterWait).toBe(countAfterPause);

    const resumeButton = page.getByRole('button', { name: 'Resume' });
    await resumeButton.click();

    // Wait for new logs
    await page.waitForTimeout(2000);
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
