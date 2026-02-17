/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Services Verification', () => {
  test('should navigate to marketplace to add a new service', async ({ page }) => {
    // 1. View Service List
    await page.goto('/upstream-services');
    await expect(page.getByRole('heading', { level: 1, name: 'Upstream Services' })).toBeVisible();

    // 2. Click Add Service
    // Wait for hydration/network idle to ensure event handlers are attached
    // Also explicitly wait for the button to be enabled and stable
    const addButton = page.getByRole('button', { name: 'Add Service' });
    await expect(addButton).toBeVisible();
    await expect(addButton).toBeEnabled();

    // It is a button that opens a dialog (initially shows Template selector)
    await addButton.click();

    // 3. Verify Sheet/Dialog Opens
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('New Service')).toBeVisible();
  });
});
