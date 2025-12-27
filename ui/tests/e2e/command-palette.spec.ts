
import { test, expect } from '@playwright/test';

test.describe('Command Palette', () => {
  test('should open command palette with Cmd+K', async ({ page }) => {
    await page.goto('/');

    // Simulate Cmd+K
    await page.keyboard.press('Meta+k');

    // Check if dialog is visible
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByPlaceholder('Type a command or search...')).toBeVisible();
  });

  test('should navigate to Playground', async ({ page }) => {
    await page.goto('/');

    await page.keyboard.press('Meta+k');
    await expect(page.getByRole('dialog')).toBeVisible();

    await page.getByPlaceholder('Type a command or search...').fill('Playground');
    await page.getByRole('option', { name: 'Playground' }).click();

    await expect(page).toHaveURL(/.*\/playground/);
  });

   test('should toggle theme', async ({ page }) => {
    await page.goto('/');

    await page.keyboard.press('Meta+k');
    await expect(page.getByRole('dialog')).toBeVisible();

    await page.getByPlaceholder('Type a command or search...').fill('Dark Mode');
    // Note: Theme toggling might be hard to verify visually without screenshot comparison,
    // but we can check if the command executes without error.
    await page.getByRole('option', { name: 'Dark Mode' }).click();

    // Check if html has class dark
    await expect(page.locator('html')).toHaveClass(/dark/);
  });
});
