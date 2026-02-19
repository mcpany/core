/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { cleanupServices } from './e2e/test-data';

test('getting started widget appears on dashboard', async ({ page, request }) => {
  // Ensure we have a clean slate so the widget appears (Real Data Law)
  await cleanupServices(request);

  await page.goto('/');
  // Login if redirected (though cleanupServices might not affect auth, but let's be safe)
  // Check if we are on login page
  if (page.url().includes('/login')) {
      await page.fill('input[name="password"]', 'test-token');
      await page.click('button[type="submit"]');
  }

  // Expect to see the Getting Started widget title
  await expect(page.getByText('Welcome to MCP Any')).toBeVisible();

  // Expect to see the "Quick Start" button
  await expect(page.getByText('Quick Start')).toBeVisible();
});
