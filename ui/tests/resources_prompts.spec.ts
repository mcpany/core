
import { test, expect } from '@playwright/test';

test('Resources page lists resources and allows toggling', async ({ page }) => {
  await page.goto('/resources');
  await expect(page.getByRole('heading', { name: 'Resources' })).toBeVisible();

  await expect(page.getByText('api_spec.yaml')).toBeVisible();

  // Test toggle
  const toggle = page.locator('tr:has-text("system_logs") button[role="switch"]');
  await expect(toggle).not.toBeChecked(); // Initially disabled
  await toggle.click();
  await expect(toggle).toBeChecked();
});

test('Prompts page lists prompts and allows toggling', async ({ page }) => {
    await page.goto('/prompts');
    await expect(page.getByRole('heading', { name: 'Prompts' })).toBeVisible();

    await expect(page.getByText('summarize_text')).toBeVisible();

    // Test toggle
    const toggle = page.locator('tr:has-text("creative_story") button[role="switch"]');
    await expect(toggle).not.toBeChecked(); // Initially disabled
    await toggle.click();
    await expect(toggle).toBeChecked();
});
