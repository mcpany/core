/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Export', () => {
  test.beforeEach(async ({ page }) => {
    // Mock resources list
    await page.route('**/api/v1/resources', async (route) => {
      await route.fulfill({
        json: {
          resources: [
            {
              uri: 'file:///logs/app.log',
              name: 'Application Logs.txt',
              mimeType: 'text/plain',
              service: 'log-service'
            }
          ]
        }
      });
    });

    // Mock resource content
    await page.route('**/api/v1/resources/read*', async (route) => {
        // Return dummy content
        await route.fulfill({
            json: {
                contents: [
                    {
                        uri: 'file:///logs/app.log',
                        mimeType: 'text/plain',
                        text: 'Sample log content'
                    }
                ]
            }
        });
    });
  });

  test('should allow downloading a resource', async ({ page }) => {
    await page.goto('/resources');

    // Select the resource
    await page.getByText('Application Logs.txt').first().click();

    // Wait for content to load (Viewer shows content)
    // We look for text content
    await expect(page.getByText('Sample log content')).toBeVisible({ timeout: 10000 });

    // Click the Download button in header
    // The button has "Download" text (from Download icon + sr-only or similar? No, standard button in header)
    // In my code: <Button ...><Download .../> Download</Button>
    const downloadPromise = page.waitForEvent('download');
    await page.getByRole('button', { name: 'Download' }).click();
    const download = await downloadPromise;

    expect(download.suggestedFilename()).toBe('Application Logs.txt');
  });

  test('should allow downloading from context menu', async ({ page }) => {
      await page.goto('/resources');

      // Right click the resource
      await page.getByText('Application Logs.txt').first().click({ button: 'right' });

      // Click Download in context menu
      const downloadPromise = page.waitForEvent('download');
      await page.getByRole('menuitem', { name: 'Download' }).click();
      const download = await downloadPromise;

      expect(download.suggestedFilename()).toBe('Application Logs.txt');
  });
});
