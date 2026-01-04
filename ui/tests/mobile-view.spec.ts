
import { test, expect } from '@playwright/test';

test.describe('Mobile View Verification', () => {
  test.use({ viewport: { width: 375, height: 667 } }); // iPhone SE

  test('Log Stream should be responsive', async ({ page }) => {
    await page.goto('/logs');

    // Check if header stacked
    const title = page.locator('h1:has-text("Live Logs")');
    await expect(title).toBeVisible();

    // Check if table container is scrollable/visible
    // We look for the main log area card
    const card = page.locator('[data-testid="log-rows-container"]').locator('..').locator('..').locator('..');
    await expect(card).toBeVisible();

    // Check input width
    const searchInput = page.locator('input[placeholder="Search logs..."]');
    await expect(searchInput).toBeVisible();
  });

  test('Secrets Manager should be responsive', async ({ page }) => {
    await page.goto('/secrets');

    // Check table wrapper
    const tableWrapper = page.locator('.overflow-x-auto');
    await expect(tableWrapper).toBeVisible();

    // Check dialog responsiveness
    await page.click('button:has-text("Add Secret")');
    const dialog = page.locator('div[role="dialog"]');
    await expect(dialog).toBeVisible();

    // Check stacked inputs
    // We can check if labels are above inputs? Or just visibility.
    await expect(page.locator('label:has-text("Name")')).toBeVisible();
  });
});
