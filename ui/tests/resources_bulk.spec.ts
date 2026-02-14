/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Bulk Actions', () => {
  const serviceName = 'e2e-fs-resources';

  test.beforeAll(async ({ request }) => {
    // Register a filesystem service
    // Note: We point to /app because that's where the server container has files (config.minimal.yaml)
    const response = await request.post('/api/v1/services', {
      data: {
        id: serviceName,
        name: serviceName,
        filesystem_service: {
          root_paths: ['/app']
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    // Unregister the service
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should allow bulk selection and actions', async ({ page }) => {
    await page.goto('/resources');

    // Wait for the resource to appear
    const resourceName = 'config.minimal.yaml';
    await expect(page.getByText(resourceName).first()).toBeVisible({ timeout: 15000 });

    // Checkbox interaction (TDD: this will fail until implemented)
    // We expect a checkbox for the item.
    // Using a reliable selector strategy, e.g., finding the row by text, then the checkbox within it.
    // Since I haven't implemented it, I'll assume I'll put a checkbox with accessible role.
    const resourceRow = page.locator('.group', { hasText: resourceName }).first();
    const checkbox = resourceRow.getByRole('checkbox');

    // Validate checkbox exists (will fail)
    await expect(checkbox).toBeVisible();
    await checkbox.check();

    // Verify "Bulk Actions" toolbar appears
    await expect(page.getByText('1 selected')).toBeVisible();

    // Verify "Copy Content" button
    const copyBtn = page.getByRole('button', { name: 'Copy Content' });
    await expect(copyBtn).toBeVisible();

    // Click it
    await copyBtn.click();

    // Verify Success Toast
    // The toast usually contains "Copied" or similar
    await expect(page.getByText('Content copied to clipboard')).toBeVisible();
  });
});
