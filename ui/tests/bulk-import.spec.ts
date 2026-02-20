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

    const validService = [
      {
        name: `test-json-${Date.now()}`,
        httpService: { address: 'https://example.com' }
      }
    ];
    const jsonString = JSON.stringify(validService);

    await page.getByRole('textbox').fill(jsonString);
    await page.getByRole('button', { name: 'Next: Review' }).click();

    // 3. Review Step
    await expect(page.getByRole('heading', { name: 'Review Services' })).toBeVisible();
    await expect(page.getByText('Found 1 services. 1 valid')).toBeVisible();
    await expect(page.getByRole('cell', { name: validService[0].name })).toBeVisible();

    // 4. Import Step
    await page.getByRole('button', { name: 'Import 1 Services' }).click();

    // 5. Success
    await expect(page.getByRole('heading', { name: 'Import Complete' })).toBeVisible();
    await expect(page.getByRole('dialog').getByText('Successfully imported 1 services.')).toBeVisible();
  });

  test('should complete import flow with valid YAML', async ({ page }) => {
    // 1. Open Dialog
    await page.getByRole('button', { name: 'Bulk Import' }).click();

    const serviceName = `test-yaml-${Date.now()}`;
    const yamlString = `
- name: ${serviceName}
  httpService:
    address: https://example.com
`;

    // 2. Input Step (YAML)
    await page.getByRole('textbox').fill(yamlString);

    // Switch to JSON/YAML tab if not active (default)
    await page.getByRole('button', { name: 'Next: Review' }).click();

    // 3. Review Step
    await expect(page.getByRole('heading', { name: 'Review Services' })).toBeVisible();

    // This assertion will fail until YAML support is implemented
    await expect(page.getByText('Found 1 services. 1 valid')).toBeVisible();
    await expect(page.getByRole('cell', { name: serviceName })).toBeVisible();

    // 4. Import
    await page.getByRole('button', { name: 'Import 1 Services' }).click();
    await expect(page.getByRole('heading', { name: 'Import Complete' })).toBeVisible();
  });

  test('should complete import flow with Claude Desktop Config', async ({ page }) => {
    // 1. Open Dialog
    await page.getByRole('button', { name: 'Bulk Import' }).click();

    const serviceName = `test-claude-${Date.now()}`;
    const claudeConfig = {
      mcpServers: {
        [serviceName]: {
          url: "https://example.com"
        }
      }
    };

    // 2. Input Step
    await page.getByRole('textbox').fill(JSON.stringify(claudeConfig));
    await page.getByRole('button', { name: 'Next: Review' }).click();

    // 3. Review Step
    await expect(page.getByRole('heading', { name: 'Review Services' })).toBeVisible();

    // Verify format detection
    await expect(page.getByText('JSON (Claude Desktop)')).toBeVisible();

    // Verify parsing success (validation might fail due to fake URL)
    await expect(page.getByText(/Found 1 services/)).toBeVisible();
    await expect(page.getByRole('cell', { name: serviceName })).toBeVisible();

    // NOTE: validation fails for fake URL in MCP mode, so we can't proceed to import.
    // The test confirms that parsing works and the UI correctly identified the format.

    // Close
    await page.getByRole('button', { name: 'Cancel' }).click();
  });
});
