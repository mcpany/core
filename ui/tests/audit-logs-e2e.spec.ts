/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedGlobalState, seedTraffic } from './e2e/test-data';

test.describe('Audit Logs Viewer E2E', () => {
    test.beforeEach(async ({ request, page }) => {
        // Seed database state which includes an audit log via executing a tool
        await seedGlobalState(request);
        await seedTraffic(request);

        // Login before each test
        await page.goto('/login');
        await page.waitForLoadState('networkidle');

        await page.fill('input[name="username"]', 'e2e-admin-core');
        await page.fill('input[name="password"]', 'password');
        await Promise.all([
          page.waitForURL('/', { timeout: 30000 }),
          page.click('button[type="submit"]', { force: true })
        ]);
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test('should display seeded audit logs and format json payloads correctly', async ({ page }) => {
        // Navigate to the audit logs page
        await page.goto('/audit');
        await expect(page.getByRole('heading', { name: 'Filters' })).toBeVisible();

        // Check if our generated audit log is present in the table.
        // We know we executed `echo_tool`.
        const row = page.locator('tr').filter({ hasText: 'echo_tool' }).first();
        await expect(row).toBeVisible({ timeout: 10000 });

        // Ensure the "Success" badge is visible
        await expect(row.getByText('Success')).toBeVisible();

        // Click "View" to open details
        await row.getByRole('button', { name: 'View' }).click();

        // Wait for dialog
        const dialog = page.getByRole('dialog', { name: 'Audit Log Detail' });
        await expect(dialog).toBeVisible();

        // Ensure the RichResultViewer is rendered inside the arguments and results
        // "Arguments" tab/section
        await expect(dialog.getByText('Arguments')).toBeVisible();
        await expect(dialog.locator('span', { hasText: 'test-audit-log-generation' }).first()).toBeVisible();

        // "Result" tab/section
        await expect(dialog.getByText('Result')).toBeVisible();
        await expect(dialog.locator('span', { hasText: 'echoed_output' }).first()).toBeVisible();

        // Verify the raw JSON dump is no longer present.
        // SyntaxHighlighter rendered `<pre>` blocks with specific syntax highlighting classes.
        // The new UI uses RichResultViewer or JsonView, which renders tabs (Rendered, JSON, Raw Output).
        await expect(dialog.getByRole('tab', { name: 'JSON' }).first()).toBeVisible();
    });
});
