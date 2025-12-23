import { test, expect } from '@playwright/test';

test('Global Search screenshot', async ({ page }) => {
  await page.goto('/');

  // Press Cmd+K (or Ctrl+K)
  await page.keyboard.press('Control+k');

  // Wait for search input
  const searchInput = page.getByPlaceholder('Type a command or search...');
  await expect(searchInput).toBeVisible();

  // Type "Settings"
  await searchInput.fill('Settings');

  // Wait a bit for animations
  await page.waitForTimeout(500);

  // Take screenshot
  await page.screenshot({ path: '.audit/ui/2025-12-23/global_search.png' });
});
