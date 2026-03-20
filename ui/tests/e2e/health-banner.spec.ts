/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('System Status Banner', () => {
  test('should not show connection error when backend is healthy', async ({ page }) => {
    // Navigate to any page
    await page.goto('/');

    // Wait for network idle to ensure the initial doctor call finishes
    await page.waitForLoadState('networkidle');

    // Verify "Connection Error" IS NOT visible
    const connectionErrorAlert = page.getByText(/Connection Error/i);
    await expect(connectionErrorAlert).not.toBeVisible();
  });
});
