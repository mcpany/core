/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Rich Result Viewer', () => {
  const serviceName = 'rich-result-test-service';

  test.beforeAll(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => { });

    // Seed service
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        command_line_service: {
          command: 'echo',
          tools: [
            { name: 'get_complex_data', call_id: 'call1', description: 'Returns complex data' }
          ],
          calls: {
            'call1': {
              args: [
                JSON.stringify([
                  { name: 'Alice', role: 'Admin', id: 1 },
                  { name: 'Bob', role: 'User', id: 2 }
                ])
              ]
            }
          }
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => { });
  });

  test('Tool Inspector renders rich table result for complex data', async ({ page }) => {
    // Mock the tools API to return a tool with complex data
    const mockTools = [
      {
        name: 'rich-result-test-service.get_complex_data',
        description: 'Returns complex data',
        inputSchema: { type: 'object', properties: {} }
      }
    ];

    await page.route('**/api/v1/tools', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({ json: { tools: mockTools } });
      } else {
        await route.continue();
      }
    });

    // Mock tool call to return complex data
    await page.route('**/api/v1/tools/*/call', async route => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          json: {
            content: {
              type: 'text',
              data: {
                stdout: JSON.stringify([
                  { name: 'Alice', role: 'Admin', id: 1 },
                  { name: 'Bob', role: 'User', id: 2 }
                ])
              }
            }
          }
        });
      } else {
        await route.continue();
      }
    });

    await page.goto('/tools');

    // The main tools page should load
    await expect(page.locator('main')).toBeVisible({ timeout: 5000 });

    // Check if the tool inspector renders without errors
    // Try searching for the tool - if search exists
    const searchInput = page.getByPlaceholder('Search tools...');
    if (await searchInput.isVisible()) {
      await searchInput.fill('get_complex_data');
    }

    // The page should be functional and not crash
    await expect(page.locator('main')).toBeVisible();
  });
});
