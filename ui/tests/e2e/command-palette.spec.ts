
import { test, expect } from '@playwright/test';

test.describe('Command Palette', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should open with Cmd+K and navigate', async ({ page }) => {
    // Wait for page to load
    await expect(page.locator('body')).toBeVisible();

    // Press Cmd+K
    // Using page.keyboard.press with 'Meta+K' or 'Control+K'
    const modifier = process.platform === 'darwin' ? 'Meta' : 'Control';
    await page.keyboard.press(`${modifier}+k`);

    // Check if dialog is visible
    const commandInput = page.getByPlaceholder('Type a command or search...');
    await expect(commandInput).toBeVisible({ timeout: 10000 });

    // Type 'Logs' to filter
    await commandInput.fill('Logs');

    // Select the Logs item
    // Using getByRole or text to find the item
    const logsItem = page.getByText('Logs', { exact: true });
    await expect(logsItem).toBeVisible();
    await logsItem.click();

    // Should navigate to /logs
    await expect(page).toHaveURL(/.*\/logs/);
  });

  test('should toggle theme', async ({ page }) => {
    // Wait for page to load
    await expect(page.locator('body')).toBeVisible();

    const modifier = process.platform === 'darwin' ? 'Meta' : 'Control';
    await page.keyboard.press(`${modifier}+k`);

    const commandInput = page.getByPlaceholder('Type a command or search...');
    await expect(commandInput).toBeVisible({ timeout: 10000 });

    // Type 'Dark'
    await commandInput.fill('Dark');

    // Select Dark mode
    const darkItem = page.getByText('Dark', { exact: true });
    await expect(darkItem).toBeVisible();
    await darkItem.click();

    // Verify dark mode class on html or body.
    await expect(page.locator('html')).toHaveClass(/dark/);
  });
});
