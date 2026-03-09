/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Smart MCP Content Data Grid', () => {
  const serviceName = 'database-test-service';

  test.beforeAll(async ({ request }) => {
    // Delete existing
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => { });

    // Ensure our seed from config is there or seed it again
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        command_line_service: {
          command: 'echo',
          tools: [
            { name: 'get_database_users', call_id: 'call1', description: 'Fetch DB users' }
          ],
          calls: {
            'call1': {
              args: [
                  JSON.stringify({
                      content: [{
                          type: "text",
                          text: JSON.stringify([
                              {id: 101, username: "admin_jane", role: "SuperAdmin", last_login: "2023-10-01"},
                              {id: 102, username: "dev_bob", role: "Developer", last_login: "2023-10-02"},
                              {id: 103, username: "support_alice", role: "Support", last_login: "2023-10-03"}
                          ])
                      }],
                      isError: false
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

  test('MCP text block containing a JSON array of objects should be rendered as a Table', async ({ page }) => {
    await page.goto('/tools');

    // Search for the test tool
    await page.getByPlaceholder('Search tools...').fill('get_database_users');
    await expect(page.getByText('database-test-service.get_database_users').first()).toBeVisible({ timeout: 10000 });

    // Open inspector
    await page.getByRole('row', { name: 'database-test-service.get_database_users' }).getByRole('button', { name: 'Inspect' }).click();

    // Wait for inspector to open
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // Execute tool
    await page.getByRole('button', { name: 'Execute' }).click();

    // Wait for result
    await expect(page.locator('label').filter({ hasText: 'Result' })).toBeVisible({ timeout: 10000 });

    // The 'Rendered' tab should be active by default for MCP Content
    // And it should have a table!
    const renderedTabContent = page.getByRole('tabpanel', { name: 'Rendered' });

    // There should be a table header with columns 'id', 'username', 'role', 'last_login'
    await expect(renderedTabContent.getByRole('columnheader', { name: 'id' })).toBeVisible();
    await expect(renderedTabContent.getByRole('columnheader', { name: 'username' })).toBeVisible();
    await expect(renderedTabContent.getByRole('columnheader', { name: 'role' })).toBeVisible();

    // There should be rows with data
    await expect(renderedTabContent.getByRole('cell', { name: 'admin_jane' })).toBeVisible();
    await expect(renderedTabContent.getByRole('cell', { name: 'SuperAdmin' })).toBeVisible();
    await expect(renderedTabContent.getByRole('cell', { name: 'dev_bob' })).toBeVisible();

    // The table footer should show '3 records'
    await expect(renderedTabContent.getByText('3 records')).toBeVisible();
  });
});
