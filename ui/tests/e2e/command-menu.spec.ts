import { test, expect } from '@playwright/test';

test.describe('Command Menu', () => {
  test('should open command menu with keyboard shortcut', async ({ page }) => {
    await page.goto('/');

    // Wait for the page to be ready
    await page.waitForLoadState('networkidle');

    // Press Cmd+K (or Ctrl+K)
    await page.keyboard.press('Control+k');

    // Check if the dialog is visible
    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible();

    // Check for input field
    const input = page.locator('input[placeholder="Type a command or search..."]');
    await expect(input).toBeVisible();
    await expect(input).toBeFocused();
  });

  test('should navigate to services page', async ({ page }) => {
    await page.goto('/');

    // Open menu
    await page.keyboard.press('Control+k');

    // Type "Services"
    await page.keyboard.type('Services');

    // Use keyboard navigation to select the item
    await page.keyboard.press('ArrowDown');
    await page.keyboard.press('Enter');

    // Verify URL
    await expect(page).toHaveURL(/\/services/);
  });
});
