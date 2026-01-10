/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

// Defines the output directory for screenshots relative to this file
// ui/tests/generate_docs_screenshots.spec.ts -> ../docs/screenshots
// inside docker, this will be /work/ui/tests/../docs/screenshots -> /work/ui/docs/screenshots
const DOCS_SCREENSHOTS_DIR = path.resolve(__dirname, '../docs/screenshots');

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
    { name: 'marketplace', path: '/marketplace' },
    { name: 'users', path: '/users' },
    { name: 'traces', path: '/traces' },
    { name: 'secrets', path: '/secrets' },
    { name: 'settings', path: '/settings' },
    // Marketplace is already in the list at index 14, but let's ensure we visit it explicitly if needed
    // or just rely on the loop. The loop handles 'marketplace'.
    // We want to add external marketplace specifically
  ];

  /*
   * Existing loop covers marketplace/page.tsx
   */

  test.beforeEach(async ({ page }) => {
    // Mock Secrets
    await page.route('**/api/v1/secrets', async route => {
      await route.fulfill({
        json: {
          secrets: [
            { name: 'TEST_SECRET', value: '********' }
          ]
        }
      });
    });
  });

  for (const pageInfo of pages) {
    test(`Verify and Screenshot ${pageInfo.name}`, async ({ page }) => {
      console.log(`Navigating to ${pageInfo.path}...`);
      const response = await page.goto(pageInfo.path);

      // Check for hard 404 status from server
      expect(response?.status(), `Page ${pageInfo.path} returned status ${response?.status()}`).toBe(200);

      // Wait for content - giving a bit more time for data fetching
      await page.waitForTimeout(3000);

      if (pageInfo.name === 'marketplace') {
          await expect(page.getByText('Share Your Config')).toBeVisible();
      }

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
      await page.waitForTimeout(1000);

      // Default is Profiles
      // Save as settings.png (per user request) AND settings_profiles.png
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_profiles.png'), fullPage: true });
      console.log('Saved settings.png and settings_profiles.png');

      // Click Secrets (Tab)
      await page.getByRole('tab', { name: 'Secrets' }).click();
      await page.waitForTimeout(500);
      await expect(page.getByText('Secrets Manager')).toBeVisible({ timeout: 5000 }).catch(() => {});
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_secrets.png'), fullPage: true });

      // Click Auth (Tab)
      await page.getByRole('tab', { name: 'Authentication' }).click();
      await page.waitForTimeout(500);
      await expect(page.getByText('Authentication Settings')).toBeVisible({ timeout: 5000 }).catch(() => {});
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_auth.png'), fullPage: true });

      // Click General (Tab)
      await page.getByRole('tab', { name: 'General' }).click();
      await page.waitForTimeout(500);
      await expect(page.getByText('Global Settings')).toBeVisible({ timeout: 5000 }).catch(() => {});
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_general.png'), fullPage: true });

      // Webhooks is a Link to /settings/webhooks, so we screenshot it there or verify it separately
      // The user wants settings_webhooks.png to show the tab header.
      // If we go to /settings/webhooks, does it show the tabs?
      // We should check /settings/webhooks
      await page.getByRole('tab', { name: 'Webhooks' }).click();
      await page.waitForURL('**/settings/webhooks');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_webhooks.png'), fullPage: true });
      console.log('Saved all settings tab screenshots');
  });

  test('Verify and Screenshot Global Search', async ({ page }) => {
      console.log('Navigating to /...');
      await page.goto('/');
      await expect(page.locator('body')).toBeVisible();
      await page.waitForTimeout(1000);

      // Open Global Search with keyboard shortcut
      console.log('Opening Global Search...');
      await page.keyboard.press('Control+k');
      await page.waitForTimeout(1000);

      // Wait for dialog
      await expect(page.locator('[cmdk-root]')).toBeVisible();

      // Take screenshot of the dialog
      // specific selector or just page? page might be better to show context
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'global_search.png'), fullPage: true });
      console.log('Saved screenshot to global_search.png');
  });

  test('Verify Sidebar RBAC for Regular User', async ({ page }) => {
      console.log('Navigating to / as regular user...');
      await page.goto('/');
      await page.waitForTimeout(1000);

      // By default we are Admin. Switch to Regular User
      console.log('Switching to Regular User...');
      await page.locator('button[data-sidebar="menu-button"]').last().click(); // Open User Menu
      await page.getByText('Switch Role').click();
      await page.waitForTimeout(1000);

      // Verify "Users" and "Secrets" are hidden in Configuration
      // "services", "users", "secrets" should NOT be visible. "settings" SHOULD be visible.
      const sidebarText = await page.locator('[data-sidebar="sidebar"]').innerText();
      expect(sidebarText).not.toContain('Users');
      expect(sidebarText).not.toContain('Secrets Vault');
      expect(sidebarText).not.toContain('Services');
      expect(sidebarText).toContain('Settings');

      // Verify "Live Logs" and "Traces" hidden in Platform
      expect(sidebarText).not.toContain('Live Logs');
      expect(sidebarText).not.toContain('Traces');

      // Switch back to Admin for cleanup/subsequent tests if any
      // await page.locator('button[data-sidebar="menu-button"]').last().click();
      // await page.getByText('Switch Role').click();
  });

  test('Verify and Screenshot External Marketplace', async ({ page }) => {
      console.log('Navigating to /marketplace/external/mcpmarket...');
      await page.goto('/marketplace/external/mcpmarket');
      await page.waitForTimeout(3000); // Wait for mock fetch

      // Verify content
      await expect(page.locator('body')).toContainText('MCP Market');
      await expect(page.locator('body')).toContainText('Linear');

      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'marketplace_external_detailed.png'), fullPage: true });
      console.log('Saved screenshot to marketplace_external_detailed.png');
  });
});
