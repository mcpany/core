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

  // Switch to JSON tab to verify raw schema
  await page.getByRole('tab', { name: 'JSON' }).click();

  // The schema content from mock: { type: "object", properties: { location: { type: "string" } } }
  // We check for "location" property in the JSON view
  await expect(page.locator('pre').filter({ hasText: /"location"/ })).toBeVisible();
  await expect(page.locator('pre').filter({ hasText: /"type": "object"/ })).toBeVisible();

  // Verify service name is shown in details (Scoped to the sheet)
  await expect(page.locator('div[role="dialog"]').getByText('weather-service')).toBeVisible();
});

test('Tool inspector loads historical metrics', async ({ page }) => {
  // Mock tools endpoint
  await page.route((url) => url.pathname.includes('/api/v1/tools'), async (route) => {
    await route.fulfill({
      json: {
        tools: [
          {
            name: 'get_metrics_tool',
            description: 'Test tool for metrics',
            source: 'configured',
            serviceId: 'metrics-service',
            inputSchema: { type: "object", properties: {} }
          }
        ]
      }
    });
  });

  // Mock audit logs endpoint
  await page.route((url) => url.pathname.includes('/api/v1/audit/logs'), async (route) => {
     const requestUrl = new URL(route.request().url());
     // Verify we are filtering by tool name
     if (!requestUrl.searchParams.get('tool_name')) {
         return route.fallback();
     }
     await route.fulfill({
         json: {
             entries: [
                 {
                     timestamp: new Date(Date.now() - 10000).toISOString(),
                     toolName: 'get_metrics_tool',
                     durationMs: 50,
                     error: ""
                 },
                 {
                     timestamp: new Date(Date.now() - 5000).toISOString(),
                     toolName: 'get_metrics_tool',
                     durationMs: 150,
                     error: "Something went wrong"
                 }
             ]
         }
     });
  });

  await page.goto('/tools');

  // Open inspector
  await page.locator('tr').filter({ hasText: 'get_metrics_tool' }).getByText('Inspect').click();

  // Click on "Performance & Analytics" tab
  await page.getByRole('tab', { name: 'Performance & Analytics' }).click();

  // Verify "Total Calls" is 2
  // We look for the "Total Calls" label and the value "2" nearby.
  await expect(page.getByText('Total Calls').locator('..').getByText('2', { exact: true })).toBeVisible();

  // Verify "Error Count" is 1
  await expect(page.getByText('Error Count').locator('..').getByText('1', { exact: true })).toBeVisible();
});
