import { test, expect } from '@playwright/test';

test('Global Search opens on Cmd+K and navigates', async ({ page }) => {
  await page.goto('/');

  // Press Cmd+K (or Ctrl+K)
  await page.keyboard.press('Control+k');

  // Check if search input is visible
  const searchInput = page.getByPlaceholder('Type a command or search...');
  await expect(searchInput).toBeVisible();

  // Type "Dash"
  await searchInput.fill('Dash');

  // Check if Dashboard option is visible
  const dashboardOption = page.getByRole('option', { name: 'Dashboard' });
  await expect(dashboardOption).toBeVisible();

  // Select Dashboard
  await dashboardOption.click();

  // Verify navigation (we are already on dashboard, so it might just close)
  // Let's try navigating to Settings first
  await page.keyboard.press('Control+k');
  await searchInput.fill('Settings');
  const settingsOption = page.getByRole('option', { name: 'Settings' });
  await settingsOption.click();
  await expect(page).toHaveURL(/\/settings/);
});
