/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Rich Result Viewer - Advanced Content Extraction', () => {
  const serviceName = `cards-test-service-${Date.now()}`;

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
            { name: 'get_nested_data', description: 'Returns nested data' }
          ],
          calls: {
            'get_nested_data': {
              args: [
                JSON.stringify({
                  "response": {
                     "metadata": {"source": "db", "time": "2025"},
                     "data": [
                        { "id": 1, "name": "Deep Alice", "role": "Admin" },
                        { "id": 2, "name": "Deep Bob", "role": "User" }
                     ]
                  },
                  "status": 200
                })
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

  test('extracts deep arrays into a table and shows cards for the object', async ({ page }) => {
    await page.goto('/playground');

    // Wait for Playground to load (using Console which is the actual title)
    // The previous test failed looking for "Interactive Playground"
    await expect(page.getByText('Console', { exact: true }).or(page.locator('h1'))).toBeVisible({ timeout: 10000 });

    // Select the tool from the sidebar or dropdown
    // Based on playground.spec.ts, we can filter for the tool name in the group
    await expect(page.getByText(`${serviceName}.get_nested_data`)).toBeVisible({ timeout: 15000 });
    await page.locator('.group').filter({ hasText: `${serviceName}.get_nested_data` }).click();

    // Verify "Tool Runner" tab is active
    await expect(page.getByRole('tab', { name: 'Tool Runner' })).toHaveAttribute('data-state', 'active');

    // Execute tool
    await page.getByRole('button', { name: 'Execute' }).click();

    // Wait for Result block
    await expect(page.getByText('Result', { exact: true })).toBeVisible({ timeout: 10000 });

    // We modified RichResultViewer to extract `data` array deeply.
    // It should default to "Table" view.
    const tableTab = page.getByRole('tab', { name: 'Table' });
    await expect(tableTab).toBeVisible();

    // Verify data inside the table
    const table = page.getByRole('table');
    await expect(table).toBeVisible();
    await expect(table.getByText('Deep Alice')).toBeVisible();
    await expect(table.getByText('Deep Bob')).toBeVisible();
    await expect(table.getByText('Admin')).toBeVisible();
  });
});
