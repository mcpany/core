import { test, expect } from '@playwright/test';

test.describe('Resource Explorer', () => {
    test.beforeEach(async ({ page }) => {
        // Mock the resources API to return hierarchical data
        await page.route('*/**/api/resources', async (route) => {
            await route.fulfill({
                json: {
                    resources: [
                        { uri: 'file:///app/src/main.go', name: 'main.go', mimeType: 'text/x-go' },
                        { uri: 'file:///app/public/index.html', name: 'index.html', mimeType: 'text/html' },
                        { uri: 'file:///app/README.md', name: 'README.md', mimeType: 'text/markdown' },
                        { uri: 'postgres://db/users/schema', name: 'users_schema', mimeType: 'application/sql' }
                    ]
                }
            });
        });

        // Mock reading resource content
        await page.route('*/**/api/resources/read*', async (route) => {
            await route.fulfill({
                json: {
                    contents: [
                        { uri: 'file:///app/src/main.go', mimeType: 'text/x-go', text: 'package main\n\nfunc main() {}' }
                    ]
                }
            });
        });

        await page.goto('/resources');
    });

    test('should render the resource tree correctly', async ({ page }) => {
        // Check for root folders
        await expect(page.getByText('app', { exact: true })).toBeVisible();
        await expect(page.getByText('postgres://', { exact: true })).toBeVisible();
    });

    test('should navigate into folders via sidebar', async ({ page }) => {
        // Click 'app' folder in sidebar
        // Note: The tree component renders items with padding.
        // We use text locator.
        await page.getByText('app', { exact: true }).click();

        // Should show 'src', 'public', 'README.md' in main area
        await expect(page.locator('.divide-y').getByText('src')).toBeVisible();
        await expect(page.locator('.divide-y').getByText('public')).toBeVisible();
        await expect(page.locator('.divide-y').getByText('README.md')).toBeVisible();
    });

    test('should navigate via breadcrumbs', async ({ page }) => {
        // Navigate deep: app -> src
        await page.getByText('app', { exact: true }).click();

        // Wait for main area to update
        await expect(page.locator('.divide-y').getByText('src')).toBeVisible();

        // Click 'src' folder in main area
        await page.locator('.divide-y').getByText('src').dblclick();

        // Check breadcrumb
        await expect(page.getByRole('button', { name: 'src' })).toBeVisible();

        // Click 'app' in breadcrumb
        await page.getByRole('button', { name: 'app' }).click();

        // Should be back in 'app' folder
        await expect(page.locator('.divide-y').getByText('src')).toBeVisible();
    });

    test('should preview file content', async ({ page }) => {
        await page.getByText('app', { exact: true }).click();
        await page.locator('.divide-y').getByText('src').dblclick();

        // Click file
        await page.locator('.divide-y').getByText('main.go').click();

        // Check preview pane header
        await expect(page.getByText('file:///app/src/main.go')).toBeVisible();

        // Check content (mocked)
        await expect(page.getByText('package main')).toBeVisible();
    });
});
