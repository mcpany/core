/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Inspector Page', () => {
  test('should allow seeding a trace from backend and viewing it', async ({ page }) => {
    // Navigate to the Inspector page
    await page.goto('/inspector');

    // Wait for the page to load by checking for the "Inspector" header
    await expect(page.getByRole('heading', { name: 'Inspector' })).toBeVisible();

    // Either we see the skeleton loader OR we already reached the empty state
    // We cannot reliably assert the skeleton loader if the API resolves instantly.
    await expect(page.getByText('Loading traces...')).not.toBeVisible();

    // Wait for the "No traces found" state since we just started
    // If the test has already run in the backend, traces might exist.
    // So we first clear to guarantee a clean slate before seeding.
    const setupClearBtn = page.getByRole('button', { name: 'Clear' });
    await expect(setupClearBtn).toBeVisible();
    await setupClearBtn.click();
    await expect(page.getByText('No traces found')).toBeVisible({ timeout: 10000 });

    // Click the "Seed Trace" button
    const seedTraceBtn = page.getByRole('button', { name: 'Seed Trace' });
    await expect(seedTraceBtn).toBeVisible();
    await seedTraceBtn.click();

    // Expect the toast notification confirming the action
    await expect(page.getByText('Trace Seeded').first()).toBeVisible();

    // The backend generated trace has the name "orchestrator-task" in the root span
    // and it appears in the table. We wait for it to be visible.
    const row = page.locator('text=orchestrator-task').first();
    await expect(row).toBeVisible({ timeout: 10000 });

    // Click the row to open the details sheet
    await row.click();

    // Verify the details sheet opens. It contains the trace ID which starts with "trace-seed-"
    const sheet = page.getByRole('dialog');
    await expect(sheet).toBeVisible();

    // Check that we see some details of the trace
    await expect(sheet.locator('text=orchestrator-task').first()).toBeVisible();

    // Close the sheet
    await page.keyboard.press('Escape');

    // Wait a bit to ensure the trace has finished populating before clear
    await page.waitForTimeout(1000);

    // Now let's test the Bulk Delete / Clear functionality
    const clearBtn = page.getByRole('button', { name: 'Clear' });
    await expect(clearBtn).toBeVisible();
    await clearBtn.click();

    // Verify it returns to the "No traces found" state
    await expect(page.getByText('No traces found')).toBeVisible({ timeout: 20000 });
  });
});
