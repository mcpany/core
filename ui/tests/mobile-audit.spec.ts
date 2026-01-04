
import { test, expect } from '@playwright/test';

test.use({ viewport: { width: 375, height: 812 } });

test('Mobile Audit Screenshots', async ({ page }) => {
  // 1. Logs
  await page.goto('/logs');
  await expect(page.locator('h1:has-text("Live Logs")')).toBeVisible();
  // Wait for some logs
  await expect(page.locator('[data-testid="log-rows-container"]')).toBeVisible();
  await page.screenshot({ path: `.audit/ui/2026-01-04/mobile_logs.png`, fullPage: true });

  // 2. Secrets
  await page.goto('/secrets');
  await expect(page.locator('table')).toBeVisible();
  // Open dialog
  await page.click('button:has-text("Add Secret")');
  await expect(page.locator('div[role="dialog"]')).toBeVisible();
  await page.screenshot({ path: `.audit/ui/2026-01-04/mobile_secrets_dialog.png` });
  await page.keyboard.press('Escape');

  // 3. Network
  await page.goto('/network');
  // Wait for graph
  await expect(page.locator('.react-flow__renderer')).toBeVisible();
  await page.screenshot({ path: `.audit/ui/2026-01-04/mobile_network.png` });
});
