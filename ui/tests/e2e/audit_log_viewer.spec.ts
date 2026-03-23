import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Audit Log Viewer', () => {
    test.beforeAll(async () => {
        await seedGlobalState();
    });

    test.beforeEach(async ({ page }) => {
        // Authenticate as a user
        await page.goto('/login');
        await page.fill('input[placeholder="Username"]', 'admin');
        await page.fill('input[type="password"]', 'admin');
        await page.click('button:has-text("Login")');
        await page.waitForURL('/');
    });

    test('should display seeded audit logs and open log details', async ({ page }) => {
        await page.goto('/audit');

        // Wait for the table row to appear
        await expect(page.locator('table')).toBeVisible();

        // Verify the seeded "process_payment" log is present
        await expect(page.locator('table >> text="process_payment"')).toBeVisible();

        const viewButton = page.locator('table tr:has-text("process_payment") button:has-text("View")');
        await expect(viewButton).toBeVisible();
        await viewButton.click();

        // Verify the dialog opens and displays details
        await expect(page.locator('div[role="dialog"]')).toBeVisible();

        // Verify the Arguments are visible as formatted JSON
        await expect(page.locator('div[role="dialog"] >> text="USD"')).toBeVisible();
        await expect(page.locator('div[role="dialog"] >> text="100"')).toBeVisible();

        // Verify the Result is visible as formatted JSON
        await expect(page.locator('div[role="dialog"] >> text="success"')).toBeVisible();
        await expect(page.locator('div[role="dialog"] >> text="txn_123"')).toBeVisible();
    });
});
