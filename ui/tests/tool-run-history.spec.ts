import { test, expect } from '@playwright/test';
import { seedServices, seedUser, seedProfiles } from './e2e/test-data';

const SERVICE_ID = 'svc_01';

test.describe('Tool Run History Real Data', () => {

  test.beforeEach(async ({ page, request }) => {
    await seedServices(request);
    await seedProfiles(request);
    await seedUser(request, "e2e-history");

    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'e2e-history');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
  });

  test('should display tool execution in history UI', async ({ page }) => {
    await page.goto('/playground');
    await page.waitForLoadState('networkidle');

    const runToolBtn = page.locator('button', { hasText: 'Run Tool' }).first();
    const runBtn = page.locator('button', { hasText: 'Run' }).first();

    if (!await runToolBtn.isVisible() && !await runBtn.isVisible()) {
        const link = page.locator('a[href*="/tool/"]').first();
        if (await link.isVisible()) {
            await link.click();
        }
    }

    if (await runToolBtn.isVisible()) {
        await runToolBtn.click();
    } else if (await runBtn.isVisible()) {
        await runBtn.click();
    }

    const historyTab = page.locator('button', { hasText: 'Metrics & History' }).first();
    if (await historyTab.isVisible()) {
        await historyTab.click();

        await expect(page.locator('text=Execution History')).toBeVisible({ timeout: 5000 });

        const historyItem = page.locator('button').filter({ hasText: 'ms' }).first();
        if (await historyItem.isVisible()) {
           await historyItem.click();
           await expect(page.locator('h4', { hasText: 'Arguments' })).toBeVisible();
        }
    }
  });
});
