/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, request } from '@playwright/test';

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const HEADERS = { 'X-API-Key': API_KEY };

test.describe('Rich Result Viewer', () => {
  const serviceName = 'complex-data-service';

  test.beforeAll(async () => {
    const context = await request.newContext({ baseURL: BASE_URL });

    // Seed the service dynamically using the API
    const serviceConfig = {
      id: serviceName,
      name: serviceName,
      version: "1.0.0",
      command_line_service: {
        command: "echo",
        tools: [
          {
            name: "get_complex_data",
            description: "Returns complex data for UI testing",
            call_id: "get_complex_data",
            input_schema: {
              type: "object",
              properties: {}
            }
          }
        ],
        calls: {
          get_complex_data: {
            // Echo back a JSON array of objects
            args: ['[{"id": 1, "name": "Alice", "role": "Admin", "details": {"active": true}}, {"id": 2, "name": "Bob", "role": "User", "details": {"active": false}}]']
          }
        }
      }
    };

    const res = await context.post('/api/v1/services', {
      data: serviceConfig,
      headers: HEADERS
    });

    if (!res.ok()) {
        console.error('Failed to seed service:', await res.text());
    }
    expect(res.ok()).toBeTruthy();
  });

  test.afterAll(async () => {
    const context = await request.newContext({ baseURL: BASE_URL });
    await context.delete(`/api/v1/services/${serviceName}`, { headers: HEADERS });
  });

  test('RichResultViewer displays complex data in table view', async ({ page }) => {
    // Go to tools page
    await page.goto('/tools');

    // Filter by our test service/tool
    await page.getByPlaceholder('Search tools...').fill('get_complex_data');

    // Open inspector
    await page.locator('tr').filter({ hasText: 'get_complex_data' }).getByText('Inspect').click();

    // Wait for inspector
    await expect(page.getByText('get_complex_data').first()).toBeVisible();

    // Click Execute
    await page.getByRole('button', { name: 'Execute' }).click();

    // Wait for Result label to appear
    await expect(page.getByText('Result', { exact: true })).toBeVisible();

    // The RichResultViewer should show tabs: "Result" and "Full Output" if parsing succeeded
    await expect(page.getByRole('tab', { name: 'Result' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Full Output' })).toBeVisible();

    // By default "Result" tab is selected.
    // Inside it, JsonView should detect table data and show "Table" button/view.

    // Wait for Table button in the toolbar
    // Note: If the table view is auto-selected (which it is for array of objects), the button might look different (secondary variant)
    await expect(page.getByRole('button', { name: 'Table' })).toBeVisible();

    // Click Table button (it might be auto-selected, but clicking ensures)
    await page.getByRole('button', { name: 'Table' }).click();

    // Check for data in the table
    await expect(page.getByRole('cell', { name: 'Alice' })).toBeVisible();
    await expect(page.getByRole('cell', { name: 'Bob' })).toBeVisible();
    await expect(page.getByRole('cell', { name: 'Admin' })).toBeVisible();
    await expect(page.getByRole('cell', { name: 'User' })).toBeVisible();
  });
});
