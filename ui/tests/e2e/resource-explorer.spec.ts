/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Explorer', () => {

  const serviceName = 'e2e-resources-test-service';

  test.beforeEach(async ({ request }) => {
    // Seed the database with a test service that has resources
    const response = await request.post('/api/v1/services', {
      data: {
        name: "e2e-resources-service",
        priority: 10,
        command_line_service: {
            command: "echo",
            args: ['{"contents": [{"uri": "file:///config.json", "mimeType": "application/json", "text": "{\\"foo\\":\\"bar\\"}"}]}']
        },
        resources: [
            {
                uri: "file:///config.json",
                name: "config.json",
                description: "A config file",
                mimeType: "application/json"
            }
        ],
        calls: {
            "file:///config.json": {
                args: []
            }
        }
      }
    });
    // We expect the backend to create the service successfully
    expect(response.ok()).toBeTruthy();
  });

  test.afterEach(async ({ request }) => {
    await request.delete(`/api/v1/services/e2e-resources-service`);
  });

  test('should load resources and allow selection', async ({ page }) => {
    // Navigate to the resources page
    await page.goto('/resources');

    // Wait for the resource list to populate (using actual API)
    await expect(page.getByText('config.json').first()).toBeVisible();

    // Verify search functionality
    const searchInput = page.getByPlaceholder('Search resources...');
    await searchInput.fill('config');
    await expect(page.getByText('config.json').first()).toBeVisible();
    await searchInput.fill('not-found');
    await expect(page.getByText('config.json')).not.toBeVisible();

    // Clear search
    await searchInput.fill('');
    await expect(page.getByText('config.json').first()).toBeVisible();

    // Select a resource
    await page.getByText('config.json').first().click();

    // Verify preview loads
    // Use first() because the list item also shows the URI
    await expect(page.getByText('file:///config.json').first()).toBeVisible(); // URI header

    // Check if content area is visible. For JSON we expect the key 'foo' and value 'bar'
    await expect(page.getByText('foo').first()).toBeVisible();
    await expect(page.getByText('bar').first()).toBeVisible();

    // Verify toolbar buttons
    await expect(page.getByTitle('List View')).toBeVisible();
    await expect(page.getByTitle('Grid View')).toBeVisible();

    // Switch to Grid view
    await page.getByTitle('Grid View').click();

    // Verify grid item exists
    await expect(page.getByText('config.json').first()).toBeVisible();
  });
});
