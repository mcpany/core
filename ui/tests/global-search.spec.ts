
import { test, expect } from '@playwright/test';

test.describe('Global Search', () => {
  test('should open command menu with Cmd+K and navigate', async ({ page }) => {
    // Go to homepage
    await page.goto('/');

    // Press Cmd+K
    await page.keyboard.press('Meta+k');

    // Check if dialog is open
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // Check if input is focused (implicit by checking we can type)
    const input = page.getByPlaceholder('Type a command or search...');
    await expect(input).toBeVisible();
    await input.fill('Services');

    // Select Services
    await page.keyboard.press('Enter');

    // Should navigate to /services
    await expect(page).toHaveURL(/\/services/);
  });
});
