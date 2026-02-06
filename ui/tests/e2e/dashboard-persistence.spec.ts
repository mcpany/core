/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Dashboard Persistence', () => {
    test('should persist widget visibility after reload', async ({ page }) => {
        // 1. Visit Dashboard
        await page.goto('/');
        await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

        // 2. Ensure we have widgets. If empty, add default set should have happened.
        // Wait for widgets to appear.
        // The Default Layout has widgets.
        await page.waitForSelector('.group\\/widget', { state: 'visible', timeout: 5000 });

        // 3. Open Layout Popover
        await page.getByRole('button', { name: 'Layout' }).click();
        await expect(page.getByText('Visible Widgets')).toBeVisible();

        // 4. Find the first checkbox
        const firstCheckbox = page.locator('div[role="dialog"] button[role="checkbox"]').first();
        // Note: Radix UI Checkbox is often a button[role="checkbox"], not input[type="checkbox"]
        await expect(firstCheckbox).toBeVisible();

        // Ensure it is checked (Visible) initially
        // If not checked, check it.
        if (await firstCheckbox.getAttribute('aria-checked') === 'false') {
            await firstCheckbox.click();
            // Wait for save
            await page.waitForTimeout(1500);
        }
        expect(await firstCheckbox.getAttribute('aria-checked')).toBe('true');

        // 5. Toggle it OFF (Hide)
        await firstCheckbox.click();
        expect(await firstCheckbox.getAttribute('aria-checked')).toBe('false');

        // Screenshot 1: Hidden
        await page.screenshot({ path: 'test-results/hidden.png' });

        // 6. Wait for debounce save (1s) + buffer
        await page.waitForTimeout(2000);

        // 7. Reload Page
        await page.reload();
        await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

        // 8. Verify Persistence
        await page.getByRole('button', { name: 'Layout' }).click();
        const firstCheckboxAfter = page.locator('div[role="dialog"] button[role="checkbox"]').first();
        await expect(firstCheckboxAfter).toBeVisible();

        expect(await firstCheckboxAfter.getAttribute('aria-checked')).toBe('false');

        // Screenshot 2: Persisted
        await page.screenshot({ path: 'test-results/persisted.png' });
    });
});
