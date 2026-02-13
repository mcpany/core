/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, APIRequestContext } from '@playwright/test';
import path from 'path';
import { seedUser, cleanupUser } from './e2e/test-data';

const DATE = new Date().toISOString().split('T')[0];
// Use test-results/artifacts which is writable in CI
const AUDIT_DIR = path.join(process.cwd(), `test-results/artifacts/audit/ui/${DATE}`);

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const HEADERS = { 'X-API-Key': API_KEY };

const DIFF_SVC_ID = "svc_diff_payment";
const DIFF_SVC_NAME = "Diff Payment Gateway";

const seedDiffService = async (requestContext: APIRequestContext) => {
    // Delete if exists (defensive)
    try {
        await requestContext.delete(`/api/v1/services/${DIFF_SVC_NAME}`, { headers: HEADERS });
    } catch (e) {}
    try {
        await requestContext.delete(`/api/v1/services/${DIFF_SVC_NAME} Updated`, { headers: HEADERS });
    } catch (e) {}

    const service = {
        id: DIFF_SVC_ID,
        name: DIFF_SVC_NAME,
        version: "v1.2.0",
        http_service: {
            address: "https://stripe.com",
            tools: [
                { name: "process_payment_diff", description: "Process a payment diff" }
            ]
        }
    };

    try {
        await requestContext.post('/api/v1/services', { data: service, headers: HEADERS });
    } catch (e) {
        console.log(`Failed to seed diff service: ${e}`);
    }
};

const cleanupDiffService = async (requestContext: APIRequestContext) => {
    try {
        await requestContext.delete(`/api/v1/services/${DIFF_SVC_NAME}`, { headers: HEADERS });
        await requestContext.delete(`/api/v1/services/${DIFF_SVC_NAME} Updated`, { headers: HEADERS });
    } catch (e) {
        console.log(`Failed to cleanup diff service: ${e}`);
    }
};

test.describe('Service Config Diff', () => {
    test.beforeEach(async ({ request, page }) => {
        await seedDiffService(request);
        await seedUser(request, "diff-admin");

        // Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'diff-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupDiffService(request);
        await cleanupUser(request, "diff-admin");
    });

    test('Shows diff when editing service config', async ({ page }) => {
        const serviceName = DIFF_SVC_NAME;
        const newServiceName = `${DIFF_SVC_NAME} Updated`;

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
