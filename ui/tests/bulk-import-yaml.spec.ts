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

  test('should parse YAML correctly in import flow', async ({ page }) => {
    // 1. Open Dialog
    await page.getByRole('button', { name: 'Bulk Import' }).click();
    await expect(page.getByRole('heading', { name: 'Bulk Service Import' })).toBeVisible();

    // 2. Input Step (YAML)
    await expect(page.getByRole('tab', { name: 'JSON / YAML' })).toBeVisible();

    const validServiceYaml = `
- name: test-service-yaml-${Date.now()}
  commandLineService:
    command: echo hello
    workingDirectory: /tmp
`;

    await page.getByRole('textbox').fill(validServiceYaml);
    await page.getByRole('button', { name: 'Next: Review' }).click();

    // 3. Review Step
    // Wait for "Review Services" header
    await expect(page.getByRole('heading', { name: 'Review Services' })).toBeVisible();

    // Verify that the parser successfully found the service from YAML input
    await expect(page.getByText('Found 1 services')).toBeVisible();

    // Verify the service name appears in the table (further proof of correct parsing)
    await expect(page.getByRole('cell', { name: 'test-service-yaml-' })).toBeVisible();
  });
});
