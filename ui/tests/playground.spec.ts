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
    await expect(page.getByText('Mock execution result')).toBeVisible();
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

  test('should create, list and use presets', async ({ page }) => {
    page.on('console', msg => console.log(`[Browser Console] ${msg.text()}`));

    // Mock Tools
    await page.route('/api/v1/tools', async route => {
      await route.fulfill({
        json: {
          tools: [{
            name: 'weather_tool',
            description: 'Get weather info',
            inputSchema: {
              type: 'object',
              properties: { city: { type: 'string' } },
              required: ['city']
            }
          }]
        }
      });
    });

    // Mock Presets CRUD
    let presets: any[] = [];
    await page.route('**/api/v1/tool-presets', async route => {
        console.log('Mock hit:', route.request().method(), route.request().url());
        if (route.request().method() === 'GET') {
            await route.fulfill({ json: presets });
        } else if (route.request().method() === 'POST') {
            const data = route.request().postDataJSON();
            data.id = 'preset-' + Date.now();
            presets.push(data);
            await route.fulfill({ json: data });
        } else {
             await route.continue();
        }
    });

    await page.route('**/api/v1/tool-presets/*', async route => {
        console.log('Mock hit detail:', route.request().method(), route.request().url());
        if (route.request().method() === 'DELETE') {
            const id = route.request().url().split('/').pop();
            presets = presets.filter(p => p.id !== id);
            await route.fulfill({ json: {} });
        } else {
             await route.continue();
        }
    });

    await page.goto('/playground');

    // 1. Create a preset
    await expect(page.getByText('weather_tool')).toBeVisible();
    await page.getByRole('button', { name: 'Use', exact: true }).click();

    // Fill form
    await page.getByLabel('city').fill('New York');

    // Open Presets Popover in Form
    await page.getByTitle('Manage Presets').click();

    // Click Create New
    await page.getByTitle('Create New Preset').click();

    // Fill Name and Save
    await page.getByPlaceholder('Preset Name').fill('NYC Weather');
    // Save button (icon only)
    const saveBtn = page.locator('button:has(svg.lucide-save)');
    await expect(saveBtn).toBeEnabled();
    await saveBtn.click();

    // Verify Toast
    // Wait for either success or failure to debug
    await expect(page.getByText(/Preset saved|Failed to save/)).toBeVisible();
    await expect(page.getByText('Preset saved')).toBeVisible();

    // Close Dialog
    await page.getByRole('button', { name: 'Cancel' }).click();

    // 2. Use Preset from Sidebar
    // Switch to Presets Tab
    await page.getByRole('tab', { name: 'Presets' }).click();

    // Verify Preset is listed
    await expect(page.getByText('NYC Weather')).toBeVisible();

    // Click Preset
    await page.getByText('NYC Weather').click();

    // Verify Dialog opens with pre-filled data
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByLabel('city')).toHaveValue('New York');
  });

});
