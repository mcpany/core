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

    // Click the "Seed Trace" button (triggers POST /api/v1/debug/traces on backend)
    // The backend actually seeds traces and broadcasts them over the WebSocket.
    const seedTraceBtn = page.getByRole('button', { name: 'Seed Trace' });
    await expect(seedTraceBtn).toBeVisible();
    await seedTraceBtn.click();

    // Expect the toast notification confirming the backend received the seed request
    await expect(page.getByText('Trace Seeded').first()).toBeVisible({ timeout: 5000 });

    // Wait briefly to allow React state to update based on WebSocket message
    await page.waitForTimeout(1000);

    // The injected trace's root span name should appear in the inspector table.
    // The backend `generateMockAuditEntries` creates traces starting with "orchestrator-task"
    const row = page.locator('text=orchestrator-task').first();
    await expect(row).toBeVisible({ timeout: 10000 });

    // Click the row to open the detail sheet
    await row.click();

    // Verify the detail sheet opens and shows trace info
    const sheet = page.getByRole('dialog');
    await expect(sheet).toBeVisible();
    await expect(sheet.locator('text=orchestrator-task').first()).toBeVisible();
  });

  test('should bulk delete traces permanently on backend', async ({ page }) => {
    await page.goto('/inspector');
    await expect(page.getByRole('heading', { name: 'Inspector' })).toBeVisible();

    // Clear any existing traces first
    const clearBtn = page.getByRole('button', { name: 'Clear All' });
    await expect(clearBtn).toBeVisible();
    await clearBtn.click();
    await page.waitForTimeout(500);

    // Seed a trace from backend
    const seedTraceBtn = page.getByRole('button', { name: 'Seed Trace' });
    await expect(seedTraceBtn).toBeVisible();
    await seedTraceBtn.click();

    // Wait for it to show up
    const row = page.locator('text=orchestrator-task').first();
    await expect(row).toBeVisible({ timeout: 10000 });

    // Select the trace using the Select All checkbox in the table header
    const selectAllCheckbox = page.getByRole('checkbox', { name: 'Select all' });
    await expect(selectAllCheckbox).toBeVisible();
    await selectAllCheckbox.click();

    // Click the Bulk Delete button
    const bulkDeleteBtn = page.getByRole('button', { name: /Delete \(\d+\)/ });
    await expect(bulkDeleteBtn).toBeVisible();
    await bulkDeleteBtn.click();

    // Expect success toast
    await expect(page.getByText('Traces Deleted').first()).toBeVisible({ timeout: 5000 });

    // Verify it disappears from the UI
    await expect(row).not.toBeVisible();
  });
});
