import { test, expect } from '@playwright/test';

test.describe('Global Search', () => {
  test('should open command menu with Cmd+K and navigate', async ({ page }) => {
    // Start from the home page
    await page.goto('/');

    // Wait for page to load
    await page.waitForLoadState('domcontentloaded');

    // Press Cmd+K (or Ctrl+K)
    if (process.platform === 'darwin') {
      await page.keyboard.press('Meta+k');
    } else {
      await page.keyboard.press('Control+k');
    }

    // Check if the command dialog is visible
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // Check if input is focused
    const input = page.getByPlaceholder('Type a command or search...');
    await expect(input).toBeVisible();
    await expect(input).toBeFocused();

    // Type 'Services'
    await input.fill('Services');

    // Wait for filtering
    const servicesOption = page.getByRole('option', { name: 'Services' });
    await expect(servicesOption).toBeVisible();

    // Select it (click or enter)
    await servicesOption.click();

    // Verify navigation
    await expect(page).toHaveURL(/.*\/services/);
  });

  test('should toggle theme', async ({ page }) => {
    await page.goto('/');

    // Press Cmd+K
    if (process.platform === 'darwin') {
        await page.keyboard.press('Meta+k');
    } else {
        await page.keyboard.press('Control+k');
    }

    // Type 'Dark'
    const input = page.getByPlaceholder('Type a command or search...');
    await input.fill('Dark');

    // Select Dark mode
    await page.getByRole('option', { name: 'Dark' }).click();

    // Verify html class changed to dark (next-themes)
    // Note: next-themes might use class or data-attribute
    // The ThemeProvider in layout.tsx uses attribute="class"
    await expect(page.locator('html')).toHaveClass(/dark/);
  });
});
