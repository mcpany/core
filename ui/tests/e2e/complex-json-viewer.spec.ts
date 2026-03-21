import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('RichResultViewer Complex JSON Table', () => {
    test.beforeEach(async ({ request }) => {
        await seedGlobalState(request);
    });

    test('renders flattened complex nested JSON in table', async ({ page }) => {
        await page.goto('/tools');

        // Wait for tools list to load
        await page.waitForSelector('text=complex_tool');

        // Find the complex_tool card and click Inspect
        const toolCard = page.locator('.card:has-text("complex_tool")').first();
        await toolCard.locator('a:has-text("Inspect")').click();

        // Ensure we are on the tool page
        await expect(page).toHaveURL(/.*\/tools\/complex_tool/);

        // Click Run Tool
        await page.getByRole('button', { name: 'Run Tool' }).click();

        // Wait for the Result section to appear
        await expect(page.locator('text=Result')).toBeVisible();

        // Assert that the table headers correspond to the flattened keys
        await expect(page.locator('th:has-text("USER.PROFILE.NAME")')).toBeVisible();
        await expect(page.locator('th:has-text("USER.ROLE")')).toBeVisible();
        await expect(page.locator('th:has-text("METADATA.PREFERENCES.THEME")')).toBeVisible();
        await expect(page.locator('th:has-text("CONTACTS")')).toBeVisible();

        // Assert that the table cells contain the correct data
        await expect(page.locator('td:has-text("Alice Liddell")')).toBeVisible();
        await expect(page.locator('td:has-text("admin")')).toBeVisible();
        await expect(page.locator('td:has-text("dark")')).toBeVisible();
        await expect(page.locator('td:has-text("type: email, value: alice@example.com")')).toBeVisible();
    });
});
