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
            { name: 'fail_tool', call_id: 'call2', description: 'Always fails' }
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
                args: ['{"error": "Simulated internal error"}']
            }
          }
        }
      }
    });
    if (!response.ok()) {
        console.error("Failed to seed service:", await response.text());
    }
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => { });
  });

  test('Rich Result Viewer displays sleek error card on failure', async ({ page }) => {
    // Navigate to the tools page
    await page.goto('/tools');

    // Backend registration is async, retry until tools are loaded
    for (let i = 0; i < 10; i++) {
        try {
            await expect(page.getByText('fail_tool').first()).toBeVisible({ timeout: 5000 });
            break;
        } catch (e) {
            await page.reload();
            await page.waitForLoadState('networkidle');
            await page.waitForTimeout(2000);
        }
    }

    // Search for the failing tool
    await page.getByPlaceholder('Search tools...').fill('fail_tool');
    await expect(page.getByText('rich-result-test-service.fail_tool').first()).toBeVisible({ timeout: 10000 });

    // Open inspector
    await page.getByRole('row', { name: 'rich-result-test-service.fail_tool' }).getByRole('button', { name: 'Inspect' }).click();

    // Execute tool
    await page.getByRole('button', { name: 'Execute' }).click();

    // Verify the Error Card is displayed instead of raw JSON
    const errorTab = page.getByRole('tab', { name: 'Error' });
    await expect(errorTab).toBeVisible({ timeout: 10000 });
    await expect(errorTab).toHaveAttribute('data-state', 'active');

    // Check for elements within the Error Card
    await expect(page.getByRole('heading', { name: 'Execution Failed' })).toBeVisible();
    // Wait for the specific error text
    await expect(page.getByText(/Simulated internal error|An unknown error occurred/i)).toBeVisible();
  });

  test('Tool Inspector renders rich table result for complex data', async ({ page }) => {
    await page.goto('/tools');

    // Backend registration is async, retry until tools are loaded
    for (let i = 0; i < 10; i++) {
        try {
            await expect(page.getByText('get_complex_data').first()).toBeVisible({ timeout: 5000 });
            break;
        } catch (e) {
            await page.reload();
            await page.waitForLoadState('networkidle');
            await page.waitForTimeout(2000);
        }
    }

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
