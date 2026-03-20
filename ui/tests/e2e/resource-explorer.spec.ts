/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Explorer', () => {

  const serviceName = 'e2e-resources-test-service';

  test.beforeEach(async ({ request }) => {
    // Seed the database with a test service that uses echo command to return JSON
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        priority: 10,
        command_line_service: {
            command: "echo",
            // The command returns a valid JSON string that matches the expected ResourceContent structure for MCP
            // Wait, MCP Resource reading returns an array of contents. The echo command output will be treated as the raw output.
            // Let's use a simple JSON string output. The MCP adapter for Command Line takes stdout.
            // But wait, the CLI adapter expects to return a proper MCP result or we configure an output transformer.
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
                // Command line call definition
                // If it's a CLI service, the call args will be appended to the command.
                args: []
            }
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterEach(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should load resources and allow selection', async ({ page }) => {
    // Navigate to the resources page
    await page.goto('/resources');

    // Wait for the resource list to populate
    await expect(page.getByText('config.json').first()).toBeVisible({ timeout: 10000 });

    // Select the resource
    await page.getByText('config.json').first().click();

    // Verify preview loads and JSON is rendered using JsonView
    // Since JsonView parses the JSON into a tree structure, we should see the key "foo" and value "bar"
    // separated, not just a raw string like "{\\"foo\\":\\"bar\\"}".
    // JsonView renders keys with a specific styling, but we can just check for text content.

    // Wait for the resource viewer to load the content
    await expect(page.getByText('foo')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('bar')).toBeVisible();

    // Also verify the URI header is visible
    await expect(page.getByText('file:///config.json').first()).toBeVisible();

    // Switch to Grid view
    await page.getByTitle('Grid View').click();
    await expect(page.getByText('config.json').first()).toBeVisible();
  });
});
