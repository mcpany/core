import { test, expect } from '@playwright/test';
import { seedGlobalState } from './e2e/test-data';

test.describe('Tool Inspector', () => {
  test.beforeEach(async ({ request }) => {
    await seedGlobalState(request);
  });

  test('Tools page loads and inspector opens with real data', async ({ page }) => {
    await page.goto('/tools');

    // Wait for a seeded tool to appear
    const toolRow = page.locator('tr').filter({ hasText: 'echo_tool' });
    await expect(toolRow).toBeVisible({ timeout: 20000 });

    // Click Inspect
    await toolRow.getByRole('button', { name: 'Inspect' }).click();

    // Verify dialog opens
    await expect(page.getByRole('dialog')).toBeVisible();

    // Check for the Visual tab in Schema section (default usually)
    await page.getByRole('tab', { name: 'Schema' }).click();
    await expect(page.getByRole('tab', { name: 'Visual' })).toBeVisible();
  });
});
