/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Upstream Service Detail Page', () => {
  const serviceName = 'e2e-detail-test-service';

  test.beforeAll(async ({ request }) => {
    // Seed the database with a test service
    // Note: We must use /api/v1/services because that's what the middleware proxies
    // and what the backend exposes (mounted at /api/v1/).
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        command_line_service: {
            command: "echo"
        },
        priority: 10
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should display ServiceEditor and save changes', async ({ page, request }) => {
    // 1. Navigate to the detail page
    await page.goto(`/upstream-services/${serviceName}`);

    // 2. Verify Page Title
    await expect(page.getByRole('heading', { level: 1 })).toContainText(serviceName);

    // 3. Navigate to Settings tab where ServiceEditor is located
    await page.getByRole('tab', { name: 'Settings' }).click();

    // 4. Verify ServiceEditor tabs are present (Evidence that ServiceEditor is used)
    // The ServiceEditor has: General, Connection, Authentication, Policies, Advanced
    await expect(page.getByRole('tab', { name: 'Connection' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Policies' })).toBeVisible();

    // 5. Modify a field
    // Go to General tab (default) and change Priority
    // Note: ServiceEditor defaults to "general" tab.
    const priorityInput = page.getByLabel('Priority');
    await expect(priorityInput).toBeVisible();
    await expect(priorityInput).toHaveValue('10');

    await priorityInput.fill('5');

    // 5. Save Changes
    const saveButton = page.getByRole('button', { name: 'Save Changes' });
    await saveButton.click();

    // 6. Verify Toast/Feedback
    // Use first() to avoid strict mode violation if multiple elements match (e.g. title and aria-live region)
    await expect(page.getByText('Service Updated').first()).toBeVisible();
    await expect(page.getByText('Configuration saved successfully').first()).toBeVisible();

    // 7. Verify Persistence via API
    const response = await request.get(`/api/v1/services/${serviceName}`);
    expect(response.ok()).toBeTruthy();
    const service = await response.json();
    expect(service.priority).toBe(5);
  });

  test('should display empty Tools definitions', async ({ page }) => {
    // Navigate to the DefinitionsTable page
    await page.goto(`/service/${serviceName}`);

    // Wait for the General tab to be active (default) and Tools table to appear
    await expect(page.getByText('Tools', { exact: true }).first()).toBeVisible();

    // The search input is only visible when data is present (if data length > 0 in component).
    await expect(page.getByText('No tools configured')).toBeVisible();
    await expect(page.getByText('This service has not exposed any tools.')).toBeVisible();
  });

  test('should display Tools definitions and filter them', async ({ page, request }) => {
    // Seed a specific test service so we don't depend on the server running with a specific config file
    // that might be missing or overridden in test environments.
    const serviceNameSearchTest = 'search-test-service';
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceNameSearchTest,
        command_line_service: {
            command: "echo",
            tools: [
                { name: "tool-alpha", description: "First tool" },
                { name: "tool-beta", description: "Second tool" }
            ]
        },
        priority: 10
      }
    });
    expect(response.ok()).toBeTruthy();

    // Navigate to the DefinitionsTable page
    await page.goto(`/service/${serviceNameSearchTest}`);

    // Wait for the General tab to be active (default) and Tools table to appear
    await expect(page.getByText('Tools', { exact: true }).first()).toBeVisible();

    // Verify both tools are initially visible
    await expect(page.getByText('tool-alpha')).toBeVisible();
    await expect(page.getByText('tool-beta')).toBeVisible();

    // Search for "alpha"
    const searchInput = page.getByPlaceholder('Search tools...');
    await expect(searchInput).toBeVisible();
    await searchInput.fill('alpha');

    // Verify "tool-alpha" is visible and "tool-beta" is hidden
    await expect(page.getByText('tool-alpha')).toBeVisible();
    await expect(page.getByText('tool-beta')).toBeHidden();

    // Clear search and verify empty state for a non-existent tool
    await searchInput.fill('nonexistent-tool');
    await expect(page.getByText('No results found')).toBeVisible();

    // Cleanup
    await request.delete(`/api/v1/services/${serviceNameSearchTest}`);
  });
});
