import { test, expect } from '@playwright/test';

test('Global Search opens with keyboard shortcut', async ({ page }) => {
  await page.goto('/');

  // Press Cmd+K
  await page.keyboard.press('Meta+k');

  if (!await page.locator('input[placeholder="Type a command or search..."]').isVisible()) {
      await page.keyboard.press('Control+k');
  }

  // Check if dialog is visible
  const searchInput = page.getByPlaceholder('Type a command or search...');
  await expect(searchInput).toBeVisible();

  // Check if Navigation group exists
  await expect(page.getByText('Navigation')).toBeVisible();

  // Test navigation. Be specific because "Services" might be in the sidebar too.
  // We want the one inside the CommandDialog.
  const dialog = page.locator('[role="dialog"]');
  await dialog.getByText('Services').click();

  await expect(page).toHaveURL(/.*\/services/);
});
