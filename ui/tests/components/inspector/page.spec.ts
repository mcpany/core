/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Inspector Page Selection', () => {
    test('renders All Status and All Types appropriately', async ({ page }) => {
        await page.goto('/inspector');
        await expect(page.getByText('All Status')).toBeVisible();
        await expect(page.getByText('All Types')).toBeVisible();
    });
});
