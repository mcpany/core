/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Service Import Wizard', () => {
  test.beforeEach(async ({ page }) => {
    // Go to the upstream services page
    await page.goto('/upstream-services');
  });

  test('should complete import flow with valid JSON', async ({ page }) => {
    // 1. Open Dialog
    await page.getByRole('button', { name: 'Bulk Import' }).click();
    await expect(page.getByRole('heading', { name: 'Bulk Service Import' })).toBeVisible();

    // 2. Input Step (JSON)
    await expect(page.getByRole('tab', { name: 'JSON / YAML' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'File Upload' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'URL Import' })).toBeVisible();

    const validService = [
      {
        name: `test-service-${Date.now()}`,
        httpService: { address: 'https://example.com' }
      }
    ];
    const jsonString = JSON.stringify(validService);

    await page.getByRole('textbox').fill(jsonString);
    await page.getByRole('button', { name: 'Next: Review' }).click();

    // 3. Review Step
    // Wait for "Review Services" header
    await expect(page.getByRole('heading', { name: 'Review Services' })).toBeVisible();

    // Wait for validation to complete (loader disappears, table populates)
    // We expect 1 valid service
    await expect(page.getByText('Found 1 services. 1 valid')).toBeVisible();

    // Verify service name is in table
    await expect(page.getByRole('cell', { name: validService[0].name })).toBeVisible();

    // 4. Import Step
    await page.getByRole('button', { name: 'Import 1 Services' }).click();

    // 5. Success
    await expect(page.getByRole('heading', { name: 'Import Complete' })).toBeVisible();
    await expect(page.getByRole('dialog').getByText('Successfully imported 1 services.')).toBeVisible();

    // Close
    // Note: There are two "Close" buttons (Dialog X and Wizard Close). The Wizard one is usually first in DOM order or we pick first.
    await page.getByRole('button', { name: 'Close' }).first().click();

    // Verify dialog closed
    await expect(page.getByRole('heading', { name: 'Bulk Service Import' })).not.toBeVisible();

    // Verify service appears in list (might need refresh or wait)
    // The wizard calls onImportSuccess which triggers fetchServices in parent
    await expect(page.getByRole('link', { name: validService[0].name })).toBeVisible();
  });

  test('should complete import flow with valid YAML', async ({ page }) => {
    // 1. Open Dialog
    await page.getByRole('button', { name: 'Bulk Import' }).click();
    await expect(page.getByRole('heading', { name: 'Bulk Service Import' })).toBeVisible();

    // 2. Input Step (YAML)
    const validService = [
        {
          name: `test-service-yaml-${Date.now()}`,
          httpService: { address: 'https://example.com/yaml' }
        }
    ];
    // Simple YAML string construction
    const yamlString = `- name: ${validService[0].name}\n  httpService:\n    address: ${validService[0].httpService.address}`;

    await page.getByRole('textbox').fill(yamlString);
    await page.getByRole('button', { name: 'Next: Review' }).click();

    // 3. Review Step
    await expect(page.getByRole('heading', { name: 'Review Services' })).toBeVisible();

    // Wait for validation to complete
    await expect(page.getByText('Found 1 services. 1 valid')).toBeVisible();

    // Verify service name is in table
    await expect(page.getByRole('cell', { name: validService[0].name })).toBeVisible();

    // 4. Import Step
    await page.getByRole('button', { name: 'Import 1 Services' }).click();

    // 5. Success
    await expect(page.getByRole('heading', { name: 'Import Complete' })).toBeVisible();
    await expect(page.getByRole('dialog').getByText('Successfully imported 1 services.')).toBeVisible();

    // Close
    await page.getByRole('button', { name: 'Close' }).first().click();

    // Verify dialog closed
    await expect(page.getByRole('heading', { name: 'Bulk Service Import' })).not.toBeVisible();

    // Verify service appears in list
    await expect(page.getByRole('link', { name: validService[0].name })).toBeVisible();
  });
});
