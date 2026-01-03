/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Secrets Manager', () => {
    test('should allow creating and deleting a secret', async ({ page }) => {
        // Go to settings page
        await page.goto('http://localhost:9002/settings');

        // Click on Secrets tab
        await page.click('text=Secrets & Keys');

        // Check if "Add Secret" button exists
        await expect(page.locator('button:has-text("Add Secret")')).toBeVisible();

        // Open add dialog
        await page.click('button:has-text("Add Secret")');

        // Fill form
        await page.fill('input[placeholder="e.g. Production OpenAI Key"]', 'E2E Test Secret');
        await page.fill('input[placeholder="e.g. OPENAI_API_KEY"]', 'E2E_KEY');
        await page.fill('input[placeholder="sk-..."]', 'sk-test-value-123');

        // Save
        await page.click('button:has-text("Save Secret")');

        // Verify toast or list update
        await expect(page.locator('text=E2E Test Secret')).toBeVisible();
        await expect(page.locator('text=E2E_KEY')).toBeVisible();

        // Verify delete
        const deleteButton = page.locator('button[aria-label="Delete secret"]').last();
        await deleteButton.click();

        // Verify it's gone
        await expect(page.locator('text=E2E Test Secret')).not.toBeVisible();
    });
});
