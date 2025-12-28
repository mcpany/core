import { test, expect } from '@playwright/test';

test('Global Search opens with keyboard shortcut', async ({ page }) => {
  await page.goto('/');

  // Press Cmd+K (or Ctrl+K)
  await page.keyboard.press('Meta+k');

  // Wait for the dialog to appear
  await expect(page.getByRole('dialog')).toBeVisible();

  // Check for the input
  await expect(page.getByPlaceholder('Type a command or search...')).toBeVisible();
});

test('Global Search navigates to Playground', async ({ page }) => {
  await page.goto('/');
  await page.keyboard.press('Meta+k');

  // Search for "Playground"
  const input = page.getByPlaceholder('Type a command or search...');
  await input.fill('Playground');

  // Click the result
  await page.getByText('Playground').click();

  // Verify navigation
  await expect(page).toHaveURL(/.*\/playground/);
});
