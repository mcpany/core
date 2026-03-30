import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Smart Result Renderer PRO', () => {
    test.beforeEach(async ({ request }) => {
        await seedGlobalState(request);
    });

    test('should render SmartTable in Playground PRO for array of objects', async ({ page }) => {
        // Navigate to the playground
        await page.goto('/playground');

        // Verify we are on the playground page
        const chatInput = page.getByPlaceholder('Enter command or select a tool...');
        await expect(chatInput).toBeVisible({ timeout: 15000 });

        // Enter a command that calls our seeded tool which returns an array of objects
        // In the CLI, calling `get_users {}` executes the get_users tool
        await chatInput.fill('get_users {}');

        // Click Send
        const sendBtn = page.getByRole('button', { name: 'Send' });
        await expect(sendBtn).toBeVisible();
        await sendBtn.click();

        // Wait for the result to come back
        // The SmartTable should render, meaning the search input should be visible
        const searchInput = page.getByPlaceholder('Search all columns...');
        await expect(searchInput).toBeVisible({ timeout: 15000 });

        // Verify that the table header for 'name' is rendered
        await expect(page.getByRole('columnheader', { name: 'name' })).toBeVisible();

        // Verify that the rows are rendered
        const table = page.locator('table').first();
        await expect(table).toBeVisible();

        // Verify some data from the seeded output is present in the table
        // We know 'Alice' is in the mocked data
        await expect(table.locator('td:has-text("Alice")').first()).toBeVisible();
    });
});
