/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Credentials Bulk Actions', () => {
    test.beforeEach(async ({ request }) => {
        // Seed multiple credentials
        const mockCredentials = [
            {
                id: "cred-1",
                name: "Test Credential 1",
                authentication: {
                    apiKey: {
                        in: 0, // Header
                        paramName: "Authorization"
                    }
                }
            },
            {
                id: "cred-2",
                name: "Test Credential 2",
                authentication: {
                    bearerToken: {}
                }
            },
            {
                id: "cred-3",
                name: "Test Credential 3",
                authentication: {
                    basicAuth: {
                        username: "admin"
                    }
                }
            }
        ];

        const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
        try {
            await request.post('/api/v1/debug/seed', {
                data: JSON.stringify({ credentials: mockCredentials }),
                headers: {
                    'X-API-Key': API_KEY,
                    'Content-Type': 'application/json'
                }
            });
        } catch (e) {
            console.error('Failed to seed credentials', e);
        }
    });

    test('should select credentials and perform bulk delete', async ({ page }) => {
        // Log in to ensure auth headers are correct
        await page.goto('/login');
        await page.getByPlaceholder('Username').fill('admin');
        await page.getByPlaceholder('Password').fill('admin');
        await page.getByRole('button', { name: 'Sign in' }).click();

        // Wait for dashboard to load
        await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

        await page.goto('/credentials');

        // Ensure page is loaded
        await expect(page.getByRole('heading', { name: 'Credentials', exact: true })).toBeVisible();

        // Verify seeded credentials are listed
        // Wait for the table to populate
        await expect(page.getByText('Test Credential 1')).toBeVisible({ timeout: 15000 });
        await expect(page.getByText('Test Credential 2')).toBeVisible();
        await expect(page.getByText('Test Credential 3')).toBeVisible();

        // Initially no bulk toolbar
        await expect(page.getByText('selected')).toBeHidden();

        // Select the first two credentials
        await page.getByLabel('Select credential Test Credential 1').check();
        await page.getByLabel('Select credential Test Credential 2').check();

        // Bulk toolbar should appear
        await expect(page.getByText('2 selected')).toBeVisible();

        // Intercept confirm dialog
        page.on('dialog', async dialog => {
            expect(dialog.message()).toContain('2 credentials');
            await dialog.accept();
        });

        // Click delete selected
        await page.getByRole('button', { name: 'Delete Selected' }).click();

        // Ensure credentials were deleted from the UI
        await expect(page.getByText('Test Credential 1')).toBeHidden();
        await expect(page.getByText('Test Credential 2')).toBeHidden();

        // Ensure unselected credential remains
        await expect(page.getByText('Test Credential 3')).toBeVisible();

        // Bulk toolbar should disappear after delete
        await expect(page.getByText('selected')).toBeHidden();
    });
});
