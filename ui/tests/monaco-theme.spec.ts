import { test, expect } from '@playwright/test';

test.describe('Monaco Editor Theme', () => {
  test('should render with dark theme when system is dark', async ({ page }) => {
    // Navigate to the Config Validator page
    await page.goto('/config-validator');

    // Emulate system dark color scheme
    await page.emulateMedia({ colorScheme: 'dark' });

    // Wait for the Monaco editor to load
    const editorSelector = '.monaco-editor';
    await page.locator(editorSelector).first().waitFor({ state: 'visible' });

    // Wait for Next-themes to hydrate and apply the dark mode classes.
    // In our implementation, Next.js root html gets a class='dark'.
    await expect(page.locator('html')).toHaveClass(/dark/);

    // After hydrating, we need to assert that Monaco editor has the 'dracula' theme applied.
    // The dracula theme sets the editor background or gives it a specific class.
    // Normally, custom themes get added as a class like '.vs-dark' or '.dracula'.
    // Monaco adds standard classes for dark themes like '.vs-dark'.
    // Let's check that the monaco editor has the .vs-dark class.

    // Sometimes custom themes in Monaco don't directly add the class name to the container,
    // but they add 'vs-dark' because they are derived from it.
    await expect(page.locator(editorSelector).first()).toHaveClass(/vs-dark/);
  });
});
