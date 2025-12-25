
import { test, expect } from '@playwright/test';

test('Settings page tabs and content verification', async ({ page }) => {
  await page.goto('/settings');
  await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();

  // Profiles Tab (default)
  await expect(page.getByRole('tab', { name: 'Profiles' })).toHaveAttribute('aria-selected', 'true');
  await expect(page.getByText('Execution Profiles')).toBeVisible();
  await expect(page.getByText('Development', { exact: true })).toBeVisible();

  // Webhooks Tab
  await page.getByRole('tab', { name: 'Webhooks' }).click();
  await expect(page.getByText('Endpoint URL')).toBeVisible();

  // Add a webhook
  await page.getByLabel('Endpoint URL').fill('https://test.com/hook');
  await page.getByRole('button', { name: 'Add Webhook' }).click();
  await expect(page.getByText('https://test.com/hook')).toBeVisible();

  // Middleware Tab
  await page.getByRole('tab', { name: 'Middleware' }).click();
  await expect(page.getByText('Middleware Pipeline')).toBeVisible();
  await expect(page.getByText('Authentication')).toBeVisible();
});
