/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('take screenshot of settings profiles', async ({ page }) => {
  await page.goto('http://localhost:9002/settings');
  await page.waitForTimeout(3000); // Wait for animations and data load
  await page.screenshot({ path: 'ui/docs/screenshots/settings_profiles_new.png' });
});
