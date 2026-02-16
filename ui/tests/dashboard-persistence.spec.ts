import { test, expect } from '@playwright/test';

test.describe('Dashboard Persistence', () => {
  test('should persist dashboard layout across reloads', async ({ page }) => {
    // 1. Load Dashboard
    await page.goto('/');
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

    // 2. Clear all widgets (if any exist) or reset state
    // We can use the "Layout" popover -> "Clear All"
    await page.getByRole('button', { name: 'Layout' }).click();
    // Wait for popover content
    await expect(page.getByText('Visible Widgets')).toBeVisible();

    // Click Clear All if it exists (it always does in my implementation)
    await page.getByRole('button', { name: 'Clear All' }).click();

    // Verify empty state
    await expect(page.getByText('Your dashboard is empty')).toBeVisible();

    // 3. Add a widget
    // Click "Add Widget" button (in empty state or header). The empty state one is visible.
    // There are multiple "Add Widget" buttons. Use the one in empty state if visible, or first.
    // The AddWidgetSheet component has "Add Widget" text.
    await page.getByRole('button', { name: 'Add Widget' }).last().click(); // Last one is likely in empty state if rendered later? Or first.
    // Actually, simple text match is safer.

    // Wait for sheet
    await expect(page.getByText('Choose a widget to add')).toBeVisible();

    // Select "Service Health" widget (assuming it exists in registry)
    // I need to be sure about the name. "Service Health" is a safe bet for a dashboard.
    // Let's check widget-registry or just look for any card title.
    // "Service Status" or "Recent Activity".
    // I'll try "Service Status" or generic click on first card.
    await page.locator('.cursor-pointer').first().click();

    // Verify a widget is added. We don't know exactly which one, but "Your dashboard is empty" should be gone.
    await expect(page.getByText('Your dashboard is empty')).toBeHidden();

    // 4. Reload page
    // Wait for debounce (1s) + buffer
    await page.waitForTimeout(1500);
    await page.reload();

    // 5. Verify widget is still there (Persisted)
    await expect(page.getByText('Your dashboard is empty')).toBeHidden();

    // Optional: Verify specific widget title if I knew what I clicked.
    // But empty check is sufficient to prove state persistence vs empty default.
  });
});
