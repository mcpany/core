/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Complex Schema Support', () => {
  test('should allow configuring and running a tool with complex nested schema', async ({ page }) => {
    // Mock the tools API to return a tool with complex schema
    await page.route('**/api/v1/tools', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          tools: [
            {
              name: 'complex_tool',
              description: 'A tool with complex schema',
              serviceName: 'test',
              inputSchema: {
                type: 'object',
                properties: {
                  user: {
                    type: 'object',
                    required: ['name'],
                    properties: {
                      name: { type: 'string' },
                      age: { type: 'integer' },
                      active: { type: 'boolean' }
                    }
                  },
                  tags: {
                    type: 'array',
                    items: { type: 'string' }
                  }
                },
                required: ['user']
              }
            }
          ]
        })
      });
    });

    // Mock the execute API
    await page.route('**/api/v1/execute', async (route) => {
        const body = JSON.parse(route.request().postData() || '{}');
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
                content: [{ type: "text", text: `Executed ${body.name} with args: ${JSON.stringify(body.arguments)}` }]
            })
        });
    });

    // Navigate to playground
    await page.goto('/playground');
    // await expect(page.getByRole('heading', { name: 'Playground' })).toBeVisible();

    // Open tools list (Sidebar is open by default)
    // await page.getByRole('button', { name: 'Available Tools' }).click();

    // Select the complex tool
    await expect(page.getByText('complex_tool')).toBeVisible();
    // The button just says "Use"
    await page.getByRole('button', { name: 'Use', exact: true }).click();

    // Verify form structure
    // Note: The UI might append type info like "user (object)", so we disable exact match
    await expect(page.getByRole('dialog').getByText('user', { exact: false })).toBeVisible();

    // Try to submit empty form (should fail validation because user.name is required)
    await page.getByRole('button', { name: 'Build Command' }).click();

    // Fill the form
    await page.getByPlaceholder('name').fill('Bob');
    await page.getByPlaceholder('0').fill('30');

    // Add tag
    await page.getByRole('button', { name: 'Add Item' }).click();
    await page.getByPlaceholder('Item 1').fill('developer');

    // Run Tool (Build Command)
    await page.getByRole('button', { name: 'Build Command' }).click();

    // Verify execution
    // The playground adds a user message with the command
    await expect(page.getByText('complex_tool', { exact: false })).toBeVisible();
    await expect(page.getByText('Bob', { exact: false })).toBeVisible();
    await expect(page.getByText('developer', { exact: false })).toBeVisible();

    // Verify result
    await expect(page.locator('text=Executed complex_tool')).toBeVisible();
  });
});
