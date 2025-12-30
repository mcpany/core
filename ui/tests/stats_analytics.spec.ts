
import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

test('verify stats page', async ({ page }) => {
  // Go to the stats page
  await page.goto('/stats');

  // Wait for the dashboard to load
  await expect(page.getByText('Analytics & Stats')).toBeVisible();

  // Check for key elements
  await expect(page.getByText('Total Requests')).toBeVisible();
  await expect(page.getByText('Avg Latency')).toBeVisible();
  await expect(page.getByText('Error Rate')).toBeVisible();

  // Check tabs
  await expect(page.getByText('Overview')).toBeVisible();
  await expect(page.getByText('Performance')).toBeVisible();
  await expect(page.getByText('Errors')).toBeVisible();

  // Wait a bit for charts to animate (if any)
  await page.waitForTimeout(2000);

  // Take a screenshot
  // The python script saved it to .audit/ui/2025-12-30/stats_analytics.png
  // .audit is in the repo root.
  // The test file is in ui/tests/stats_analytics.spec.ts
  // So repo root is ../../
  const screenshotDir = path.resolve(__dirname, '../../.audit/ui/2025-12-30');
  if (!fs.existsSync(screenshotDir)) {
      fs.mkdirSync(screenshotDir, { recursive: true });
  }
  await page.screenshot({ path: path.join(screenshotDir, 'stats_analytics.png'), fullPage: true });
});
