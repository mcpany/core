/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Users Bulk Actions', () => {
    const usersToCreate = ['e2e-user-1', 'e2e-user-2', 'e2e-user-3'];

    test.beforeEach(async ({ request }) => {
        // Seed database with real data via API
        for (const uid of usersToCreate) {
            await request.post('/api/v1/users', {
                data: {
                    user: {
                        id: uid,
                        roles: ['viewer']
                    }
                }
            });
        }
    });

    test.afterEach(async ({ request }) => {
        // Cleanup remaining users
        for (const uid of usersToCreate) {
            await request.delete(`/api/v1/users/${uid}`);
        }
    });

    test('should allow selecting multiple users and performing bulk delete', async ({ page, request }) => {
        // Navigate to the users page
        await page.goto('/users');

        // Use a more resilient selector for the main heading since there could be multiple headings during load
        await expect(page.getByRole('heading', { name: 'Users', exact: true }).first()).toBeVisible({ timeout: 15000 });

        // Wait for the table to load the users
        for (const uid of usersToCreate) {
            await expect(page.locator(`tr[data-testid="user-row-${uid}"]`)).toBeVisible({ timeout: 10000 });
        }

        // Select the first two users
        await page.locator(`tr[data-testid="user-row-${usersToCreate[0]}"]`).getByRole('checkbox').click({ force: true });
        await page.locator(`tr[data-testid="user-row-${usersToCreate[1]}"]`).getByRole('checkbox').click({ force: true });

        // The floating action bar should appear
        const floatingBar = page.getByText('users selected');
        await expect(floatingBar).toBeVisible();
        await expect(page.getByText('2', { exact: true })).toBeVisible();

        // Click "Delete Selected"
        // Handle the confirm dialog automatically
        page.on('dialog', dialog => dialog.accept());
        await page.getByRole('button', { name: 'Delete Selected' }).click();

        // Verify toast message appears indicating success
        await expect(page.getByText('Successfully deleted 2 user(s).')).toBeVisible();

        // Verify the deleted users are no longer in the list visually
        await expect(page.locator(`tr[data-testid="user-row-${usersToCreate[0]}"]`)).toBeHidden();
        await expect(page.locator(`tr[data-testid="user-row-${usersToCreate[1]}"]`)).toBeHidden();

        // Verify the 3rd user is still present visually
        await expect(page.locator(`tr[data-testid="user-row-${usersToCreate[2]}"]`)).toBeVisible();

        // Verify backend state changed: First two users should be deleted
        const res1 = await request.get(`/api/v1/users/${usersToCreate[0]}`);
        expect(res1.status()).toBe(404);

        const res2 = await request.get(`/api/v1/users/${usersToCreate[1]}`);
        expect(res2.status()).toBe(404);

        // Third user should still exist
        const res3 = await request.get(`/api/v1/users/${usersToCreate[2]}`);
        expect(res3.status()).toBe(200);
    });
});
