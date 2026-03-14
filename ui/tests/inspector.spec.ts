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

    // Click the row to open the details sheet
    await row.click();

    // Verify the details sheet opens. It contains the trace ID which starts with "trace-seed-"
    const sheet = page.getByRole('dialog');
    await expect(sheet).toBeVisible();

    // Check that we see some details of the trace
    await expect(sheet.locator('text=orchestrator-task').first()).toBeVisible();

    // Find a tool span inside the details sheet to replay
    // The mock trace has a child span named 'search-tool'
    const searchToolRow = page.locator('text=search-tool').first();
    await expect(searchToolRow).toBeVisible({ timeout: 10000 });
    await searchToolRow.click();

    // Check that we are now looking at the search-tool details
    await expect(sheet.locator('h3:has-text("Root Output")')).toBeVisible();

    // Click the "Replay & Diff" button
    const replayBtn = page.getByRole('button', { name: 'Replay & Diff' });
    await expect(replayBtn).toBeVisible();
    await replayBtn.click();

    // Verify the Replay & Diff dialog opens
    const replayDialog = page.getByRole('dialog').filter({ hasText: 'Replay & Diff Analysis' });
    await expect(replayDialog).toBeVisible();

    // Verify the DiffViewer component is rendered (Monaco editor container)
    // We check for the generic Monaco class or the original output heading
    await expect(page.getByText('Original Output')).toBeVisible();
  });
});
