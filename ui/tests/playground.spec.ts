/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Tool Configuration', () => {
  test('should allow configuring and running a tool via form', async ({ page }) => {
    // Mock the tools API response
    await page.route('/api/v1/tools', async route => {
      const json = {
        tools: [
          {
            name: 'weather_tool',
            description: 'Get weather info',
            inputSchema: {
              type: 'object',
              properties: {
                city: { type: 'string', description: 'City name' },
                days: { type: 'integer', description: 'Number of days' }
              },
              required: ['city']
            }
          }
        ]
      };
      await route.fulfill({ json });
    });

    // Mock the tool execution
    await page.route('/api/v1/execute', async route => {
      // Mock successful execution since we are using a fake tool 'weather_tool'
      // that doesn't exist on the backend.
      await route.fulfill({
        json: {
          content: [
            {
              type: 'text',
              text: 'Mock execution result'
            }
          ],
          isError: false
        }
      });
    });

    await page.goto('/playground');

    // Open Available Tools
    await page.getByRole('button', { name: /available tools/i }).click();

    // Click "Use Tool" for weather_tool
    // The sheet might be animating, so wait a bit or just look for the text
    await expect(page.getByText('weather_tool')).toBeVisible();
    await page.getByRole('button', { name: /use tool/i }).click();

    // Dialog should open
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('Configure weather_tool')).toBeVisible();

    // Fill form
    await page.getByLabel('city', { exact: false }).fill('San Francisco');
    await page.getByLabel('days').fill('5');

    // Run Tool
    await page.getByRole('button', { name: /run tool/i }).click();

    // Verify chat message
    // The message should appear in the chat.
    // "weather_tool {"city":"San Francisco","days":5}"
    await expect(page.getByText('weather_tool {"city":"San Francisco","days":5}')).toBeVisible();

    // Verify result (mock result)
    // "Mock execution result"
    await expect(page.getByText('Mock execution result')).toBeVisible();
  });
});
