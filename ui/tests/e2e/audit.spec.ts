/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Audit Smoke Tests', () => {
    // Instead of screenshots, we ensure critical elements are present on all key pages.
  test('should verify all main pages load correctly', async ({ page }) => {
    const pages = [
      { url: '/', check: 'Dashboard' },
      { url: '/services', check: 'Services' },
      { url: '/tools', check: 'Tools' },
      { url: '/resources', check: 'Resources' },
      { url: '/prompts', check: 'Prompts' },
      { url: '/profiles', check: 'Profiles' },
      { url: '/middleware', check: 'Middleware Pipeline' },
      { url: '/webhooks', check: 'Webhooks' },
    ];

    for (const p of pages) {
      await page.goto(p.url);
      await expect(page.locator('body')).toContainText(p.check);
    }

    // Corner case: 404
    await page.goto('/non-existent-page');
    await expect(page.locator('body')).toContainText('404');
  });
});
