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
    // It is a button that opens a dialog (initially shows Template selector)
    await page.getByRole('button', { name: 'Add Service' }).click();

    // 3. Verify Sheet/Dialog Opens
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('New Service')).toBeVisible();
  });
});
