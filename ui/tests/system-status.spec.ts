/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('System Status', () => {
  test('should display status indicator in header', async ({ page }) => {
    // Mock the doctor status endpoint
    await page.route('**/api/v1/doctor/status', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          status: 'healthy',
          timestamp: new Date().toISOString(),
          checks: {
            database: { status: 'ok', message: 'Connected' },
            network: { status: 'ok', latency: '10ms' }
          }
        }),
      });
    });

    await page.goto('/');

    // Wait for the indicator to appear
    const indicator = page.getByTitle('System Status');
    await expect(indicator).toBeVisible();
    await expect(indicator).toHaveText(/Healthy/);
  });

  test('should open status sheet on click', async ({ page }) => {
    // Mock the doctor status endpoint
    await page.route('**/api/v1/doctor/status', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          status: 'healthy',
          timestamp: new Date().toISOString(),
          checks: {
            database: { status: 'ok', message: 'Connected' },
            network: { status: 'ok', latency: '10ms' }
          }
        }),
      });
    });

    await page.goto('/');

    const indicator = page.getByTitle('System Status');
    await indicator.click();

    // Verify sheet content
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('System Status', { exact: true })).toBeVisible();
    await expect(page.getByText('Real-time diagnostics')).toBeVisible();
    await expect(page.getByText('Connected')).toBeVisible();
  });
});
