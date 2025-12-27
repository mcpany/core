import { test, expect } from '@playwright/test';

test.describe('Command Palette', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the dashboard
    await page.goto('/');
  });

  test('opens command palette with Cmd+K', async ({ page }) => {
    // Ensure the command palette is initially hidden
    await expect(page.getByPlaceholder('Type a command or search...')).not.toBeVisible();

    // Press Cmd+K
    await page.keyboard.press('Meta+k');

    // Verify it is visible
    await expect(page.getByPlaceholder('Type a command or search...')).toBeVisible();
  });

  test('opens command palette with Ctrl+K', async ({ page }) => {
    // Ensure the command palette is initially hidden
    await expect(page.getByPlaceholder('Type a command or search...')).not.toBeVisible();

    // Press Ctrl+K
    await page.keyboard.press('Control+k');

    // Verify it is visible
    await expect(page.getByPlaceholder('Type a command or search...')).toBeVisible();
  });

  test('navigates to Services page', async ({ page }) => {
    // Press Cmd+K
    await page.keyboard.press('Meta+k');

    // Type "Services"
    await page.getByPlaceholder('Type a command or search...').fill('Services');

    // Select the option (assuming it's the first one or we can find it by text)
    await page.getByRole('option', { name: 'Services' }).click();

    // Verify URL
    await expect(page).toHaveURL(/\/services/);
  });

  test('closes when pressing Escape', async ({ page }) => {
      // Press Cmd+K
      await page.keyboard.press('Meta+k');
      await expect(page.getByPlaceholder('Type a command or search...')).toBeVisible();

      // Press Escape
      await page.keyboard.press('Escape');
      await expect(page.getByPlaceholder('Type a command or search...')).not.toBeVisible();
  });
});
