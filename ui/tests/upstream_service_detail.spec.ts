/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { login } from './e2e/auth-helper';
import { seedUser, cleanupUser } from './e2e/test-data';

test.describe('Upstream Service Detail Page', () => {
  const serviceName = 'e2e-detail-test-service';

  test.beforeEach(async ({ page, request }) => {
    await seedUser(request, "e2e-admin");

    // Seed test service with authentication
    const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
    const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

    // Try clean up first
    await request.delete(`/api/v1/services/${serviceName}`, { headers: HEADERS }).catch(() => {});

    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        http_service: {
            address: "http://example.com"
        },
        priority: 10
      },
      headers: HEADERS
    });
    expect(response.ok()).toBeTruthy();

    await login(page);
  });

  test.afterEach(async ({ request }) => {
    const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
    const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };
    await request.delete(`/api/v1/services/${serviceName}`, { headers: HEADERS });
    await cleanupUser(request, "e2e-admin");
  });

  test('should display ServiceEditor and save changes', async ({ page, request }) => {
    // 1. Navigate to the detail page
    await page.goto(`/upstream-services/${serviceName}`);

    // 2. Verify Page Title
    await expect(page.getByRole('heading', { level: 1 })).toContainText(serviceName);

    // 3. Verify ServiceEditor tabs are present
    await expect(page.getByRole('tab', { name: 'Connection' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Policies' })).toBeVisible();

    // 4. Modify a field
    const priorityInput = page.getByLabel('Priority');
    await expect(priorityInput).toBeVisible();
    await expect(priorityInput).toHaveValue('10');

    await priorityInput.fill('5');

    // 5. Save Changes
    const saveButton = page.getByRole('button', { name: 'Save Changes' });
    await saveButton.click();

    // 6. Verify Toast/Feedback
    await expect(page.getByText('Service Updated').first()).toBeVisible();
    await expect(page.getByText('Configuration saved successfully').first()).toBeVisible();

    // 7. Verify Persistence via API (requires auth)
    const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
    const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };
    const response = await request.get(`/api/v1/services/${serviceName}`, { headers: HEADERS });
    expect(response.ok()).toBeTruthy();
    const service = await response.json();
    expect(service.priority).toBe(5);
  });
});
