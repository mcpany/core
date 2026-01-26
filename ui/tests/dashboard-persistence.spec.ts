/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Dashboard Persistence', () => {
  test.beforeEach(async ({ page }) => {
    // Abort backend calls to ensure we fail fast and fall back to localStorage
    // This simulates "Backend Down" scenario where persistence should still work locally.
    await page.route('**/api/v1/users/*', route => route.abort());
    await page.route('**/api/v1/dashboard/*', route => route.abort());
    await page.route('**/debug/entries', route => route.abort());
  });

  test('should persist layout changes after reload', async ({ page }) => {
    // 1. Go to Dashboard
    await page.goto('/');

    // Wait for "Recent Activity" widget to be visible
    const widgetTitle = 'Recent Activity';
    const widgetLocator = page.locator('.group\\/widget', { hasText: widgetTitle });

    // We expect it to be there initially
    await expect(widgetLocator).toBeVisible({ timeout: 10000 });

    // 2. Remove the widget
    // Hover to trigger any hover effects (though dropdown should be clickable without hover if implemented right,
    // but the code says opacity-0 group-hover:opacity-100)
    await widgetLocator.hover();

    // Click the MoreHorizontal icon (dropdown trigger)
    // It's inside the widget container.
    // Use force: true because of the hover opacity transition which might cause Playwright to think it's not actionable yet
    // Target the dropdown trigger by aria attribute
    const trigger = widgetLocator.locator('[aria-haspopup="menu"]');
    await trigger.click({ force: true });

    // Click "Remove" in the dropdown
    await page.getByRole('menuitem', { name: 'Remove' }).click();

    // Verify it is gone
    await expect(widgetLocator).not.toBeVisible();

    // 3. Reload the page
    await page.reload();

    // 4. Verify persistence
    // "Metrics Overview" widget should still be there.
    // Since backend is mocked/aborted, it might show "Loading dashboard metrics..." or actual metrics if mocked.
    // But checking for the "Recent Activity" absence is the key.

    // Verify "Recent Activity" is NOT there
    await expect(page.locator('.group\\/widget', { hasText: widgetTitle })).not.toBeVisible();

    // Verify at least one other widget IS there (e.g. Metrics Overview showing loading state)
    // "Metrics Overview" component doesn't render its own title, but it renders "Loading dashboard metrics..." if empty
    // OR "Request Rate" etc if loaded.
    // Let's check for the loading state or any other widget title that defaults to present
    // Top Tools is also default.
    // Let's check for "Loading dashboard metrics..." OR "Request Rate"
    const loadingOrMetrics = page.locator('text=Loading dashboard metrics...').or(page.locator('text=Request Rate'));
    await expect(loadingOrMetrics).toBeVisible({ timeout: 10000 });
  });
});
