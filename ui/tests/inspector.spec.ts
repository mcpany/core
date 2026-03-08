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

    // Wait for initial load
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

    // The child span (e.g., 'search-tool') should not be visible initially
    const childSpan = page.locator('text=search-tool').first();
    await expect(childSpan).not.toBeVisible();

    // Find the expand chevron for the row
    // The chevron is placed inside a div with class 'cursor-pointer' within the table cell
    const expandIcon = page.locator('tr').filter({ hasText: 'orchestrator-task' }).locator('.cursor-pointer').first();
    await expect(expandIcon).toBeVisible();

    // Click the expand icon to reveal children
    await expandIcon.click();

    // Verify the child span is now visible
    await expect(childSpan).toBeVisible();

    // Verify the details sheet does NOT open (this is the bug we fixed)
    const sheet = page.getByRole('dialog');
    await expect(sheet).not.toBeVisible();

    // Now click the text of the row itself to verify the details sheet STILL opens correctly
    await row.click();
    await expect(sheet).toBeVisible();
  });
});
