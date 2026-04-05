/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('System Status Banner', () => {
  test('should show connection error when backend is unreachable', async ({ page }) => {
    // Mock health failure
    await page.route('**/api/v1/doctor', route => route.fulfill({ status: 500 }));

    // Navigate to any page
    await page.goto('/');

    // Verify "Backend Unreachable" IS visible
    const connectionErrorAlert = page.getByText(/Backend Unreachable/i);
    await expect(connectionErrorAlert).toBeVisible({ timeout: 15000 });

    // Verify specific description IS visible
    const errorDescription = page.getByText(/We've lost connection to the MCP Any backend/i);
    await expect(errorDescription).toBeVisible({ timeout: 15000 });
  });
});
