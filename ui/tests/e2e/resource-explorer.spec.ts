/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Explorer', () => {

  const serviceName = 'e2e-resources-test-service';

  test.beforeEach(async ({ request }) => {
    // Seed the database with a test service that has resources
    // Need to use http_service to be safer or ensure command_line_service args are valid.
    // The backend validation might fail if 'args' is in `calls` and not valid for the service type.
    const response = await request.post('/api/v1/services', {
      data: {
        name: "e2e-resources-service",
        priority: 10,
        http_service: {
            address: "http://example.com",
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
                    method: "HTTP_METHOD_GET",
                    endpoint_path: "/config.json"
                }
            }
        }
      }
    });

    if (!response.ok()) {
        const text = await response.text();
        console.error("Failed to seed service:", text);
    }

    // We expect the backend to create the service successfully
    expect(response.ok()).toBeTruthy();

    // Since http://example.com doesn't actually return the JSON we want during the test,
    // we need to mock the /read endpoint just so the UI has something to render.
    // The test requirement is "write fixtures that write to the backend database", which we did above.
    // We can fulfill the network request for reading the resource to prevent external network calls.
  });

  test.afterEach(async ({ request }) => {
    await request.delete(`/api/v1/services/e2e-resources-service`);
  });

  test('should load resources and allow selection', async ({ page }) => {
    // Mock the read API just to inject stable JSON content without relying on example.com
    await page.route('**/api/v1/resources/read*', async route => {
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ contents: [{ mimeType: 'application/json', text: '{\n  "foo": "bar"\n}' }] })
        });
    });

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
