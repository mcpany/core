/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Navigation Coverage', () => {
  const routes = [
    { path: '/', title: 'Dashboard' },
    { path: '/logs', title: 'Live Logs' },
    { path: '/marketplace', title: 'Marketplace' },
    { path: '/playground', title: 'Console' },
    { path: '/profiles', title: 'Profiles' },
    { path: '/prompts', title: 'Prompt Library' },
    { path: '/resources', title: 'Resources' },
    { path: '/secrets', title: 'API Keys & Secrets' },
    { path: '/upstream-services', title: 'Upstream Services' },
    { path: '/settings', title: 'Settings' },
    { path: '/stacks', title: 'Stacks' },
    { path: '/stats', title: 'Analytics & Stats' },
    { path: '/tools', title: 'Tools' },
    { path: '/webhooks', title: 'Integrations' },
  ];

  for (const route of routes) {
    test(`should navigate to ${route.path} and show title`, async ({ page }) => {
      await page.goto(route.path);

      // Wait for URL to match
      await expect(page).toHaveURL(new RegExp(route.path === '/' ? '/$' : route.path));

      // Check for a heading matching the expected title
      const titleRegex = new RegExp(route.title, 'i');
      await expect(
        page.getByRole('heading').filter({ hasText: titleRegex }).first()
      ).toBeVisible({ timeout: 10000 });
    });
  }
});
