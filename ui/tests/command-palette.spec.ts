import { test, expect } from '@playwright/test';

test.describe('Command Palette', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should open command palette with Cmd+K shortcut', async ({ page }) => {
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

  test('should open command palette with sidebar button', async ({ page }) => {
    // Click the search button in sidebar
    await page.getByText('Search...').click();

    // Check if dialog is visible
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
  });

  test('should filter items and navigate', async ({ page }) => {
    await page.keyboard.press('Meta+k');

    // Type "Services"
    await page.getByPlaceholder('Type a command or search...').fill('Services');

    // Verify Services option is present
    const servicesOption = page.getByRole('option', { name: 'Services' });
    await expect(servicesOption).toBeVisible();

    // Press Enter to navigate
    await page.keyboard.press('Enter');

    // Verify URL changed
    await expect(page).toHaveURL(/.*\/services/);
  });

  test('should close on Escape', async ({ page }) => {
    await page.keyboard.press('Meta+k');
    await expect(page.getByRole('dialog')).toBeVisible();

    await page.keyboard.press('Escape');
    await expect(page.getByRole('dialog')).not.toBeVisible();
  });
});
