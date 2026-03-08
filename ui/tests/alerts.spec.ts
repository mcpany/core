/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Alerts Page', () => {
  test('should load alerts page and display key elements', async ({ page }) => {
    // Navigate to alerts page
    await page.goto('/alerts');

    // Check header
    await expect(page.getByRole('heading', { name: 'Alerts & Incidents' })).toBeVisible();

    // Check stats cards labels
    await expect(page.getByText('Active Critical')).toBeVisible();
    await expect(page.getByText('Active Warnings')).toBeVisible();
    await expect(page.getByText('MTTR (Today)')).toBeVisible();
    await expect(page.getByText('Total Incidents')).toBeVisible();

    // Check that stats values are fetched from the API and rendered
    // The backend mock data has Active Critical: 1, Active Warning: 1, MTTR: 14m, Total: >0 (we can just verify it's rendered)
    await expect(page.getByText('14m')).toBeVisible();

    // Check table content (mock data)
    await expect(page.getByText('High CPU Usage')).toBeVisible();
    await expect(page.getByText('API Latency Spike')).toBeVisible();
  });

  test('should filter alerts', async ({ page }) => {
    await page.goto('/alerts');

    // Type in search box - use getByPlaceholder if available, else locator
    const searchBox = page.locator('input[placeholder="Search alerts by title, message, service..."]');
    await searchBox.fill('CPU');

    // Should see CPU alert
    await expect(page.getByText('High CPU Usage')).toBeVisible();

    // Should NOT see Latency alert
    await expect(page.getByText('API Latency Spike')).toBeHidden();
  });

  test('should open create rule dialog', async ({ page }) => {
    await page.goto('/alerts');

    // Click "New Alert Rule" button
    await page.getByRole('button', { name: 'New Alert Rule' }).click();

    // Check dialog opens
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Create Alert Rule' })).toBeVisible();

    // Close it
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('dialog')).toBeHidden();
  });

  test('should acknowledge alert via dropdown', async ({ page }) => {
    await page.goto('/alerts');

    // Find an active alert row (mock data usually has some)
    // We target the row with "High CPU Usage" which is active in mock
    const row = page.getByRole('row').filter({ hasText: 'High CPU Usage' });

    // Wait for row to be stable before acting
    await expect(row).toBeVisible();

    // Click the "More Actions" dropdown button in that row
    await row.getByRole('button', { name: 'Open menu' }).click();

    const ackMenuitem = page.getByRole('menuitem', { name: 'Acknowledge' });
    await ackMenuitem.waitFor({ state: 'visible' });

    // Click Acknowledge and wait for the mock API to process it so the state updates
    await Promise.all([
      page.waitForResponse(resp => resp.url().includes('/api/v1/alerts') && resp.status() === 200),
      ackMenuitem.evaluate((node) => node.click())
    ]);

    // Verify toast appears indicating success
    await expect(page.getByText('Status Updated')).toBeVisible();

    // Re-fetch the row to avoid staleness and verify text contains 'acknowledged'
    await page.reload();
    await expect(page.getByRole('row').filter({ hasText: 'High CPU Usage' }).locator('span.capitalize')).toHaveText('acknowledged', { timeout: 10000 });
  });

  test('should resolve alert via dropdown', async ({ page }) => {
    await page.goto('/alerts');

    // Find an acknowledged or active alert
    const row = page.getByRole('row').filter({ hasText: 'Disk Space Low' });

    await expect(row).toBeVisible();

    // Click "More Actions"
    await row.getByRole('button', { name: 'Open menu' }).click();

    // Click "Resolve" using evaluate to bypass non-standard disabled checks
    const resolveMenuitem = page.getByRole('menuitem', { name: 'Resolve' });
    await resolveMenuitem.waitFor({ state: 'visible' });

    await Promise.all([
      page.waitForResponse(resp => resp.url().includes('/api/v1/alerts') && resp.status() === 200),
      resolveMenuitem.evaluate((node) => node.click())
    ]);

    await expect(page.getByText('Status Updated')).toBeVisible();

    await page.reload();
    await expect(page.getByRole('row').filter({ hasText: 'Disk Space Low' }).locator('span.capitalize')).toHaveText('resolved', { timeout: 10000 });
  });

  test('should bulk acknowledge alerts', async ({ page }) => {
    await page.goto('/alerts');

    // Wait for the alerts table to be populated
    await expect(page.getByText('High CPU Usage')).toBeVisible();
    await expect(page.getByText('API Latency Spike')).toBeVisible();

    // Select the first alert row using the checkbox
    const firstRow = page.getByRole('row').filter({ hasText: 'High CPU Usage' });
    await firstRow.getByRole('checkbox').click();

    // Select the second alert row using the checkbox
    const secondRow = page.getByRole('row').filter({ hasText: 'API Latency Spike' });
    await secondRow.getByRole('checkbox').click();

    // Verify that the bulk action bar is visible
    const actionBarText = page.getByText('2 selected');
    await expect(actionBarText).toBeVisible();

    // Click the bulk "Acknowledge" button
    const acknowledgeButton = page.getByRole('button', { name: 'Acknowledge', exact: true });

    // We expect two PUT requests. Wait for the responses to settle.
    const requestCount = 2;
    let completedRequests = 0;
    const responsePromise = new Promise<void>((resolve, reject) => {
        const timeout = setTimeout(() => reject(new Error('Timeout waiting for responses')), 10000);
        page.on('response', response => {
            if (response.url().includes('/api/v1/alerts') && response.request().method() === 'PUT' && response.status() === 200) {
                completedRequests++;
                if (completedRequests === requestCount) {
                    clearTimeout(timeout);
                    resolve();
                }
            }
        });
    });

    await acknowledgeButton.click();
    await responsePromise;

    // Verify toast notification appears (checking for specific text or role)
    await expect(page.getByText('Bulk Update Successful', { exact: true })).toBeVisible();

    // Verify the bulk action bar is hidden after success
    await expect(actionBarText).toBeHidden();

    await page.reload();
    await expect(page.getByRole('row').filter({ hasText: 'High CPU Usage' }).locator('span.capitalize')).toHaveText('acknowledged', { timeout: 10000 });
    await expect(page.getByRole('row').filter({ hasText: 'API Latency Spike' }).locator('span.capitalize')).toHaveText('acknowledged', { timeout: 10000 });
  });
});
