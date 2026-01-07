/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test('Tools page loads and inspector opens', async ({ page }) => {
  // Mock tools endpoint
  await page.route((url) => url.pathname.includes('/api/v1/tools'), async (route) => {
    await route.fulfill({
      json: {
        tools: [
          {
            name: 'get_weather',
            description: 'Get weather for a location',
            source: 'configured',
            serviceId: 'weather-service',
            inputSchema: {
               type: "object",
               properties: {
                 location: { type: "string" }
               }
            }
          }
        ]
      }
    });
  });

  await page.goto('/tools');

  // Check if tools are listed
  await expect(page.getByText('get_weather')).toBeVisible();

  // Open inspector for get_weather
  await page.getByRole('row', { name: 'get_weather' }).getByRole('button', { name: 'Inspect' }).click();

  // Check if inspector sheet is open
  await expect(page.getByRole('heading', { name: 'get_weather' })).toBeVisible();

  // Check if schema is displayed (using the new Sheet layout)
  await expect(page.getByText('Schema')).toBeVisible();

  // The schema content from mock: { type: "object", properties: { location: { type: "string" } } }
  // We check if "location" is visible in the dialog (using strict text match is flaky with syntax highlighting)
  const dialog = page.locator('div[role="dialog"]');
  await expect(dialog).toContainText('location');
  await expect(dialog).toContainText('"type": "object"');

  // Verify service name is shown in details (Scoped to the sheet)
  await expect(page.locator('div[role="dialog"]').getByText('weather-service')).toBeVisible();
});
