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
  await expect(page.getByRole('tab', { name: 'Schema' })).toBeVisible();

  // Switch to Schema tab to verify raw schema
  await page.getByRole('tab', { name: 'Schema' }).click();

  // The schema content from mock: { type: "object", properties: { location: { type: "string" } } }
  // We check for "location" property in the JSON View component (which renders as divs/spans, not textarea or pre usually)
  // But ToolArgumentsEditor uses JsonView which often renders text.
  // Actually, ToolArgumentsEditor has Tabs: Form, JSON, Schema.
  // The inspector defaults to "Test & Execute" -> Form.
  // The test clicks "JSON" tab of the Editor? No, previous test clicked "JSON".
  // Wait, I am editing the "Test & Execute" tab content.
  // The editor has 3 sub-tabs: Form, JSON, Schema.
  // If I click Schema tab, I see the schema.

  // Let's look at the code:
  // <TabsTrigger value="schema">Schema</TabsTrigger>

  // So:
  await page.getByRole('tab', { name: 'Schema' }).click();

  // JsonView usually renders keys and values.
  await expect(page.getByText('"location"')).toBeVisible();
  await expect(page.getByText('"string"')).toBeVisible();

  // Verify service name is shown in details (Scoped to the sheet)
  await expect(page.locator('div[role="dialog"]').getByText('weather-service')).toBeVisible();
});
