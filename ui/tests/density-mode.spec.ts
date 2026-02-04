/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';

test.describe('Dashboard Density Mode', () => {
  test.beforeEach(async ({ request, page }) => {
      await seedServices(request);
      await seedUser(request, "density-admin");

      await page.goto('/login');
      await page.fill('input[name="username"]', 'density-admin');
      await page.fill('input[name="password"]', 'password');
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL('/', { timeout: 15000 });
  });

  test.afterEach(async ({ request }) => {
      await cleanupServices(request);
      await cleanupUser(request, "density-admin");
  });

  test('Toggles between comfortable and compact mode', async ({ page }) => {
    // 1. Verify default comfortable mode (gap-4)
    // The grid container should have gap-4.
    // Note: There might be multiple grids. The dashboard main grid is likely the first big one.
    // Or we can look for the specific one.
    const grid = page.locator('.grid.grid-cols-12').first();
    await expect(grid).toHaveClass(/gap-4/);

    // 2. Click Compact toggle
    const compactBtn = page.locator('button[title="Compact Density"]');
    await expect(compactBtn).toBeVisible();
    await compactBtn.click();

    // 3. Verify compact mode (gap-2)
    await expect(grid).toHaveClass(/gap-2/);

    // 4. Verify widget styling changes (e.g. padding in MetricsOverview)
    // Metric cards have padding. In comfortable it is p-6 (default CardHeader/Content).
    // In compact it is p-3.
    // We check for the presence of 'p-3' class which we explicitly added.
    const metricCardHeader = page.locator('.text-sm.font-medium.text-muted-foreground').first().locator('..'); // Parent is CardHeader
    await expect(metricCardHeader).toHaveClass(/p-3/);

    // 5. Reload and verify persistence
    await page.reload();
    await expect(grid).toHaveClass(/gap-2/);

    // 6. Switch back to comfortable
    const comfortableBtn = page.locator('button[title="Comfortable Density"]');
    await comfortableBtn.click();
    await expect(grid).toHaveClass(/gap-4/);
  });
});
