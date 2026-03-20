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

    // Wait for Playground to load
    await expect(page.getByText('Interactive Playground').first()).toBeVisible();

    // Select the tool from the dropdown
    const toolTrigger = page.locator('button', { hasText: 'Select Tool' });
    await toolTrigger.click();

    // Search and pick our seeded tool
    await page.getByPlaceholder('Search tools...').fill('get_nested_data');
    await page.getByText(`${serviceName}.get_nested_data`).click();

    // Execute tool
    await page.getByRole('button', { name: 'Execute' }).click();

    // Wait for Result block
    await expect(page.locator('label', { hasText: 'RESULT' })).toBeVisible({ timeout: 10000 });

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
