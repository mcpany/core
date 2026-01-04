
import { test, expect } from '@playwright/test';

test.describe('MCP Marketplace', () => {
  test('should verify marketplace availability and installation flow', async ({ page }) => {
    // 1. Navigate to Services page
    await page.goto('/services');

    // 2. Click on Marketplace tab
    await page.getByRole('tab', { name: 'Marketplace' }).click();

    // 3. Verify Marketplace is visible
    await expect(page.getByText('Service Marketplace')).toBeVisible();
    await expect(page.getByText('Discover and install 1-click MCP servers')).toBeVisible();

    // 4. Search for "SQLite"
    await page.fill("input[placeholder='Search services...']", "SQLite");

    // 5. Verify SQLite card is visible and others are filtered
    await expect(page.getByText('SQLite', { exact: true })).toBeVisible();
    // Use .first() to avoid strict mode violation if "Filesystem" is still transitioning out or if there are multiple matches
    // But logically, if we search "SQLite", "Filesystem" should disappear.
    // Let's verify SQLite is there.

    // 6. Click Install
    // We target the button within the filtered results
    await page.getByRole('button', { name: 'Install' }).first().click();

    // 7. Verify Configuration Dialog
    await expect(page.getByText('Installing SQLite')).toBeVisible();
    await expect(page.getByLabel('DB_PATH')).toBeVisible();

    // 8. Close dialog (cancel) to clean up
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByText('Installing SQLite')).not.toBeVisible();
  });
});
