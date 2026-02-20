/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Filesystem Service Configuration', () => {
  const serviceName = 'e2e-fs-test-service-' + Math.random().toString(36).substring(7);

  test.afterAll(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should create a new Filesystem service via UI', async ({ page }) => {
    // 1. Navigate to Services Page
    await page.goto('/upstream-services');

    // 2. Click Add Service
    await page.getByRole('button', { name: 'Add Service' }).click();

    // 3. Select "Custom Service" template
    // The template card has "Custom Service" text.
    await page.getByText('Custom Service').click();

    // 4. Fill Basic Info
    await page.getByLabel('Service Name').fill(serviceName);
    await page.getByLabel('Version').fill('1.0.0');

    // 5. Go to Connection Tab
    await page.getByRole('tab', { name: 'Connection' }).click();

    // 6. Select Filesystem Type
    await page.getByLabel('Service Type').click();
    await page.getByRole('option', { name: 'Local / Remote Filesystem' }).click();

    // 7. Verify Filesystem Config UI appears
    await expect(page.getByRole('tab', { name: 'Mounts & General' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Backend Storage' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Security' })).toBeVisible();

    // 8. Add a Mount Point
    // Click "Add Path"
    await page.getByRole('button', { name: 'Add Path' }).click();

    // Fill Key/Value (Virtual/Physical)
    const keyInputs = await page.getByPlaceholder('Key').all();
    const valInputs = await page.getByPlaceholder('Value').all();

    await keyInputs[0].fill('/virtual/data');
    await valInputs[0].fill('/tmp/data');

    // 9. Go to Backend Storage and verify OS is default
    await page.getByRole('tab', { name: 'Backend Storage' }).click();
    await expect(page.getByText('Uses the local server filesystem')).toBeVisible();

    // 10. Save
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // 11. Verify Success Toast
    await expect(page.getByText('Service Created').first()).toBeVisible();

    // 12. Verify in List
    // Wait for list to reload
    await expect(page.getByRole('cell', { name: serviceName })).toBeVisible();

    // Verify badge "FS"
    await expect(page.getByText('FS', { exact: true })).toBeVisible();

    // Verify mount count "1 mount"
    await expect(page.getByText('1 mount')).toBeVisible();
  });
});
