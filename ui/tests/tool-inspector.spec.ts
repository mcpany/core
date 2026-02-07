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
  await page.locator('tr').filter({ hasText: 'get_weather' }).getByText('Inspect').click();

  // Check if inspector sheet is open (Wait for title)
  await expect(page.getByText('get_weather').first()).toBeVisible();

  // Check if schema is displayed (using the new Sheet layout)
  // With SchemaForm change, "Schema" might be inside a sub-tab or the label "Schema" might be on the page
  await expect(page.getByText('Arguments')).toBeVisible();

  // Switch to JSON tab to verify raw schema (since Form is now default)
  // Use exact: true to distinguish from other potential matches
  await page.getByRole('tab', { name: 'JSON', exact: true }).click();

  // The schema content from mock: { type: "object", properties: { location: { type: "string" } } }
  // We check for "location" property in the JSON view (textarea now)
  // Wait for the textarea to have the value or contain the text
  await expect(page.locator('textarea#args')).toBeVisible();

  // Since input is initially empty {}, and schema is in another tab (Schema tab)
  // We should check the Schema tab if we want to see "location" definition,
  // OR check that the Form has the "location" input.

  // Let's check the Schema tab for the definition
  await page.getByRole('tab', { name: 'Schema', exact: true }).click();
  // By default Schema tab shows "Visual", switch to "Raw" to see JSON pre block if needed,
  // or check Visual view. Visual view shows "location".
  await expect(page.getByText('location')).toBeVisible();

  // Switch back to Form to verify input exists
  await page.getByRole('tab', { name: 'Form', exact: true }).click();
  await expect(page.getByLabel('location')).toBeVisible();

  // Verify service name is shown in details (Scoped to the sheet)
  await expect(page.locator('div[role="dialog"]').getByText('weather-service')).toBeVisible();
});
