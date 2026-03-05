/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';

test.describe('System Status Banner', () => {
  // Use real backend failure instead of mocking.
  test('should show connection error when backend is down', async ({ page }) => {
    // Navigate to a page to ensure app loads with backend running first
    await page.goto('/');
    await expect(page.locator('text=MCP Any')).toBeVisible({ timeout: 15000 });

    // Stop the backend process for a true E2E test of connection failure
    try {
        execSync('pkill -f "build/test/bin/server" || true');
    } catch(e) {}

    // Verify "Connection Error" IS visible
    const connectionErrorAlert = page.getByText(/Connection Error/i);
    await expect(connectionErrorAlert).toBeVisible({ timeout: 15000 });

    // Verify specific description IS visible
    const errorDescription = page.getByText(/Could not connect to the server health check/i);
    await expect(errorDescription).toBeVisible({ timeout: 15000 });

    // Restart the backend to clean up for other tests
    try {
        execSync('make -C ../server run-test-server &');
    } catch(e) {}
  });
});
