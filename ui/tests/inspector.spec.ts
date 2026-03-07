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

    // Assert that the skeleton loading UI appears when there are no traces yet
    // The skeleton has class "animate-pulse"
    await expect(page.locator('.animate-pulse').first()).toBeVisible({ timeout: 5000 });

    // Wait for initial load to finish (and skeleton to potentially disappear if traces load fast,
    // but in a fresh DB it will stay or show "No traces found")
    await page.waitForTimeout(1000);

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
  });
});
