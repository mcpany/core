/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Schema Viewer Table', () => {
  test('should display schema as a formatted table', async ({ page, request }) => {
    // 1. Database Seeding: Inject a test service with a tool schema
    const serviceConfig = {
      id: "schema-test-service",
      name: "schema-test-service",
      version: "1.0.0",
      openapi_service: {
        address: "http://localhost:8080",
        tools: [
          {
            name: "test_schema_tool",
            description: "A tool to test the schema viewer table",
            input_schema: {
              type: "object",
              properties: {
                user_id: {
                  type: "string",
                  description: "The ID of the user"
                },
                tags: {
                  type: "array",
                  items: {
                    type: "string"
                  },
                  description: "List of tags"
                }
              },
              required: ["user_id"]
            }
          }
        ]
      }
    };

    const res = await request.post('/api/v1/services', {
      data: serviceConfig
    });
    expect(res.ok()).toBeTruthy();

    // 2. Navigate to Tools page
    await page.goto('/tools');

    // 3. Find the tool row
    const toolRow = page.locator('tr').filter({ hasText: 'test_schema_tool' });
    await expect(toolRow).toBeVisible({ timeout: 10000 });

    // 4. Click Inspect
    await toolRow.getByRole('button', { name: 'Inspect' }).click();

    // 5. Navigate to Schema tab
    const schemaTab = page.getByRole('tab', { name: 'Schema' });
    await expect(schemaTab).toBeVisible();
    await schemaTab.click();

    // 6. Verify Table headers
    const tableHeader = page.locator('thead');
    await expect(tableHeader).toContainText('Property');
    await expect(tableHeader).toContainText('Type');
    await expect(tableHeader).toContainText('Required');
    await expect(tableHeader).toContainText('Description');

    // 7. Verify Table contents (seeded data)
    const userRow = page.locator('tr').filter({ hasText: 'user_id' });
    await expect(userRow).toBeVisible();
    await expect(userRow).toContainText('string');
    await expect(userRow).toContainText('Yes'); // Required
    await expect(userRow).toContainText('The ID of the user');

    const tagsRow = page.locator('tr').filter({ hasText: 'tags' });
    await expect(tagsRow).toBeVisible();
    await expect(tagsRow).toContainText('array');
    await expect(tagsRow).toContainText('No'); // Not Required
    await expect(tagsRow).toContainText('List of tags');

    // 8. Cleanup (unregister the service)
    await request.delete('/api/v1/services/schema-test-service');
  });
});
