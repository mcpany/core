/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Inspector Page Selection', () => {
  test('renders All Status and All Types appropriately', async ({ page }) => {
    // Navigate to the Inspector page
    await page.goto('/inspector');

    // Wait for the page to load by checking for the "Inspector" header
    await expect(page.getByRole('heading', { name: 'Inspector' })).toBeVisible();

    // Verify the fix for the truth reconciliation audit report:
    // The "Status" dropdown should say "All Statuses"
    await expect(page.getByText('All Statuses')).toBeVisible();

    // The "Type" dropdown should say "All Types"
    await expect(page.getByText('All Types')).toBeVisible();
  });
});
