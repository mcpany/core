import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './test-data';

test.describe('Dashboard Persistence', () => {
    test.beforeEach(async ({ request }) => {
        await seedUser(request, "test-dashboard-user");
    });

    test.afterEach(async ({ request }) => {
        await cleanupUser(request, "test-dashboard-user");
    });

    test('persists widget layout after reload', async ({ page }) => {
        // 1. Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'test-dashboard-user');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });

        // Check initial count of "Quick Actions" widgets (Default layout might have one)
        const widgetSelector = '.group\\/widget';
        const quickActions = page.locator(widgetSelector, { hasText: 'Quick Actions' });
        await expect(page.locator(widgetSelector).first()).toBeVisible(); // Wait for dashboard to load
        const initialCount = await quickActions.count();

        // 2. Add "Quick Actions" widget
        await page.click('button:has-text("Add Widget")');
        await expect(page.locator('[role="dialog"]')).toBeVisible();
        await expect(page.getByText('Choose a widget')).toBeVisible();

        // Add "Quick Actions" by clicking its card title inside the dialog
        await page.click('[role="dialog"] >> text=Quick Actions');
        await expect(page.locator('[role="dialog"]')).toBeHidden();

        // 3. Verify it is added (Count should increase by 1)
        await expect(quickActions).toHaveCount(initialCount + 1);

        // 4. Wait for debounce save (1s) + persistence
        await page.waitForTimeout(2000);

        // 5. Reload the page
        await page.reload();
        await page.waitForLoadState('networkidle');

        // 6. Verify persistence: Count should still be initial + 1
        await expect(quickActions).toHaveCount(initialCount + 1);
    });
});
