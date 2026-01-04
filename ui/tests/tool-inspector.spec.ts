/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test('Tools page loads and inspector opens', async ({ page }) => {
  // Mock tools endpoint
  await page.route((url) => url.pathname.includes('/api/tools'), async (route) => {
    await route.fulfill({
      json: [
          {
            name: 'get_weather',
            description: 'Get weather for a location',
            source: 'configured',
            serviceId: 'weather-service',
            schema: {
               type: "object",
               properties: {
                 location: { type: "string" }
               }
            }
          }
        ]
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
  // We can check if "location" is visible in the pre block
  await expect(page.getByText('"location"')).toBeVisible();
  await expect(page.getByText('"type": "object"')).toBeVisible();

  // Verify service name is shown in details (Scoped to the sheet)
  await expect(page.locator('div[role="dialog"]').getByText('weather-service')).toBeVisible();
});
