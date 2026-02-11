/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Service Config Viewer', () => {
  const serviceName = 'e2e-config-viewer-test';

  test.beforeAll(async ({ request }) => {
    // Seed the database with a test service containing various config types
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        command_line_service: {
            command: "python script.py",
            working_directory: "/app",
            env: {
                "API_KEY": "secret-api-key-123",
                "DEBUG": "true"
            }
        },
        upstream_auth: {
            api_key: {
                in: 0, // Header
                param_name: "X-My-Key",
                value: { plain_text: "super-secret-key" }
            }
        },
        priority: 5,
        tags: ["test", "e2e"]
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should display service configuration correctly', async ({ page }) => {
    // 1. Navigate to the detail page
    await page.goto(`/upstream-services/${serviceName}`);

    // 2. Verify Page Title
    await expect(page.getByRole('heading', { level: 1 })).toContainText(serviceName);

    // 3. Click on "Configuration" tab
    await page.getByRole('tab', { name: 'Configuration' }).click();

    // 4. Verify Identity Section
    await expect(page.getByText('Service Identity')).toBeVisible();
    await expect(page.getByText(serviceName)).toBeVisible();
    await expect(page.getByText('Priority')).toBeVisible();
    await expect(page.getByText('5', { exact: true })).toBeVisible();
    await expect(page.getByText('test')).toBeVisible(); // Tag

    // 5. Verify Connection Details
    await expect(page.getByText('Connection Details')).toBeVisible();
    await expect(page.getByText('Command Line')).toBeVisible(); // Service Type
    await expect(page.getByText('python script.py')).toBeVisible();
    await expect(page.getByText('/app')).toBeVisible();

    // 6. Verify Authentication
    await expect(page.getByText('Authentication')).toBeVisible();
    await expect(page.getByText('API Key', { exact: true })).toBeVisible(); // Type badge
    await expect(page.getByText('X-My-Key')).toBeVisible(); // Param Name
    await expect(page.getByText('Header', { exact: true })).toBeVisible(); // Location
    // Verify secret masking
    await expect(page.getByText('••••••••••••••••').first()).toBeVisible();

    // 7. Verify Environment Variables
    await expect(page.getByText('Environment Variables')).toBeVisible();
    await expect(page.getByText('API_KEY')).toBeVisible();
    await expect(page.getByText('DEBUG')).toBeVisible();
    // Verify secret masking for env vars
    // We expect at least one more masked value
    await expect(page.getByText('••••••••••••••••').nth(1)).toBeVisible();

    // 8. Verify Edit tab still exists
    await expect(page.getByRole('tab', { name: 'Edit' })).toBeVisible();
  });
});
