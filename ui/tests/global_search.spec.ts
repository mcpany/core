/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Global Search', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/');
    });

    test('should open command palette with shortcut', async ({ page }) => {
        // Press Cmd+K
        await page.keyboard.press('Meta+k');
        await expect(page.getByPlaceholder('Type a command or search...')).toBeVisible();
    });

    test('should open command palette with click', async ({ page }) => {
        await page.getByText('Search...').first().click();
        await expect(page.getByPlaceholder('Type a command or search...')).toBeVisible();
    });

    test('should navigate to Services', async ({ page }) => {
        await page.keyboard.press('Meta+k');
        await page.getByPlaceholder('Type a command or search...').fill('Services');
        // Click the first result that matches "Services" in the group "Suggestions"
        await page.getByRole('option', { name: 'Services' }).first().click();
        await expect(page).toHaveURL('/services');
    });

    test('should switch theme', async ({ page }) => {
        await page.keyboard.press('Meta+k');
        await page.getByPlaceholder('Type a command or search...').fill('Light');
        await page.getByRole('option', { name: 'Light' }).click();

        // Check if the html element has the class 'light' or data-theme='light' (depending on next-themes config)
        // next-themes usually adds class='light' or 'dark' to html
        await expect(page.locator('html')).toHaveClass(/light/);
    });
});
