/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('HITL Dashboard Flow', () => {
    test('renders pending approvals and handles actions without MFA', async ({ page }) => {
        // Assume API is already seeded by the backend in api_hitl.go
        await page.goto('/hitl');

        // Wait for the specific tool card
        const card = page.locator('.rounded-xl.border', { hasText: 'aws.terminate_instance' });
        await expect(card).toBeVisible();

        // Click the Approve button
        const approveBtn = card.getByRole('button', { name: 'Approve' });
        await approveBtn.click();

        // Verify it was removed or actioned
        await expect(card).not.toBeVisible();
    });

    test('renders pending approvals and handles actions with MFA', async ({ page }) => {
        await page.goto('/hitl');

        // Wait for the specific tool card
        const card = page.locator('.rounded-xl.border', { hasText: 'database.drop_table' });
        await expect(card).toBeVisible();

        // Click the Approve button
        const approveBtn = card.getByRole('button', { name: 'Approve' });
        await approveBtn.click();

        // Should open MFA dialog
        const dialog = page.locator('[role="dialog"]', { hasText: 'Multi-Factor Authentication Required' });
        await expect(dialog).toBeVisible();

        // Enter MFA code
        await dialog.getByPlaceholder('MFA Code').fill('123456');

        // Submit
        await dialog.getByRole('button', { name: 'Verify & Approve' }).click();

        // Verify dialog closes and card disappears
        await expect(dialog).not.toBeVisible();
        await expect(card).not.toBeVisible();
    });
});
