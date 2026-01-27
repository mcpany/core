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
    // It is a link now
    await page.getByRole('link', { name: 'Add Service' }).click();

    // 3. Verify Navigation
    await expect(page).toHaveURL(/\/marketplace\?tab=local/);
    await expect(page.getByRole('heading', { name: 'Marketplace' })).toBeVisible();
  });
});
