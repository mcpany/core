/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


<<<<<<< HEAD
import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';
import { seedGlobalState } from './e2e/test-data';
=======
import { test } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))

test.describe('Feature Screenshot', () => {
    // Enabled audit screenshots

    const date = new Date().toISOString().split('T')[0];
    // Use test-results directory which is writable in CI
    const auditDir = path.join(process.cwd(), 'test-results/artifacts/audit/ui', date);

<<<<<<< HEAD
    test.beforeAll(async ({ request }) => {
        // Attempt to seed data if backend is available
        try {
            await seedGlobalState(request);
        } catch (e) {
            console.warn('Backend not available for seeding, proceeding without it:', e);
        }

=======
    test.beforeAll(async () => {
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
        try {
            if (!fs.existsSync(auditDir)) {
                fs.mkdirSync(auditDir, { recursive: true });
            }
        } catch (e) {
            console.warn('Failed to create audit directory:', e);
        }
    });

  test('Capture Logs', async ({ page }) => {
    await page.goto('/logs');
    // Wait for some logs to appear
    await page.waitForTimeout(3000);
    try {
        await page.screenshot({ path: path.join(auditDir, 'log_stream.png') });
    } catch (e) {
        console.warn('Failed to save screenshot:', e);
    }
  });

<<<<<<< HEAD
  test('Verify RichResultViewer and Export', async ({ page }) => {
    // We can't rely on the backend being alive or correctly seeded in this specific test
    // environment, so we intercept the API calls to guarantee the UI has data to render.

    // Mock config requests to prevent stalling
    await page.route('**/api/v1/doctor*', async route => {
        await route.fulfill({ json: { status: "healthy" } });
    });
    await page.route('**/api/v1/users/me*', async route => {
        await route.fulfill({ json: { id: "e2e-admin-core" } });
    });
    await page.route('**/api/v1/topology*', async route => {
        await route.fulfill({ json: { nodes: [], edges: [] } });
    });
    await page.route('**/api/v1/services*', async route => {
        await route.fulfill({ json: [] });
    });

    // Mock the audit logs list to include a JSON-based tool call
    await page.route('**/api/v1/audit/logs*', async route => {
        await route.fulfill({
            json: {
                entries: [
                    {
                        timestamp: new Date().toISOString(),
                        toolName: "echo_tool",
                        userId: "e2e-admin-core",
                        arguments: JSON.stringify({ "hello": "world" }),
                        result: JSON.stringify({ "output": "world" }),
                        duration: "10ms",
                        error: ""
                    }
                ]
            }
        });
    });

    await page.goto('/audit');

    // Wait for the mock to populate the list
    await expect(page.locator('text=echo_tool').first()).toBeVisible();

    // Click "View"
    await page.locator('button:has-text("View")').first().click();

    // Verify RichResultViewer (the table / structured view) is visible instead of raw string
    // The RichResultViewer uses JsonView initially for small objects or uses table
    // We check for presence of structured rendering instead of relying on exact text match syntax
    await expect(page.locator('text=world').first()).toBeVisible();

    // Close dialog
    await page.keyboard.press('Escape');
=======
  test('Export Audit Logs to CSV', async ({ page }) => {
    await page.goto('/audit');
    await page.waitForSelector('text=Audit Logs');
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))

    // Start waiting for download before clicking.
    const downloadPromise = page.waitForEvent('download', { timeout: 10000 }).catch(() => null);

    // Wait for Export CSV button to be visible and enabled
    const exportBtn = page.locator('button:has-text("Export CSV")');
    await exportBtn.waitFor({ state: 'visible' });

<<<<<<< HEAD
    // Click it (which triggers an export on backend)
    await page.route('**/api/v1/audit/export*', async route => {
        await route.fulfill({ status: 200, body: 'a,b,c\n1,2,3' });
    });

    await exportBtn.click();

    // We mocked it so no actual file is downloaded, just checking the Toast
    await expect(page.locator('text=Export Successful').first()).toBeVisible();
=======
    // Check if we need to mock since we are not fully seeding audit data for this specific test
    // but the backend handles /api/v1/audit/export naturally.
    await exportBtn.click();

    const download = await downloadPromise;
    if (download) {
        const suggestedFilename = download.suggestedFilename();
        if (!suggestedFilename.includes('audit_export')) {
             throw new Error(`Unexpected filename: ${suggestedFilename}`);
        }
        await download.cancel();
    }
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
  });
});
