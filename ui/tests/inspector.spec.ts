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
    const seedTraceBtn = page.getByRole('button', { name: 'Seed Trace' });
    await expect(seedTraceBtn).toBeVisible();
    await seedTraceBtn.click();

    // Expect the toast notification confirming the backend received the seed request
    await expect(page.getByText('Trace Seeded').first()).toBeVisible({ timeout: 5000 });

    // Wait briefly to allow React state to update based on WebSocket message
    await page.waitForTimeout(500);

    // The trace's root span name should appear in the inspector table.
    const row = page.locator('text=orchestrator-task').first();
    await expect(row).toBeVisible({ timeout: 10000 });

    // Click the row to open the detail sheet
    await row.click();

    // Verify the detail sheet opens and shows trace info
    const sheet = page.getByRole('dialog');
    await expect(sheet).toBeVisible();
    await expect(sheet.locator('text=orchestrator-task').first()).toBeVisible();
  });

  test('should clear traces permanently on backend when Clear is clicked', async ({ page }) => {
    // Navigate to the Inspector page
    await page.goto('/inspector');

    // Wait for the page to load by checking for the "Inspector" header
    await expect(page.getByRole('heading', { name: 'Inspector' })).toBeVisible();

    // Click the "Seed Trace" button
    const seedTraceBtn = page.getByRole('button', { name: 'Seed Trace' });
    await expect(seedTraceBtn).toBeVisible();
    await seedTraceBtn.click();

    // Expect the toast notification
    await expect(page.getByText('Trace Seeded').first()).toBeVisible({ timeout: 5000 });

    const row = page.locator('text=orchestrator-task').first();
    await expect(row).toBeVisible({ timeout: 10000 });

    // Click the Clear button
    const clearBtn = page.getByRole('button', { name: 'Clear' });
    await expect(clearBtn).toBeVisible();
    await clearBtn.click();

    // Verify the table is empty
    await expect(row).not.toBeVisible();
  });
});
