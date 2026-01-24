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

test('Inspector shows historical metrics', async ({ page }) => {
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
            inputSchema: { type: "object", properties: {} }
          }
        ]
      }
    });
  });

  // Mock audit logs endpoint
  await page.route((url) => url.pathname.includes('/api/v1/audit/logs'), async (route) => {
    await route.fulfill({
      json: {
        entries: [
          {
            timestamp: new Date().toISOString(),
            tool_name: 'get_weather',
            durationMs: 123,
            error: ""
          },
          {
            timestamp: new Date().toISOString(),
            tool_name: 'get_weather',
            durationMs: 456,
            error: "timeout"
          }
        ]
      }
    });
  });

  await page.goto('/tools');

  // Open inspector for get_weather
  await page.locator('tr').filter({ hasText: 'get_weather' }).getByText('Inspect').click();

  // Switch to Metrics tab
  await page.getByRole('tab', { name: 'Performance & Analytics' }).click();

  // Check stats
  // Total calls: 2
  await expect(page.locator('div').filter({ hasText: 'Total Calls' }).last().getByText('2', { exact: true })).toBeVisible();

  // Success Rate: 50%
  await expect(page.locator('div').filter({ hasText: 'Success Rate' }).last().getByText('50%')).toBeVisible();

  // Avg Latency: (123 + 456) / 2 = 289.5 -> 290ms
  await expect(page.locator('div').filter({ hasText: 'Avg Latency' }).last().getByText('290ms')).toBeVisible();
});
