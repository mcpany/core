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

    // Wait a moment for rendering
    await page.waitForTimeout(1000);

    await expect(page.getByText('rich-result-test-service').first()).toBeVisible({ timeout: 10000 });

    // Open inspector
    await page.getByRole('row', { name: /get_complex_data/ }).getByRole('button', { name: 'Inspect' }).click();

    // Wait for inspector to open
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // Execute tool (default args should work as they are empty object in seeded tool)
    await page.getByRole('button', { name: 'Execute' }).click();

    // Wait for result
    // Use precise selector to avoid matching service name "rich-result-test-service"
    await expect(page.locator('label').filter({ hasText: 'Result' })).toBeVisible({ timeout: 10000 });

    // Check if Table tab is active or available
    const tableTab = page.getByRole('tab', { name: 'Table' });
    await expect(tableTab).toBeVisible();

    // It might default to Table view because it's eligible
    // Verify content in table
    const table = page.getByRole('table');
    await expect(table).toBeVisible();

    // Verify data
    await expect(table.getByText('Alice')).toBeVisible();
    await expect(table.getByText('Bob')).toBeVisible();
    await expect(table.getByText('Admin')).toBeVisible();

    // Switch to JSON tab
    // Let's try finding the tab list containing "Raw Output" which is unique to RichResultViewer
    const viewerTabs = page.locator('[role="tablist"]', { hasText: 'Raw Output' });
    await viewerTabs.getByRole('tab', { name: 'JSON' }).click();

    // Check for JSON content - tokenized render may split punctuation into spans
    await expect(page.getByText('Alice')).toBeVisible();

    // Switch to Raw Output tab
    await viewerTabs.getByRole('tab', { name: 'Raw Output' }).click();
    await expect(page.getByText('"stdout":')).toBeVisible();
  });
});
