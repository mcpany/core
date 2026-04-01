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

    // Ensure traces are cleared first
    const clearBtn = page.getByRole('button', { name: 'Clear' });
    await clearBtn.click();

    // Click the "Seed Trace" button
    const seedTraceBtn = page.getByRole('button', { name: 'Seed Trace' });
    await expect(seedTraceBtn).toBeVisible();
    await seedTraceBtn.click();

    // Expect the toast notification
    await expect(page.getByText('Trace Seeded').first()).toBeVisible({ timeout: 5000 });

    // The injected trace's root span name should appear in the inspector table.
    const row = page.locator('text=orchestrator-task').first();
    await expect(row).toBeVisible({ timeout: 10000 });

    // Click the row to open the detail sheet
    await row.click();

    // Verify the detail sheet opens and shows trace info
    const sheet = page.getByRole('dialog');
    await expect(sheet).toBeVisible();
    await expect(sheet.locator('text=orchestrator-task').first()).toBeVisible();

    // Test the Replay & Diff feature
    const replayBtn = sheet.getByRole('button', { name: 'Replay & Diff' });
    await expect(replayBtn).toBeVisible();
    await replayBtn.click();

    const replayDialog = page.getByRole('dialog').filter({ hasText: 'Replay & Diff Analysis' });
    await expect(replayDialog).toBeVisible();

    // The replay dialog should run the rerun and we should see "Revenue up 15%" from our real backend mock
    await expect(replayDialog.locator('text=Revenue up 15%').first()).toBeVisible({ timeout: 10000 });
  });

  test('should clear traces permanently on backend when Clear is clicked', async ({ page }) => {
    await page.goto('/inspector');
    await expect(page.getByRole('heading', { name: 'Inspector' })).toBeVisible();

    // Seed to ensure we have something to clear
    const seedTraceBtn = page.getByRole('button', { name: 'Seed Trace' });
    await seedTraceBtn.click();
    await expect(page.getByText('Trace Seeded').first()).toBeVisible({ timeout: 5000 });

    const row = page.locator('text=orchestrator-task').first();
    await expect(row).toBeVisible({ timeout: 10000 });

    const clearBtn = page.getByRole('button', { name: 'Clear' });
    await expect(clearBtn).toBeVisible();
    await clearBtn.click();

    await expect(row).not.toBeVisible();
  });
});
