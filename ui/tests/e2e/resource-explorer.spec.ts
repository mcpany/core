import { test, expect } from '@playwright/test';

test.describe('Resource Explorer', () => {
  test('should load resources and allow selection', async ({ page }) => {
    // Navigate to the resources page
    await page.goto('http://localhost:9002/resources');

    // Wait for the resource list to populate (using mock data)
    // Use first() because the URI display might also contain the text "config.json"
    await expect(page.getByText('config.json').first()).toBeVisible();
    await expect(page.getByText('README.md').first()).toBeVisible();

    // Verify search functionality
    const searchInput = page.getByPlaceholder('Search resources...');
    await searchInput.fill('script');
    await expect(page.getByText('script.py').first()).toBeVisible();
    await expect(page.getByText('config.json')).not.toBeVisible();

    // Clear search
    await searchInput.fill('');
    await expect(page.getByText('config.json').first()).toBeVisible();

    // Select a resource
    await page.getByText('config.json').first().click();

    // Verify preview loads
    // Use first() because the list item also shows the URI
    await expect(page.getByText('file:///app/config.json').first()).toBeVisible(); // URI header

    // Check if content area is visible (looking for syntax highlighter or code)
    // The mock returns JSON content
    await expect(page.locator('.prose, pre, code').first()).toBeVisible();

    // Verify toolbar buttons
    await expect(page.getByTitle('List View')).toBeVisible();
    await expect(page.getByTitle('Grid View')).toBeVisible();

    // Switch to Grid view
    await page.getByTitle('Grid View').click();

    // Verify grid item exists
    // In grid view, items are cards. We check for text again but layout changes.
    await expect(page.getByText('config.json').first()).toBeVisible();
  });
});
