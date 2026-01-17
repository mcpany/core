/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('System Status Banner', () => {

  test.beforeEach(async ({ page }) => {
    // Reset any previous mocks
    await page.unrouteAll({ behavior: 'ignoreErrors' });
  });

  test('should not be visible when system is healthy', async ({ page }) => {
    // Mock healthy doctor response
    await page.route('**/doctor', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ status: 'healthy', checks: {} })
      });
    });

    await page.goto('/');
    await expect(page.locator('div[role="alert"]').filter({ hasText: 'System Status' })).not.toBeVisible();
    await expect(page.getByText('Connection Error')).not.toBeVisible();
    await expect(page.getByText('Configuration Error')).not.toBeVisible();
  });

  test('should show connection error when backend is unreachable', async ({ page }) => {
     // Mock network error
     await page.route('**/doctor', async route => {
       await route.abort('failed');
     });

     await page.goto('/');
     await expect(page.getByText('Connection Error')).toBeVisible();
     await expect(page.getByText('Could not connect to the server health check')).toBeVisible();
  });

  test('should show configuration error when config check fails', async ({ page }) => {
    await page.route('**/doctor', async route => {
      await route.fulfill({
        status: 200, // The endpoint might still return 200 even if checks fail, or 503. The frontend handles the body.
        contentType: 'application/json',
        body: JSON.stringify({
          status: 'degraded',
          checks: {
            configuration: {
              status: 'error',
              message: 'Invalid YAML syntax in config.yaml'
            }
          }
        })
      });
    });

    await page.goto('/');
    await expect(page.getByText('Configuration Error')).toBeVisible();
    await expect(page.getByText('Invalid YAML syntax in config.yaml')).toBeVisible();
  });

  test('should show degraded status for other check failures', async ({ page }) => {
    await page.route('**/doctor', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          status: 'degraded',
          checks: {
            database: {
              status: 'error',
              message: 'Connection timeout'
            },
            cache: {
                status: 'error',
                message: 'Redis unavailable'
            }
          }
        })
      });
    });

    await page.goto('/');
    const banner = page.locator('div[role="alert"]').filter({ hasText: 'System Status: Degraded' });
    await expect(banner).toBeVisible();
    await expect(banner.getByText('Database: Connection timeout')).toBeVisible();
    await expect(banner.getByText('Cache: Redis unavailable')).toBeVisible();
  });

});
