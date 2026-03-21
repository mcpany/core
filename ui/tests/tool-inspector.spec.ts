/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test('Tools page loads and inspector opens', async ({ page, request }) => {
  // Seed a real service with a tool
  const serviceName = 'weather-service';
  await request.post('/api/v1/services', {
    data: {
      name: serviceName,
      command_line_service: {
        command: 'echo',
        tools: [
          {
            name: 'get_weather',
            description: 'Get weather for a location',
            inputSchema: {
               type: "object",
               properties: {
                 location: { type: "string" }
               }
            }
          }
        ],
        calls: { 'get_weather': { args: ['{}'] } }
      }
    }
  });

  await page.goto('/tools');

  // Check if tools are listed
  await expect(page.getByText('get_weather')).toBeVisible();

  // Open inspector for get_weather
  await page.locator('tr').filter({ hasText: 'get_weather' }).getByText('Inspect').click();

  // Check if inspector sheet is open (Wait for title)
  await expect(page.getByText('get_weather').first()).toBeVisible();

  // Switch to Schema tab
  await page.getByRole('tab', { name: 'Schema' }).click();

  // Switch to JSON sub-tab to verify raw schema
  // Ensure the Schema tab content is visible first
  // Switch to JSON sub-tab to verify raw schema
  // Ensure the Schema tab content is visible first. We identify it by the "Visual" tab inside it.
  const schemaPanel = page.getByRole('tabpanel').filter({ hasText: 'Visual' });
  await expect(schemaPanel).toBeVisible();

  // Click the JSON trigger inside the schema content
  await schemaPanel.getByRole('tab', { name: 'JSON' }).click();

  // The schema content from mock: { type: "object", properties: { location: { type: "string" } } }
  // We check for "location" property in the JSON view
  await expect(page.locator('pre').filter({ hasText: /"location"/ })).toBeVisible();
  await expect(page.locator('pre').filter({ hasText: /"type": "object"/ })).toBeVisible();

  // Verify service name is shown in details (Scoped to the sheet)
  await expect(page.locator('div[role="dialog"]').getByText('weather-service')).toBeVisible();
});
