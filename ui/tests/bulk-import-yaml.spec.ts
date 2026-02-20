/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Service Import Wizard (YAML)', () => {
  test.beforeEach(async ({ page }) => {
    // Go to the upstream services page
    await page.goto('/upstream-services');
  });

  test('should complete import flow with valid YAML', async ({ page }) => {
    // 1. Open Dialog
    await page.getByRole('button', { name: 'Bulk Import' }).click();
    await expect(page.getByRole('heading', { name: 'Bulk Service Import' })).toBeVisible();

    // 2. Input Step (YAML)
    await expect(page.getByRole('tab', { name: 'JSON / YAML' })).toBeVisible();

    // Create valid YAML content
    const serviceName = `test-service-yaml-${Date.now()}`;
    const yamlString = `
- name: ${serviceName}
  httpService:
    address: https://example.com
    tools: []
    resources: []
    prompts: []
`;

    // Fill the textarea with YAML
    await page.getByRole('textbox').fill(yamlString);

    // Click Next
    await page.getByRole('button', { name: 'Next: Review' }).click();

    // 3. Review Step
    // Wait for "Review Services" header
    // This is where it should currently FAIL because the code uses JSON.parse() and throws an error
    await expect(page.getByRole('heading', { name: 'Review Services' })).toBeVisible({ timeout: 5000 });

    // Wait for validation to complete
    await expect(page.getByText('Found 1 services. 1 valid')).toBeVisible();

    // Verify service name is in table
    await expect(page.getByRole('cell', { name: serviceName })).toBeVisible();

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
    await expect(page.getByRole('link', { name: serviceName })).toBeVisible();
  });
});
