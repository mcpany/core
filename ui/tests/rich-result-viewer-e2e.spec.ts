/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Rich Result Viewer Formatted Table E2E', () => {
  const serviceName = 'rich-result-table-e2e-service';

  test.beforeAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    // Seed a mock service that returns JSON stringified output to simulate real world tool
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        command_line_service: {
          command: 'echo',
          tools: [
            {
              name: 'get_users',
              description: 'Returns a list of users',
              call_id: 'get_users'
            }
          ],
          calls: {
            'get_users': {
              args: [
                JSON.stringify([
                  { id: 1, name: 'Alice', active: true },
                  { id: 2, name: 'Bob', active: false }
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
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});
  });

  test('should render formatted table instead of raw JSON', async ({ page }) => {
    await page.goto('/tools');

    await page.getByPlaceholder('Search tools...').fill('get_users');
    await expect(page.getByText(`${serviceName}.get_users`).first()).toBeVisible({ timeout: 10000 });

    await page.getByRole('row', { name: `${serviceName}.get_users` }).getByRole('button', { name: 'Inspect' }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    await page.getByRole('button', { name: 'Execute' }).click();

    await expect(page.locator('label').filter({ hasText: 'Result' })).toBeVisible({ timeout: 10000 });

    // Ensure it's not a raw JSON dump by checking for the Table tab
    const tableTab = page.getByRole('tab', { name: 'Table' });
    await expect(tableTab).toBeVisible();

    const table = page.getByRole('table');
    await expect(table).toBeVisible();

    // Verify parsed formatting
    await expect(table.getByText('Alice')).toBeVisible();
    await expect(table.getByText('true')).toBeVisible();
    await expect(table.getByText('Bob')).toBeVisible();
    await expect(table.getByText('false')).toBeVisible();
  });
});
