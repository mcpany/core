/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Dashboard Persistence', () => {
  test('persists widget visibility across reloads', async ({ page }) => {
    // 1. Go to Dashboard
    await page.goto('/');

    // 2. Wait for widgets to load (look for a common widget)
    const widgetTitle = 'Recent Activity';
    await expect(page.locator(`text=${widgetTitle}`).first()).toBeVisible({ timeout: 10000 });

    // 3. Open Layout Popover
    await page.getByRole('button', { name: 'Layout' }).click();

    // 4. Uncheck the widget to hide it
    // The checkbox id is `show-{instanceId}`, but we can find it by label
    const checkboxLabel = page.locator('label', { hasText: widgetTitle });
    await expect(checkboxLabel).toBeVisible();
    await checkboxLabel.click();

    // Close the popover to ensure we aren't matching the label inside it
    await page.keyboard.press('Escape');

    // 5. Verify widget is gone (we need to be specific to ensure we aren't matching other text)
    // The widget usually has a header or card with the title.
    // If the widget is hidden, the text "Recent Activity" should NOT be visible on the page (except maybe in the closed popover if DOM persists?)
    // Playwright `toBeHidden` checks visibility.
    await expect(page.locator(`text=${widgetTitle}`)).toBeHidden();

    // 6. Wait for debounce (1s) + buffer
    await page.waitForTimeout(2000);

    // 7. Reload page
    await page.reload();

    // 8. Verify widget is STILL gone
    await expect(page.locator(`text=${widgetTitle}`)).toBeHidden({ timeout: 10000 });

    // 9. Cleanup: Restore visibility
    await page.getByRole('button', { name: 'Layout' }).click();
    await checkboxLabel.click();
    await page.keyboard.press('Escape');
    await expect(page.locator(`text=${widgetTitle}`)).toBeVisible();

    // Take a verification screenshot
    await page.screenshot({ path: 'verification/dashboard_persistence_verified.png' });

    // Wait for debounce to save cleanup
    await page.waitForTimeout(1500);
  });
});
