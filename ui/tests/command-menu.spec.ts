import { test, expect } from '@playwright/test';

test.describe('Command Menu', () => {
  test.beforeEach(async ({ page }) => {
    // Set a large viewport to ensure sidebar is likely expanded
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/');
    // Ensure the page is loaded
    await expect(page.locator('h2', { hasText: 'Dashboard' })).toBeVisible();
  });

  test('should open with Cmd+K', async ({ page }) => {
    // Press Cmd+K
    await page.keyboard.press('Meta+k');

    // Check if dialog is visible
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // Check if input is focused
    const input = page.getByPlaceholder('Type a command or search...');
    await expect(input).toBeVisible();
    await expect(input).toBeFocused();
  });

  test('should open via Sidebar button', async ({ page }) => {
    // If the sidebar is collapsed, the text "Search" might be hidden.
    // However, the button itself should still be present (showing the icon).
    // Let's try to find the button that has the tooltip "Search (Cmd+K)" or contains the Search icon.

    // We can target the sidebar footer.
    const sidebarFooter = page.locator('[data-sidebar="footer"]');

    // Find a button inside the footer.
    const searchButton = sidebarFooter.locator('button').first();

    await expect(searchButton).toBeVisible();
    await searchButton.click();

    // Check if dialog is visible
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
  });

  test('should navigate to pages', async ({ page }) => {
    await page.keyboard.press('Meta+k');

    // Type "Services"
    const input = page.getByPlaceholder('Type a command or search...');
    await input.fill('Services');

    // Select the item
    await page.keyboard.press('Enter');

    // Verify URL
    await expect(page).toHaveURL(/.*\/services/);
  });

  test('should filter results', async ({ page }) => {
     await page.keyboard.press('Meta+k');
     const input = page.getByPlaceholder('Type a command or search...');

     // Type something that shouldn't match
     await input.fill('Xylophone');

     // Scope search to the command list to avoid matching text on the background page
     const list = page.locator('[cmdk-list]');

     await expect(list.getByText('No results found.')).toBeVisible();

     // Type "Log"
     await input.fill('Log');

     // Check that "Logs" item is in the list
     await expect(list.getByText('Logs', { exact: true })).toBeVisible();

     // Check that "Dashboard" item is NOT in the list
     await expect(list.getByText('Dashboard', { exact: true })).not.toBeVisible();
  });
});
