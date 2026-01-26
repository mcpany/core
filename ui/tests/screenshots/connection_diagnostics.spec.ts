import { test, expect } from '@playwright/test';

test('capture connection diagnostics screenshot', async ({ page }) => {
  // Mock Service Details
  await page.route('**/api/v1/services/test-service', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        service: {
            id: 'test-service',
            name: 'Test Service',
            httpService: { address: 'https://example.com' },
            disable: false,
        }
      }),
    });
  });

  // Mock Validation Response
  await page.route('**/api/v1/services/validate', async route => {
      await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
              valid: false,
              latency_ms: 123,
              steps: [
                  { name: 'Configuration Syntax', status: 'success' },
                  { name: 'Connectivity Check', status: 'error', message: 'dial tcp: lookup example.com: no such host' }
              ],
              error: 'dial tcp: lookup example.com: no such host',
              details: 'Connectivity check failed'
          })
      });
  });

  await page.goto('/upstream-services/test-service');

  // Click Run Diagnostics
  await page.click('button:has-text("Run Diagnostics")');

  // Wait for result
  await page.waitForSelector('text=Connection Failed');

  // Screenshot
  await page.screenshot({ path: 'ui/docs/screenshots/connection-diagnostics.png' });
});
