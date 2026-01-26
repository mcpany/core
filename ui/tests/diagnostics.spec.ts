import { test, expect } from '@playwright/test';

test.describe('Service Diagnostics', () => {
  const serviceId = 'test-service';
  const serviceName = 'Test Service';

  test.beforeEach(async ({ page }) => {
    // Mock the service detail response
    await page.route(`**/api/v1/services/${serviceId}`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          service: {
            id: serviceId,
            name: serviceName,
            http_service: { address: 'http://example.com' },
            version: '1.0.0',
            priority: 1,
            last_error: '', // Start healthy
          }
        }),
      });
    });

    // Mock validate endpoint (Success case)
    await page.route(`**/api/v1/services/validate`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ valid: true }),
      });
    });
  });

  test('should display diagnostics tab and run check successfully', async ({ page }) => {
    await page.goto(`/upstream-services/${serviceId}`);

    // Check if Diagnostics tab exists
    const diagTab = page.getByRole('tab', { name: 'Diagnostics' });
    await expect(diagTab).toBeVisible();
    await diagTab.click();

    // Check "Run Diagnostics" button
    const runBtn = page.getByRole('button', { name: 'Run Diagnostics' });
    await expect(runBtn).toBeVisible();

    // Click run
    await runBtn.click();

    // Expect success message
    await expect(page.getByText('Configuration Valid & Reachable')).toBeVisible();
  });

  test('should display validation errors', async ({ page }) => {
    // Override mock for failure
    await page.route(`**/api/v1/services/validate`, async (route) => {
      await route.fulfill({
        status: 200, // API returns 200 with valid: false
        contentType: 'application/json',
        body: JSON.stringify({
          valid: false,
          error: 'Connection refused',
          details: 'Failed to reach http://example.com'
        }),
      });
    });

    await page.goto(`/upstream-services/${serviceId}`);
    await page.getByRole('tab', { name: 'Diagnostics' }).click();
    await page.getByRole('button', { name: 'Run Diagnostics' }).click();

    // Expect error message
    // Use :text-is to match exact text or be specific
    await expect(page.getByText('Check Failed')).toBeVisible();
    // The previous test failed because "Connection refused" appeared in multiple places (toast + detailed error)
    // We can target the specific container or use .first() if we just want to ensure it's visible somewhere.
    await expect(page.getByText('Connection refused').first()).toBeVisible();
    await expect(page.getByText('Failed to reach http://example.com').first()).toBeVisible();
  });
});
