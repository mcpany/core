/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

// Defines the output directory for screenshots relative to this file
// ui/tests/generate_docs_screenshots.spec.ts -> ../../docs/ui/screenshots
// inside docker, this will be /work/ui/tests/../../docs/ui/screenshots -> /work/docs/ui/screenshots
const DOCS_SCREENSHOTS_DIR = path.resolve(__dirname, '../../docs/ui/screenshots');

// Ensure directory exists
if (!fs.existsSync(DOCS_SCREENSHOTS_DIR)) {
  fs.mkdirSync(DOCS_SCREENSHOTS_DIR, { recursive: true });
}

test.describe('Generate Docs Screenshots and Verify UI', () => {
  const pages = [
    { name: 'dashboard', path: '/' },
    { name: 'services', path: '/services' },
    { name: 'tools', path: '/tools' },
    { name: 'resources', path: '/resources' },
    { name: 'prompts', path: '/prompts' },
    { name: 'profiles', path: '/profiles' },
    { name: 'middleware', path: '/middleware' },
    { name: 'webhooks', path: '/webhooks' },
    { name: 'network', path: '/network' },
    { name: 'logs', path: '/logs' },
    { name: 'playground', path: '/playground' },
    { name: 'stats', path: '/stats' },
    { name: 'stacks', path: '/stacks' },
  ];

  for (const pageInfo of pages) {
    test(`Verify and Screenshot ${pageInfo.name}`, async ({ page }) => {
      console.log(`Navigating to ${pageInfo.path}...`);
      const response = await page.goto(pageInfo.path);

      // Check for hard 404 status from server
      expect(response?.status(), `Page ${pageInfo.path} returned status ${response?.status()}`).toBe(200);

      // Wait for content - giving a bit more time for data fetching
      await page.waitForTimeout(3000);

      // simple visual check
      await expect(page.locator('body')).toBeVisible();

      // Verify no error toasts or alerts are visible (unless expected)
      const errorToasts = await page.locator('.toast-error').count();
      if (errorToasts > 0) {
        console.warn(`Warning: Found ${errorToasts} error toasts on page ${pageInfo.name}`);
      }

      // Take screenshot
      const screenshotPath = path.join(DOCS_SCREENSHOTS_DIR, `${pageInfo.name}.png`);
      await page.screenshot({ path: screenshotPath, fullPage: true });
      console.log(`Saved screenshot to ${screenshotPath}`);
    });
  }

  test('Verify and Screenshot Settings Tabs', async ({ page }) => {
      console.log('Navigating to /settings...');
      await page.goto('/settings');
      await expect(page.locator('body')).toBeVisible();
      await page.waitForTimeout(1000);

      // Default is Profiles
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_profiles.png'), fullPage: true });
      console.log('Saved screenshot to settings_profiles.png');

      // Click General
      console.log('Clicking General tab...');
      await page.getByRole('tab', { name: 'General' }).click();
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_general.png'), fullPage: true });
      console.log('Saved screenshot to settings_general.png');
  });
});
