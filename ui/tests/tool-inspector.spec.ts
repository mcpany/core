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
  await expect(page.getByText('Schema', { exact: true })).toBeVisible();

  // Default is "Visual" schema view, let's verify visual elements first (ToolForm)
  await expect(page.getByLabel('location')).toBeVisible();

  // Switch to JSON tab to verify raw schema view if available in SchemaViewer logic
  // The inspector uses ToolForm, which has "Form", "JSON", "Schema" tabs.
  // The Schema tab shows the inputSchema JSON.
  await page.getByRole('tab', { name: 'Schema' }).click();

  // Check for schema JSON content
  // Note: JsonView component might render differently than a simple pre tag, but let's check for text content in the view.
  // It usually renders with keys and values.
  await expect(page.getByText('"location"')).toBeVisible();
  // "type" might appear multiple times (root type, property type), so we use .first()
  await expect(page.getByText('"type"').first()).toBeVisible();

  // Verify service name is shown in details (Scoped to the sheet)
  await expect(page.locator('div[role="dialog"]').getByText('weather-service')).toBeVisible();
});
