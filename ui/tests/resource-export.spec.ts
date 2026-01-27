/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Export', () => {
  const testFileName = `test-resource-${Date.now()}.txt`;
  const testFileContent = 'This is a test resource for export.';
  const serviceName = `test-fs-service-${Date.now()}`;
  const tempDir = '/tmp/mcp-test-resources';

  test.beforeAll(async ({ request }) => {
    const fs = require('fs');
    const path = require('path');

    if (!fs.existsSync(tempDir)) {
      fs.mkdirSync(tempDir);
    }
    fs.writeFileSync(path.join(tempDir, testFileName), testFileContent);

    // Register Filesystem Service via API with explicit resource definition
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        filesystem_service: {
          root_paths: {
            "/test-resources": tempDir
          },
          read_only: true,
          // Explicitly define the resource to ensure it is listed
          resources: [
            {
              name: testFileName,
              uri: `file:///test-resources/${testFileName}`,
              mime_type: "text/plain"
            }
          ]
        }
      }
    });

    if (!response.ok()) {
        console.error('Failed to register service:', await response.text());
    }
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should allow downloading a resource', async ({ page }) => {
    await page.goto('/resources');

    // Robust wait for resource to appear (polling/reload)
    let found = false;
    for (let i = 0; i < 5; i++) {
        try {
            await expect(page.getByText(testFileName)).toBeVisible({ timeout: 2000 });
            found = true;
            break;
        } catch (e) {
            await page.reload();
            await page.waitForLoadState('networkidle');
        }
    }
    expect(found, "Resource not found after retries").toBeTruthy();

    // Select it
    await page.getByText(testFileName).click();

    // Verify content loaded
    await expect(page.getByText(testFileContent)).toBeVisible({ timeout: 10000 });

    // Test Download
    const downloadPromise = page.waitForEvent('download');
    await page.getByRole('button', { name: 'Download' }).click();
    const download = await downloadPromise;

    expect(download.suggestedFilename()).toBe(testFileName);
  });
});
