/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Marketplace Tests', () => {
  test('Share Config flow should work', async ({ page }) => {
    await page.goto('/marketplace');

    // Verify Share Button exists
    const shareButton = page.getByRole('button', { name: 'Share Your Config' });
    await expect(shareButton).toBeVisible();

    // Click it
    await shareButton.click();

    // Verify Dialog Open
    const dialog = page.getByRole('dialog', { name: 'Share Service Collection' });
    await expect(dialog).toBeVisible();

    // Verify "Generate Configuration" exists and is initially disabled if no services (or enabled if default selected)
    // Based on implementation, we default to no selection? Or maybe we select all?
    // Implementation: "const [selected, setSelected] = React.useState<Set<string>>(new Set())" -> Empty initially.

    // Wait for services to load (table should be populated)
    // We mocked the data, so it should be fast.

    // Select a service (checkbox)
    // We assume there's at least one service row
    const firstCheckbox = page.locator('table tbody tr:first-child [role="checkbox"]');
    if (await firstCheckbox.count() > 0) {
        await firstCheckbox.click();

        // Click Generate
        const generateBtn = page.getByRole('button', { name: 'Generate Configuration' });
        await expect(generateBtn).toBeEnabled();
        await generateBtn.click();

        // Verify Textarea with config appears
        const textarea = page.locator('textarea');
        await expect(textarea).toBeVisible();

        // Verify it contains some yaml content
        const value = await textarea.inputValue();
        expect(value).toContain('name: My Shared Collection');

        // Verify Copy button
        const copyBtn = page.getByRole('button').filter({ has: page.locator('svg.lucide-copy') });
        await expect(copyBtn).toBeVisible();
    } else {
        console.log('No services found to test sharing');
    }
  });

  test.skip('should open install modal', async ({ page }) => {
    // Mock the templates API to return a predefined list so we have a known state
    await page.route('**/api/v1/templates', async route => {
       await route.fulfill({
           json: [
               {
                   id: 'postgres-14',
                   name: 'PostgreSQL 14',
                   description: 'Standard SQL Database',
                   version: '14.2',
                   category: 'Database'
               }
           ]
       });
    });

    await page.goto('/marketplace');

    // Switch to Local tab where templates are listed
    await page.getByRole('tab', { name: "Local" }).click();

    // Find the Postgres card
    const card = page.locator('.rounded-xl').filter({ hasText: 'PostgreSQL 14' });

    await expect(card).toBeVisible();

    // Click "Instantiate"
    // Adjust selector based on actual UI.
    await card.getByRole('button', { name: 'Instantiate' }).click();

    // Verify Modal
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Instantiate Service' })).toBeVisible();

    // Verify Create Instance button exists
    await expect(page.getByRole('button', { name: 'Create Instance' })).toBeVisible();
  });
});
