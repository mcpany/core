import { test, expect } from '@playwright/test';

test('Global Search Integration', async ({ page }) => {
  // 1. Navigate to the dashboard
  await page.goto('/');

  // 2. Open Global Search with Cmd+K (simulated)
  // Note: Cmd+K can be tricky to simulate reliably across OS, but we'll try standard way
  await page.keyboard.press('Meta+k');

  // Wait for dialog to appear
  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible();

  // 3. Search for "Services"
  const searchInput = page.getByPlaceholder('Type a command or search...');
  await searchInput.fill('Services');

  // 4. Verify "Services" option is visible and select it
  const servicesOption = page.getByRole('option', { name: 'Services' }).first();
  await expect(servicesOption).toBeVisible();

  // Click it
  await servicesOption.click();

  // 5. Verify navigation to /services
  await expect(page).toHaveURL(/\/services/);
});
