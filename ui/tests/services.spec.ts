
import { test, expect } from '@playwright/test';

test('Services page lists services and allows toggling', async ({ page }) => {
  await page.goto('/services');

  await expect(page.getByRole('heading', { name: 'Services' })).toBeVisible();

  // Check if mock services are listed
  await expect(page.getByText('Payment Gateway')).toBeVisible();
  await expect(page.getByText('User Service')).toBeVisible();

  // Check toggle functionality
  // Initially Payment Gateway is enabled (disable: false)
  const toggle = page.locator('tr:has-text("Payment Gateway") button[role="switch"]');
  await expect(toggle).toBeChecked();

  // Click to disable
  await toggle.click();
  await expect(toggle).not.toBeChecked();
  await expect(page.locator('tr:has-text("Payment Gateway")')).toContainText('Disabled');

  // Reload to persist (mock works in-memory per process, next dev might reload process so state might not persist if not global var, but for this test run it should be fine if same process)
  // Actually, in playwright test, the server is running separately. If I use `next dev`, the backend state in memory is persistent until restart.
});
