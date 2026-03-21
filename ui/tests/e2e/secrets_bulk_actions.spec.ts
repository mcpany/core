/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Secrets Manager Bulk Actions', () => {
    let createdSecrets: string[] = [];

    test.beforeAll(async ({ request }) => {
        // Ensure a clean slate
        await seedGlobalState(request);

        // Seed 3 secrets into the real database using the backend API
        for (let i = 1; i <= 3; i++) {
            const response = await request.post('/api/v1/secrets', {
                data: {
                    name: `e2e-bulk-secret-${i}`,
                    key: `TEST_KEY_${i}`,
                    value: `test-value-${i}`,
                    provider: 'custom'
                }
            });
            expect(response.ok()).toBeTruthy();
            const data = await response.json();
            createdSecrets.push(data.id || data.name);
        }
    });

    test.afterAll(async ({ request }) => {
        // Cleanup any remaining secrets
        for (const id of createdSecrets) {
            try {
                await request.delete(`/api/v1/secrets/${id}`);
            } catch (_e) {
                // Ignore if already deleted
            }
        }
    });

    test('should allow selecting multiple secrets and deleting them in bulk', async ({ page }) => {
        await page.goto('/secrets');

        // Wait for secrets to load (using real data from backend)
        await expect(page.getByText('e2e-bulk-secret-1')).toBeVisible({ timeout: 10000 });

        // 1. Verify "Select all" checkbox
        const selectAllCheckbox = page.getByRole('checkbox', { name: 'Select all secrets' });
        await expect(selectAllCheckbox).toBeVisible();

        // 2. Select individual secrets
        const check1 = page.getByRole('checkbox', { name: 'Select e2e-bulk-secret-1' });
        const check2 = page.getByRole('checkbox', { name: 'Select e2e-bulk-secret-2' });

        await check1.check();
        await check2.check();

        // 3. Verify floating bulk actions bar appears
        const bulkDeleteButton = page.getByRole('button', { name: 'Bulk Delete' });
        await expect(bulkDeleteButton).toBeVisible();
        await expect(page.getByText('2 selected')).toBeVisible();

        // 4. Click Bulk Delete
        await bulkDeleteButton.click();

        // 5. Verify success toast and removal
        await expect(page.getByText('2 secret(s) deleted.')).toBeVisible();

        await expect(page.getByText('e2e-bulk-secret-1')).not.toBeVisible();
        await expect(page.getByText('e2e-bulk-secret-2')).not.toBeVisible();

        // The third secret should still be there
        await expect(page.getByText('e2e-bulk-secret-3')).toBeVisible();
    });
});
