/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';

const DATE = new Date().toISOString().split('T')[0];
// Use test-results/artifacts which is writable in CI
const AUDIT_DIR = path.join(process.cwd(), `test-results/artifacts/audit/ui/${DATE}`);

test.describe('Service Config Diff', () => {
    test.beforeEach(async ({ request, page }) => {
        // Create a dedicated service for this test to avoid conflicts with other tests running in parallel
        await request.post('/api/v1/services', {
            data: {
                id: "svc_diff_test",
                name: "Diff Test Service",
                version: "v1.0.0",
                http_service: {
                    address: "https://example.com",
                    tools: [{ name: "diff_tool", description: "Tool for diff testing" }]
                }
            },
            headers: { 'X-API-Key': 'test-token' }
        });

        await seedUser(request, "diff-admin");

        // Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'diff-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        // Clean up the dedicated service
        await request.delete('/api/v1/services/Diff Test Service', { headers: { 'X-API-Key': 'test-token' } });
        // Also try to clean up the renamed version if the test passed
        await request.delete('/api/v1/services/Diff Test Service Updated', { headers: { 'X-API-Key': 'test-token' } });

        await cleanupUser(request, "diff-admin");
    });

    test('Shows diff when editing service config', async ({ page }) => {
        const serviceName = "Diff Test Service";
        const newServiceName = "Diff Test Service Updated";

        // Go to Upstream Services page to find the service
        await page.goto('/upstream-services');
        await expect(page.getByText(serviceName, { exact: true })).toBeVisible();

        // Click Edit (via Dropdown)
        // Locate the row by text (use filter to ensure exact match on name cell if possible, or just first)
        const row = page.locator('tr').filter({ hasText: serviceName }).first();
        // Click the dropdown menu trigger
        await row.getByRole('button', { name: 'Open menu' }).click();
        // Click Edit option
        await page.getByRole('menuitem', { name: 'Edit' }).click();

        // Wait for sheet (Edit Service)
        await expect(page.getByText('Edit Service')).toBeVisible();

        // Change name
        await page.getByLabel('Service Name').fill(newServiceName);
        await page.getByLabel('Service Name').blur();

        // Click "Changes" tab
        await page.getByRole('tab', { name: 'Changes' }).click();

        // Verify Diff Viewer is present
        await expect(page.getByText('Configuration Changes')).toBeVisible();

        // Verify Diff Editor content exists (Monaco diff editor class)
        await expect(page.locator('.monaco-diff-editor')).toBeVisible();

        // Take screenshot
        try {
            await page.screenshot({ path: path.join(AUDIT_DIR, 'service_config_diff_real.png') });
        } catch (e) {
            console.error("Failed to take screenshot:", e);
        }

        // Save
        await page.getByRole('button', { name: 'Save Changes' }).click();

        // Verify success toast or closed sheet
        // Use exact match to avoid matching description or other elements
        await expect(page.getByText('Service Updated', { exact: true }).first()).toBeVisible();

        // Verify list updates
        // Scope to the table to avoid matching the Monaco editor if it's still in the DOM
        await expect(page.locator('table').getByText(newServiceName, { exact: true })).toBeVisible();
    });
});
