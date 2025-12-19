/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('UI loads and displays title', async ({ page }) => {
  page.on('console', msg => console.log('PAGE LOG:', msg.text()));
  await page.goto('/');
  // Debug title
  console.log('Page Title:', await page.title());
  await expect(page).toHaveTitle(/MCPAny Manager/);
});

test('UI can list services (mocked)', async ({ page }) => {
  page.on('console', msg => console.log('PAGE LOG:', msg.text()));
  page.on('request', req => console.log('REQ:', req.url()));

  // Mock the API response for listing services
  await page.route(url => url.href.includes('/v1/services'), async route => {
    console.log("MOCKED ROUTE HIT:", route.request().url());
    const json = {
      services: [
        {
           id: "mock-service-1",
           name: "Mock Service",
           http_service: { address: "http://example.com" }
        }
      ]
    };
    await route.fulfill({ json });
  });

  await page.goto('/');
  // Verify that the mock service is displayed
  await expect(page.getByText('Mock Service')).toBeVisible({ timeout: 10000 });
});
