/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('System Status', () => {
  test('should display status indicator in header', async ({ page }) => {
    // Monitor console logs for debugging
    page.on('console', msg => {
      if (msg.type() === 'error')
        console.log(`[Browser Console Error] ${msg.text()}`);
      else
        console.log(`[Browser Console] ${msg.text()}`);
    });

    // Mock the doctor status endpoint
    await page.route('**/api/v1/doctor', async (route) => {
      console.log('Mocking /api/v1/doctor response');
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

    // Check if we hit the Disconnected state which implies mock failure or fetch error
    // We expect it to NOT be visible or contain "Disconnected"
    await expect(indicator).toBeVisible();

    // Explicitly check for Healthy to confirm data load
    // Using a more lenient text matcher or ensuring we wait long enough
    await expect(indicator).toContainText('Healthy', { timeout: 10000 });
  });

  test('should open status sheet on click', async ({ page }) => {
    page.on('console', msg => console.log(`[Browser Console] ${msg.text()}`));

    // Mock the doctor status endpoint
    await page.route('**/api/v1/doctor', async (route) => {
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
