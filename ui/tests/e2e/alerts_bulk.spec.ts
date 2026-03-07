/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

// Use the exact same mocked data test structure as ui/tests/alerts.spec.ts,
// since that file also tests the Alerts Page but WITHOUT mocks. It relies on the default
// dev server startup state that provides these exact alerts via the real mock data server
// running in backend during tests.
test.describe('Alerts Bulk Actions (Real Data)', () => {

  test('should select multiple alerts and show bulk action bar', async ({ page }) => {
    await page.goto('/alerts');

    // Wait for the alerts to load. The backend mock server always provides "High CPU Usage"
    await expect(page.getByText('High CPU Usage')).toBeVisible();
    await expect(page.getByText('API Latency Spike')).toBeVisible();

    // Check "Select All" checkbox
    const selectAllCheckbox = page.getByRole('checkbox', { name: 'Select all' });
    await selectAllCheckbox.check();

    // Verify bulk action bar appears (number of default alerts might vary, so we just check for action bar existence)
    await expect(page.getByRole('button', { name: 'Acknowledge Selected' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Resolve Selected' })).toBeVisible();
  });

  test('should process bulk actions and optimistically update', async ({ page }) => {
      await page.goto('/alerts');
      await expect(page.getByText('High CPU Usage')).toBeVisible();
      await expect(page.getByText('API Latency Spike')).toBeVisible();

      // Select first two specific alerts we know the backend provides in dev/test mode
      await page.getByRole('checkbox', { name: 'Select High CPU Usage' }).check();
      await page.getByRole('checkbox', { name: 'Select API Latency Spike' }).check();

      // Click Bulk Acknowledge
      await page.getByRole('button', { name: 'Acknowledge Selected' }).click();

      // Verify optimistic update and toast. The mock backend will process these just fine.
      await expect(page.getByText('2 alerts marked as acknowledged')).toBeVisible();

      // Ensure the checkboxes are cleared and action bar is gone
      await expect(page.getByRole('button', { name: 'Acknowledge Selected' })).toBeHidden();
  });

});
