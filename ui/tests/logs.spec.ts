
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
      // Set filter to ERROR
      const filterSelect = page.getByText('All Levels'); // Trigger
      await filterSelect.click();
      await page.getByRole('option', { name: 'Error' }).click();

      await page.waitForTimeout(2000);

      // Check if all visible logs are ERROR
      const logLevels = page.locator('.text-red-400'); // ERROR color class
      const allLogs = page.locator('.group');

      // Note: This check is a bit loose because new logs might arrive,
      // but filtered view should only show matching logs.
      // We check if text content of visible logs contains "ERROR"
      // Filter out sidebar items that might match '.group' if reusing classes or simply scope to log area
      const logArea = page.locator('.space-y-1');
      const logRows = logArea.locator('.group');

      const logTexts = await logRows.allInnerTexts();
      for (const text of logTexts) {
          expect(text).toContain('ERROR');
      }
  });

  test('should clear logs', async ({ page }) => {
      await page.waitForTimeout(2000);
      const clearButton = page.getByRole('button', { name: 'Clear' });
      await clearButton.click();

      const logs = page.locator('.group');
      // Might strictly be 0, or 1 if a new log arrived just now.
      const count = await logs.count();
      expect(count).toBeLessThan(3);
  });
});
