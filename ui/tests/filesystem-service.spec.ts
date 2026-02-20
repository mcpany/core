/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Filesystem Service Configuration', () => {
  const serviceName = `e2e-fs-test-${Date.now()}`;

  test.afterAll(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});
  });

  test('should create a Filesystem service via UI', async ({ page, request }) => {
    // 1. Navigate to Services Page
    await page.goto('/upstream-services');

    // 2. Open Add Service Sheet
    await page.getByRole('button', { name: 'Add Service' }).click();

    // 3. Select "Custom Service" Template
    await expect(page.getByText('Custom Service')).toBeVisible();
    await page.getByText('Custom Service').click();

    await expect(page.getByLabel('Service Name')).toBeVisible();

    // 4. Fill Basic Info
    await page.getByLabel('Service Name').fill(serviceName);

    // 5. Select Filesystem Type
    await page.getByRole('tab', { name: 'Connection' }).click();
    // Assuming "Service Type" label is associated with the trigger
    await page.getByLabel('Service Type').click();
    await page.getByRole('option', { name: 'Filesystem' }).click();

    // 6. Verify Filesystem Config appears
    await expect(page.getByText('Backend Storage')).toBeVisible();
    await expect(page.getByText('Mount Points')).toBeVisible();

    // 7. Configure Mount Point
    await page.getByRole('button', { name: 'Add Mount Point' }).click();

    // We need to target the inputs. Since they are dynamically added, we use placeholders
    // or we can target by index if we are careful.
    // The placeholder for virtual is "/workspace" and for physical is dependent on type.
    // Default type is OS.
    await page.getByPlaceholder('/workspace').first().fill('/test-mount');

    // Check if type is OS (default)
    await expect(page.getByText('Local Filesystem (OS)')).toBeVisible();

    // Fill physical path
    await page.getByPlaceholder('/home/user/projects').first().fill('/tmp/test');

    // 8. Save
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // 9. Verify Toast
    await expect(page.getByText('Service Created').first()).toBeVisible();

    // 10. Verify via API
    const response = await request.get(`/api/v1/services/${serviceName}`);
    expect(response.ok()).toBeTruthy();
    const service = await response.json();

    expect(service.name).toBe(serviceName);
    // API returns snake_case
    expect(service.filesystem_service).toBeDefined();
    const roots = service.filesystem_service.root_paths;
    expect(roots['/test-mount']).toBe('/tmp/test');
  });
});
