import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Context Dashboard', () => {
    test.beforeEach(async ({ request, page }) => {
        // Seed the database with real data to ensure the UI has something to fetch.
        await seedGlobalState(request);
    });

    test('should load actual data without requiring manual seed', async ({ page }) => {
        // Log in first to ensure API calls succeed.
        await page.goto('/login');
        await page.waitForLoadState('networkidle');
        await page.fill('input[name="username"]', 'e2e-admin-core');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');

        // Go to the Context Dashboard
        await page.goto('/context');

        // Verify the title is present
        await expect(page.getByRole('heading', { name: 'Recursive Context Dashboard' })).toBeVisible();

        // Verify the "Seed Data" button is GONE (we removed it)
        await expect(page.getByRole('button', { name: /Seed Data/i })).not.toBeVisible();

        // Check if the Treemap or Simulator loaded.
        const simulatorHeading = page.getByRole('heading', { name: 'Simulator' });
        await expect(simulatorHeading).toBeVisible();

        // Check for either the Treemap SVG or the empty state message.
        // Either means it successfully fetched data (even if empty).
        const hasEmptyState = await page.getByText('No active tools in context.').isVisible();
        if (!hasEmptyState) {
            // If tools were discovered, we should see the Treemap container
            await expect(page.locator('.recharts-responsive-container')).toBeVisible();
        }
    });
});
