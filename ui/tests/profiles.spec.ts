/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Profile Management', () => {
    const profileName = `e2e-test-profile-${Date.now()}`;

    test.beforeEach(async ({ request }) => {
        // Ensure clean state
        await request.delete(`/api/v1/profiles/${profileName}`).catch(() => {});
    });

    test.afterEach(async ({ request }) => {
        // Clean up
        await request.delete(`/api/v1/profiles/${profileName}`).catch(() => {});
    });

    test('should allow creating and listing a profile', async ({ page, request }) => {
        // 1. Visit Settings
        await page.goto('/settings');

        // 2. Click New Profile
        await page.getByRole('button', { name: 'New Profile' }).click();

        // 3. Fill Form
        await page.getByLabel('Profile Name').fill(profileName);
        await page.getByLabel('Selector Tags').fill('e2e, test');
        await page.getByLabel('Required Roles').fill('admin');

        // 4. Save
        await page.getByRole('button', { name: 'Create Profile' }).click();

        // 5. Verify it appears
        await expect(page.getByText(profileName)).toBeVisible();
        await expect(page.getByText('e2e')).toBeVisible();

        // 6. Verify backend state (Real Data Law)
        const res = await request.get('/api/v1/profiles');
        const profiles = await res.json();
        const found = profiles.find((p: { name: string, selector: { tags: string[] } }) => p.name === profileName);
        expect(found).toBeTruthy();
        expect(found.selector.tags).toContain('e2e');
    });

    test('should allow deleting a profile', async ({ page, request }) => {
        // Seed
        await request.post('/api/v1/profiles', {
            data: { name: profileName, selector: { tags: ['to-be-deleted'] } }
        });

        await page.goto('/settings');
        await expect(page.getByText(profileName)).toBeVisible();

        // Open menu
        const card = page.locator('.group', { hasText: profileName });
        await card.hover();
        await card.getByRole('button', { name: 'Open menu' }).click();

        // Handle confirmation dialog
        page.on('dialog', dialog => dialog.accept());

        // Click Delete
        await page.getByRole('menuitem', { name: 'Delete' }).click();

        // Verify gone
        await expect(page.getByText(profileName)).not.toBeVisible();

        // Verify backend state
        const res = await request.get('/api/v1/profiles');
        const profiles = await res.json();
        const found = profiles.find((p: { name: string }) => p.name === profileName);
        expect(found).toBeFalsy();
    });
});
