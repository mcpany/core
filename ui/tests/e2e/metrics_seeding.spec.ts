import { test, expect } from '@playwright/test';
import { seedDashboard, cleanupUser, seedUser } from './test-data';

test.describe('Metrics Seeding & Real Data', () => {
    test.beforeAll(async ({ request }) => {
        await seedDashboard(request);
        await seedUser(request, "e2e-admin-metrics");
    });

    test.afterAll(async ({ request }) => {
        await cleanupUser(request, "e2e-admin-metrics");
    });

    test.beforeEach(async ({ page }) => {
        // Login as the seeded user
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-admin-metrics');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await page.waitForURL('/');
    });

    test('should display seeded historical metrics in tool runner', async ({ page }) => {
        // Navigate to tools list
        await page.goto('/tools');

        // Search and click on a known seeded tool
        await page.fill('input[placeholder="Search tools..."]', 'calculate_sum');
        // Click the open inspector button for calculate_sum tool
        await page.click('button[title="Open Inspector"]');

        // Go to Metrics & History tab
        await page.click('button:has-text("Metrics & History")');

        // Wait for metrics to load
        await page.waitForSelector('text=Total Calls');

        // Verify the seeded metrics are displayed properly.
        // We seeded 100 requests in seedTraffic, but ToolRunner gets stats from ToolUsage endpoint.
        // Since we also ran seedTraces, there should be at least 1 total call for calculate_sum.

        // Assert that Total Calls card is present and has a value > 0
        const totalCallsText = await page.locator('.bg-background\\/50 >> text=Total Calls').locator('..').locator('p.text-2xl').innerText();
        expect(parseInt(totalCallsText, 10)).toBeGreaterThan(0);

        // Assert Success Rate card is present
        const successRateText = await page.locator('.bg-background\\/50 >> text=Success Rate').locator('..').locator('p.text-2xl').innerText();
        expect(successRateText).toContain('%');

        // Assert execution latency chart container exists
        await expect(page.locator('text=Execution Latency (ms)')).toBeVisible();
    });
});
