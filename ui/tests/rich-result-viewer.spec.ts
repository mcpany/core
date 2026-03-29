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
            { name: 'get_complex_data', call_id: 'call1', description: 'Returns complex data' },
            { name: 'get_nested_complex_data', call_id: 'call2', description: 'Returns deeply nested complex data' }
          ],
          calls: {
            'call1': {
              args: [
                JSON.stringify([
                  { name: 'Alice', role: 'Admin', id: 1 },
                  { name: 'Bob', role: 'User', id: 2 }
                ])
              ]
            },
            'call2': {
              args: [
                JSON.stringify([
                  {
                    id: 'usr_1',
                    profile: { name: 'Charlie', email: 'charlie@example.com' },
                    settings: { theme: 'dark', notifications: { email: true, push: false } },
                    metadata: { lastLogin: '2023-10-27T10:00:00Z', roles: ['superuser', 'admin'] }
                  },
                  {
                    id: 'usr_2',
                    profile: { name: 'Dana', email: 'dana@example.com' },
                    settings: { theme: 'light', notifications: { email: true, push: true } },
                    metadata: { lastLogin: '2023-10-26T15:30:00Z', roles: ['user'] }
                  }
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

  test('SmartTable renders inline popover with grid for nested objects', async ({ page }) => {
    await page.goto('/tools');

    await page.getByPlaceholder('Search tools...').fill('get_nested_complex_data');
    await expect(page.getByText('rich-result-test-service.get_nested_complex_data').first()).toBeVisible({ timeout: 10000 });

    await page.getByRole('row', { name: 'rich-result-test-service.get_nested_complex_data' }).getByRole('button', { name: 'Inspect' }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await expect(dialog.getByRole('heading', { name: 'rich-result-test-service.get_nested_complex_data', exact: true })).toBeVisible();

    await page.getByRole('button', { name: 'Execute' }).click();

    await expect(page.locator('label').filter({ hasText: 'Result' })).toBeVisible({ timeout: 10000 });

    const tableTab = page.getByRole('tab', { name: 'Table' });
    await expect(tableTab).toBeVisible();

    const table = page.getByRole('table');
    await expect(table).toBeVisible();

    // Verify base table content
    await expect(table.getByText('usr_1')).toBeVisible();

    // Click the cell containing the nested "profile" object (button reads "Object {2}")
    // Since we have multiple "Object {2}" cells, we scope to the first row's profile cell
    // but a simpler approach is just to click the first button with text "Object"
    const objectButton = page.getByRole('button', { name: /Object \{2\}/ }).first();
    await objectButton.click();

    // The popover should appear containing the key "name" and value "Charlie"
    const popover = page.getByRole('dialog').filter({ hasText: 'Charlie' });
    await expect(popover).toBeVisible();

    // Verify the grid-like structure is present (keys and values)
    await expect(popover.getByText('name')).toBeVisible();
    await expect(popover.getByText('Charlie')).toBeVisible();
    await expect(popover.getByText('email')).toBeVisible();
    await expect(popover.getByText('charlie@example.com')).toBeVisible();
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
    // Note: There might be multiple "JSON" tabs (one for schema, one for args, one for result)
    // We want the one in the result viewer. Since it's likely the last one rendered or scoped.
    // The tabs in RichResultViewer are: Table, JSON, Raw Output.
    // We can scope by finding the container.
    // Or just click the one that follows "Result".

    // Scoping to the result area
    // const resultArea = page.locator('.grid', { hasText: 'Result' }).last();
    // Actually "Result" label is inside a grid div.

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
