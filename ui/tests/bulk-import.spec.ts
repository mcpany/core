/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Service Import', () => {
  test('should open bulk import wizard and navigate steps', async ({ page }) => {
    // Navigate to page
    await page.goto('/upstream-services');
    await expect(page.getByRole('heading', { level: 1, name: 'Upstream Services' })).toBeVisible();

    // Open Wizard
    await page.getByRole('button', { name: 'Bulk Import' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Bulk Service Import' })).toBeVisible();

    // Step 1: Source
    await expect(page.getByText('1. Source')).toHaveClass(/font-bold/);
    await expect(page.getByRole('tab', { name: 'Paste JSON/YAML' })).toBeVisible();

    // Use a CLI service that should pass validation on the server (ls command)
    const validCmdConfig = JSON.stringify([{
        name: "test-service-cmd-bulk",
        commandLineService: { command: "ls", workingDirectory: "/tmp" } // /tmp usually exists
    }]);

    await page.getByLabel('Configuration Content').fill(validCmdConfig);
    await page.getByRole('button', { name: 'Next: Validate' }).click();

    // Step 2: Validation
    // Wait for the step to change (header changes)
    await expect(page.getByText('2. Validate & Select')).toHaveClass(/font-bold/);

    // Check table row exists
    await expect(page.getByRole('cell', { name: 'test-service-cmd-bulk' })).toBeVisible();

    // Wait for validation success (Check icon or "CLI" badge)
    await expect(page.getByRole('cell', { name: 'CLI' })).toBeVisible();

    // Since it's valid, it should be selected by default and "Import Selected" enabled.
    const importBtn = page.getByRole('button', { name: 'Import Selected' });
    await expect(importBtn).toBeEnabled();

    // Click Import
    await importBtn.click();

    // Step 3: Result
    await expect(page.getByText('3. Import')).toHaveClass(/font-bold/); // Or Result step
    // My component shows "3. Import" highlighted for both import and result steps.

    await expect(page.getByText('Import Complete')).toBeVisible();
    await expect(page.getByText('Successfully imported 1 services.')).toBeVisible();

    // Close
    // Note: Dialog has a close 'X' button which also matches "Close" aria-label.
    // We target the explicit button in the wizard content which appears first in this context or matches exact text better.
    // Using .first() as the error output showed it as the first match.
    await page.getByRole('button', { name: 'Close' }).first().click();
    await expect(page.getByRole('dialog')).not.toBeVisible();

    // Verify it's in the list
    // Might need reload or it fetches automatically
    // The page fetches on success.
    await expect(page.getByRole('cell', { name: 'test-service-cmd-bulk', exact: true })).toBeVisible();
  });
});
