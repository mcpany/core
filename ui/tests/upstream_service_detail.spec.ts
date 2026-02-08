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
        http_service: {
            address: "http://example.com"
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

    // 3. Verify ServiceEditor tabs are present (Evidence that ServiceEditor is used)
    // The old page had tabs: Overview, Config, Auth, Webhooks
    // The ServiceEditor has: General, Connection, Authentication, Policies, Advanced

    // Navigate to Settings tab where the editor is located
    await page.getByRole('tab', { name: 'Settings' }).click();

    await expect(page.getByRole('tab', { name: 'Connection' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Policies' })).toBeVisible();

    // 4. Modify a field
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
});
