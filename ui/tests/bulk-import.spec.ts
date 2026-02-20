/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';

test.describe('Bulk Import', () => {
  const yamlPath = path.join(__dirname, 'test-services.yaml');
  const claudePath = path.join(__dirname, 'test-claude.json');

  test.beforeAll(() => {
    // Create test files
    fs.writeFileSync(yamlPath, `
services:
  - name: yaml-service
    httpService:
      address: http://localhost:8080
`);

    fs.writeFileSync(claudePath, JSON.stringify({
      mcpServers: {
        "claude-service": {
          "command": "echo",
          "args": ["hello"]
        }
      }
    }));
  });

  test.afterAll(() => {
    // Cleanup
    if (fs.existsSync(yamlPath)) fs.unlinkSync(yamlPath);
    if (fs.existsSync(claudePath)) fs.unlinkSync(claudePath);
  });

  test('should import YAML configuration', async ({ page }) => {
    await page.goto('/upstream-services');
    await page.getByRole('button', { name: 'Bulk Import' }).click();

    // Select File Upload tab
    await page.getByRole('tab', { name: 'File Upload' }).click();

    // Upload YAML
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(yamlPath);

    // Click Next
    await page.getByRole('button', { name: 'Next: Review' }).click();

    // Check if service is listed
    await expect(page.getByText('yaml-service')).toBeVisible();
    // It might show as HTTP
    const row = page.getByRole('row').filter({ hasText: 'yaml-service' });
    await expect(row.getByText('HTTP', { exact: true })).toBeVisible();
  });

  test('should import Claude Desktop configuration', async ({ page }) => {
    await page.goto('/upstream-services');
    await page.getByRole('button', { name: 'Bulk Import' }).click();

    // Select File Upload tab
    await page.getByRole('tab', { name: 'File Upload' }).click();

    // Upload Claude JSON
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(claudePath);

    // Click Next
    await page.getByRole('button', { name: 'Next: Review' }).click();

    // Check if service is listed
    await expect(page.getByText('claude-service')).toBeVisible();
    // It should be CLI type
    const row = page.getByRole('row').filter({ hasText: 'claude-service' });
    await expect(row.getByText('CLI', { exact: true })).toBeVisible();
  });
});
