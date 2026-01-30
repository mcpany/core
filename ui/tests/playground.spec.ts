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

    // Open Available Tools (Sidebar is open by default)
    // await page.getByRole('button', { name: /available tools/i }).click();

    // Click "Use Tool" for weather_tool
    // The sheet might be animating, so wait a bit or just look for the text
    await expect(page.getByText('weather_tool')).toBeVisible();
    await page.getByRole('button', { name: 'Use', exact: true }).click();

    // Dialog should open
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'weather_tool' })).toBeVisible();

    // Fill form
    await page.getByLabel('city', { exact: false }).fill('San Francisco');
    await page.getByLabel('days').fill('5');

    // Run Tool
    await page.getByRole('button', { name: /build command/i }).click();
    await page.getByLabel('Send').click();

    // Verify chat message
    // The message should appear in the chat.
    // "weather_tool {"city":"San Francisco","days":5}"
    await expect(page.getByText('weather_tool {"city":"San Francisco","days":5}')).toBeVisible();

    // Verify result (mock result)
    // "Mock execution result"
    // Note: JsonView renders strings with quotes
    await expect(page.getByText('"Mock execution result"')).toBeVisible();
  });

  test('should display smart error diagnostics and allow retry', async ({ page }) => {
    // Mock the tools API response
    await page.route('/api/v1/tools', async route => {
      const json = {
        tools: [
          {
            name: 'timeout_tool',
            description: 'A tool that times out',
            inputSchema: { type: 'object', properties: {} }
          }
        ]
      };
      await route.fulfill({ json });
    });

    // Mock the tool execution failure
    await page.route('/api/v1/execute', async route => {
        await route.fulfill({
            status: 500,
            json: { error: "upstream request timed out after 30s" }
        });
    });

    await page.goto('/playground');

    // Wait for tool to appear
    await expect(page.getByText('timeout_tool')).toBeVisible();

    // Click "Use"
    await page.getByRole('button', { name: 'Use', exact: true }).click();

    // Build command (empty args)
    await page.getByRole('button', { name: /build command/i }).click();

    // Send
    await page.getByLabel('Send').click();

    // Verify error message appears
    await expect(page.getByText('upstream request timed out after 30s', { exact: true })).toBeVisible();

    // Verify Retry button appears
    const retryBtn = page.getByLabel('Retry command');
    await expect(retryBtn).toBeVisible();

    // Verify Smart Suggestion appears
    await expect(page.getByText('Suggestion')).toBeVisible();

    // Click Retry
    await retryBtn.click();

    // Verify input is populated
    await expect(page.getByRole('textbox', { name: /enter command/i })).toHaveValue(/timeout_tool/);
  });

});
