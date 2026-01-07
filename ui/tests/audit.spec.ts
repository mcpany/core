/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test } from '@playwright/test';
import path from 'path';

import fs from 'fs';

test('capture screenshots', async ({ page }) => {
  const date = new Date().toISOString().split('T')[0];
  const auditDir = path.join(__dirname, `../test-results/audit/ui/${date}`);
  if (!fs.existsSync(auditDir)) {
    fs.mkdirSync(auditDir, { recursive: true });
  }

  await page.goto('/');
  await page.waitForTimeout(1000); // Wait for animations
  await page.screenshot({ path: `${auditDir}/dashboard.png`, fullPage: true });

  await page.goto('/services');
  await page.waitForTimeout(1000);
  await page.screenshot({ path: `${auditDir}/services.png`, fullPage: true });

  await page.goto('/tools');
  await page.waitForTimeout(1000);
  await page.screenshot({ path: `${auditDir}/tools.png`, fullPage: true });

  await page.goto('/settings');
  await page.waitForTimeout(1000);
  await page.screenshot({ path: `${auditDir}/settings.png`, fullPage: true });
});
