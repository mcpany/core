
import { test, expect } from '@playwright/test';

test('services page loads and lists services', async ({ page }) => {
  await page.goto('/services');
  await expect(page.getByRole('heading', { name: 'Services' })).toBeVisible();

  // Check for mock data
  await expect(page.getByText('Payment Gateway')).toBeVisible();
  await expect(page.getByText('User Service')).toBeVisible();

  // Test toggle interaction (visual check mainly as it's mocked)
  const toggle = page.locator('button[role="switch"]').first();
  await expect(toggle).toBeVisible();
  await toggle.click();
});
