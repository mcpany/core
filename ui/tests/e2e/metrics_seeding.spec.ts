import { test, expect, request as pwRequest } from '@playwright/test';
import { seedDashboard, cleanupUser, seedUser } from './test-data';

test.describe('Metrics Seeding & Real Data', () => {
    test.beforeAll(async () => {
        const context = await pwRequest.newContext();
        await seedDashboard(context);
        await seedUser(context, "e2e-admin-metrics");
        await context.dispose();
    });

    test.afterAll(async () => {
        const context = await pwRequest.newContext();
        await cleanupUser(context, "e2e-admin-metrics");
        await context.dispose();
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
        await page.click('button:has-text("Inspect")');

        // Go to Metrics & History tab
        await page.click('button:has-text("Metrics & History")');

        // Wait for metrics to load
        await page.waitForSelector('text=Total Calls');

        // Verify the seeded metrics are displayed properly.
        // We seeded 100 requests in seedTraffic, but ToolRunner gets stats from ToolUsage endpoint.
        // Since we also ran seedTraces, there should be at least 1 total call for calculate_sum.

        // Assert that Total Calls card is present and has a value > 0
        const totalCallsText = await page.locator('div').filter({ hasText: 'Total Calls' }).locator('p.text-2xl').innerText();
        expect(parseInt(totalCallsText, 10)).toBeGreaterThan(0);

        // Assert Success Rate card is present
        const successRateText = await page.locator('div').filter({ hasText: 'Success Rate' }).locator('p.text-2xl').innerText();
        expect(successRateText).toContain('%');

        // Assert execution latency chart container exists
        await expect(page.locator('text=Execution Latency (ms)')).toBeVisible();
    });
});
