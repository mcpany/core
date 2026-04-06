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
    await page.goto('/tools');

    // Search for the test tool
    await page.getByPlaceholder('Search tools...').fill('get_complex_data');
    await expect(page.getByText('rich-result-test-service.get_complex_data').first()).toBeVisible({ timeout: 10000 });

    // Open inspector
    await page.getByRole('row', { name: 'rich-result-test-service.get_complex_data' }).getByRole('button', { name: 'Inspect' }).click();

    // Wait for inspector to open
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await expect(dialog.getByRole('heading', { name: 'rich-result-test-service.get_complex_data', exact: true })).toBeVisible();

    // Execute tool (default args should work as they are empty object in seeded tool)
    await page.getByRole('button', { name: 'Execute' }).click();

    // Wait for result
    // Use precise selector to avoid matching service name "rich-result-test-service"
    await expect(page.locator('label').filter({ hasText: 'Result' })).toBeVisible({ timeout: 10000 });

    // Check if Table toggle button is available
    const tableBtn = page.getByRole('button', { name: /Table/i });
    await expect(tableBtn).toBeVisible();

    // It might default to Table view because it's eligible
    // Verify content in table
    const table = page.getByRole('table');
    await expect(table).toBeVisible();

    // Verify data
    await expect(table.getByText('Alice')).toBeVisible();
    await expect(table.getByText('Bob')).toBeVisible();
    await expect(table.getByText('Admin')).toBeVisible();

    // Switch to JSON view
    // SmartResultRenderer renders a button with name JSON instead of a tab
    await page.getByRole('button', { name: /JSON/i }).last().click();

    // Check for JSON content - tokenized render may split punctuation into spans
    // In raw JSON view, Alice will be rendered
    await expect(page.getByText('Alice').first()).toBeVisible();
  });
});
