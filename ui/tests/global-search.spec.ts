import { test, expect } from '@playwright/test';

test('Global Search (Cmd+K) works', async ({ page }) => {
  page.on('console', msg => console.log(`Browser console: ${msg.text()}`));
  page.on('pageerror', exception => console.log(`Browser exception: "${exception}"`));

  await page.goto('/');

  // Wait for hydration
  await page.waitForTimeout(2000);

  // Press Cmd+K (or Ctrl+K)
  await page.keyboard.press('Meta+k');
  await page.keyboard.press('Control+k');

  // Check if dialog is open
  // cmdk dialog usually has role="dialog" or we can look for the input placeholder
  await expect(page.getByPlaceholder('Type a command or search...')).toBeVisible({ timeout: 10000 });

  // Search for "stripe"
  await page.keyboard.type('stripe');

  // Check if stripe_charge option is visible.
  const stripeOption = page.getByRole('option').filter({ hasText: 'stripe_charge' }).first();
  await expect(stripeOption).toBeVisible();

  // Take screenshot of UI
  await page.screenshot({ path: '.audit/ui/2025-12-22/global_search_ui.png' });

  // Press Enter to select
  await page.keyboard.press('Enter');

  // Verify URL changed to /tools
  await expect(page).toHaveURL(/\/tools/);

  // Take screenshot for audit
  await page.screenshot({ path: '.audit/ui/2025-12-22/global_search.png', fullPage: true });
});
